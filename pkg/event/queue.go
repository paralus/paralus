package event

import (
	"sync"

	"github.com/paralus/paralus/pkg/log"
)

const (
	// internalBufferSize is size of the buffer that couple in and out chan
	defaultBufferSize = 100
	// defaultIOBufferSize is a proxy for max parallelism while reading from
	// out chan

	defaultIOBufferSize = 10
)

var (
	_log = log.GetLogger()
)

// uniqueQueue is the containing type for set-style / unique queues
type uniqueQueue struct {
	in       <-chan Resource
	out      chan<- Resource
	inBuffer map[Resource]struct{}
	exists   sync.Map
	buffer   chan Resource
}

// NewUniqueQueue returns a queue for events which ensures that events in the queue are unique
// it takes optional arguments ioSize and bufSize
// ioSize is the size of in and out buffered channels
// bufSize is the size of internal buffer in the queue
func NewUniqueQueue(stop <-chan struct{}, size ...int) (inChan chan<- Resource, outChan <-chan Resource) {

	if len(size) > 2 {
		panic("more than 2 args supplied for size")
	}

	ioSize := defaultIOBufferSize
	bufSize := defaultBufferSize

	if len(size) >= 1 {
		if size[0] > 0 {
			ioSize = size[0]
		}
	}

	if len(size) > 1 {
		if size[1] > 0 {
			bufSize = size[1]
		}
	}

	in := make(chan Resource, ioSize)
	out := make(chan Resource, ioSize)
	buffer := make(chan Resource, bufSize)

	queue := &uniqueQueue{in: in, out: out, buffer: buffer}

	go queue.run(stop)

	return in, out
}

func (q *uniqueQueue) run(stop <-chan struct{}) {

	var _wg sync.WaitGroup
	_wg.Add(2)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
	pushLoop:
		for {
			select {
			case <-stop:
				break pushLoop
			case item := <-q.in:

				if _, ok := q.exists.Load(item); ok {
					continue
				}

				q.buffer <- item
				q.exists.Store(item, struct{}{})

			}
		}
	}(&_wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
	popLoop:
		for {
			select {
			case <-stop:
				break popLoop
			// messages can be dropped when the system is unable to
			// dequeue as fast as
			case item := <-q.buffer:
				select {
				case q.out <- item:
				default:
					_log.Infow("unable to dequeu, dropping")
				}

				q.exists.Delete(item)
			}
		}
	}(&_wg)

	_wg.Wait()
	close(q.buffer)

}
