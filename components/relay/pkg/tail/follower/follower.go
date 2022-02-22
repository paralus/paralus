package follower

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	bufSize  = 4 * 1024
	peekSize = 1024
)

var (
	_ = fmt.Print
)

type Line struct {
	bytes     []byte
	discarded int
}

func (l *Line) Bytes() []byte {
	return l.bytes
}

func (l *Line) String() string {
	return string(l.bytes)
}

func (l *Line) Discarded() int {
	return l.discarded
}

type Config struct {
	Offset int64
	Whence int
	Reopen bool
}

type Follower struct {
	once     sync.Once
	file     *os.File
	filename string
	lines    chan Line
	err      error
	config   Config
	reader   *bufio.Reader
	watcher  *fsnotify.Watcher
	offset   int64
	closeCh  chan struct{}
}

func New(filename string, config Config) (*Follower, error) {
	t := &Follower{
		filename: filename,
		lines:    make(chan Line),
		config:   config,
		closeCh:  make(chan struct{}),
	}

	err := t.reopen()
	if err != nil {
		return nil, err
	}

	go t.once.Do(t.run)

	return t, nil
}

func (t *Follower) Lines() chan Line {
	return t.lines
}

func (t *Follower) Err() error {
	return t.err
}

func (t *Follower) Close() {
	t.closeCh <- struct{}{}
}

func (t *Follower) run() {
	t.close(t.follow())
}

func (t *Follower) follow() error {
	_, err := t.file.Seek(t.config.Offset, t.config.Whence)
	if err != nil {
		return err
	}

	var (
		eventChan = make(chan fsnotify.Event)
		errChan   = make(chan error, 1)
	)

	t.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	defer t.watcher.Close()
	go t.watchFileEvents(eventChan, errChan)

	t.watcher.Add(t.filename)

	for {
		for {
			// discard leading NUL bytes
			var discarded int

			for {
				b, _ := t.reader.Peek(peekSize)
				i := bytes.LastIndexByte(b, '\x00')

				if i > 0 {
					n, _ := t.reader.Discard(i + 1)
					discarded += n
				}

				if i+1 < peekSize {
					break
				}
			}

			s, err := t.reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				return err
			}

			// if we encounter EOF before a line delimiter,
			// ReadBytes() will return the remaining bytes,
			// so push them back onto the buffer, rewind
			// our seek position, and wait for further file changes.
			// we also have to save our dangling byte count in the event
			// that we want to re-open the file and seek to the end
			if err == io.EOF {
				l := len(s)

				t.offset, err = t.file.Seek(-int64(l), io.SeekCurrent)
				if err != nil {
					return err
				}

				t.reader.Reset(t.file)
				break
			}

			t.sendLine(s, discarded)
		}

		// we're now at EOF, so wait for changes
		select {
		case evt := <-eventChan:
			switch evt.Op {

			// as soon as something is written, go back and read until EOF.
			case fsnotify.Chmod:
				fallthrough

			case fsnotify.Write:
				fi, err := t.file.Stat()
				if err != nil {
					if !os.IsNotExist(err) {
						return err
					}

					// it's possible that an unlink can cause fsnotify.Chmod,
					// so attempt to rewatch if the file is missing
					if err := t.rewatch(); err != nil {
						return err
					}

					continue
				}

				// file was truncated, seek to the beginning
				if t.offset > fi.Size() {
					t.offset, err = t.file.Seek(0, io.SeekStart)
					if err != nil {
						return err
					}

					t.reader.Reset(t.file)
				}

				continue

			// if a file is removed or renamed
			// and re-opening is desired, see if it appears
			// again within a 1 second deadline. this should be enough time
			// to see the file again for log rotation programs with this behavior
			default:
				if !t.config.Reopen {
					return nil
				}

				if err := t.rewatch(); err != nil {
					return err
				}

				continue
			}

		// any errors that come from fsnotify
		case err := <-errChan:
			return err

		// a request to stop
		case <-t.closeCh:
			t.watcher.Remove(t.filename)
			return nil

		// fall back to 10 second polling if we haven't received any fsevents
		// stat the file, if it's still there, just continue and try to read bytes
		// if not, go through our re-opening routine
		case <-time.After(10 * time.Second):
			fi1, err := t.file.Stat()
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			fi2, err := os.Stat(t.filename)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if os.SameFile(fi1, fi2) {
				continue
			}

			if err := t.rewatch(); err != nil {
				return err
			}

			continue
		}
	}

}

func (t *Follower) rewatch() error {
	t.watcher.Remove(t.filename)
	if err := t.reopen(); err != nil {
		return err
	}

	t.watcher.Add(t.filename)
	return nil
}

func (t *Follower) reopen() error {
	if t.file != nil {
		t.file.Close()
		t.file = nil
	}

	file, err := os.Open(t.filename)
	if err != nil {
		return err
	}

	t.file = file
	t.reader = bufio.NewReaderSize(t.file, bufSize)

	return nil
}

func (t *Follower) close(err error) {
	t.err = err

	if t.file != nil {
		t.file.Close()
	}

	close(t.lines)
}

func (t *Follower) sendLine(l []byte, d int) {
	t.lines <- Line{l[:len(l)-1], d}
}

func (t *Follower) watchFileEvents(eventChan chan fsnotify.Event, errChan chan error) {
	for {
		select {
		case evt, ok := <-t.watcher.Events:
			if !ok {
				return
			}

			// debounce write events, but send all others
			switch evt.Op {
			case fsnotify.Write:
				select {
				case eventChan <- evt:
				default:
				}

			default:
				select {
				case eventChan <- evt:
				case err := <-t.watcher.Errors:
					errChan <- err
					return
				}
			}

		// die on a file watching error
		case err := <-t.watcher.Errors:
			errChan <- err
			return
		}
	}
}
