package tail

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/tail/follower"
)

const (
	filterName = "lines.cuckoo"
)

var (
	_log = log.GetLogger()
)

func isStaleDir(path string) bool {
	autitLogPath := fmt.Sprintf("%s/audit.log", path)
	info, err := os.Stat(autitLogPath)
	if err != nil {
		_log.Infow("unable to stat file", "path", autitLogPath, "error", err)
		return false
	}

	// delete directory that has no modified files in last 7 days.
	if info.ModTime().Before(time.Now().Add(-time.Hour * 24 * 7)) {
		_log.Infow("audit log not updated for 7 days", "path", autitLogPath)
		return true
	}

	return false
}

func hasLogs(path string) bool {

	matches, _ := filepath.Glob(fmt.Sprintf("%s/audit*.log", path))

	return len(matches) > 0
}

func findStaleDirs(root string) ([]string, error) {
	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}

	var staleDirs []string
	var nonStaleDirs []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			_log.Infow("prevent panic by handling failure in accessing the path", path, "error", err)
			return err
		}
		if info.IsDir() && hasLogs(path) {
			if isStaleDir(path) {
				staleDirs = append(staleDirs, path)
			} else {
				nonStaleDirs = append(nonStaleDirs, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(nonStaleDirs) <= 0 {
		// looks like all are stale directories, there are no logs for
		// 7 days hence can't identify which one to delete.
		// skip stale directory cleanup for now.
		return nil, fmt.Errorf("can not identify stale vs non-stale directory")
	}

	return staleDirs, nil
}

func findLogDirs(root string) ([]string, error) {

	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}

	var auditDirs []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			_log.Infow("prevent panic by handling failure in accessing the path", path, "error", err)
			return err
		}
		if info.IsDir() && hasLogs(path) && !isStaleDir(path) {
			auditDirs = append(auditDirs, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return auditDirs, nil
}

func tailRotatedFile(ctx context.Context, filePath string, tailChan chan<- LogMsg) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	limiter := rate.NewLimiter(rate.Every(time.Minute/200), 200)

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {

		err := limiter.Wait(ctx)
		if err != nil {
			return err
		}

		var lm LogMsg
		err = json.Unmarshal(scanner.Bytes(), &lm)
		if err != nil {
			continue
		}

		tailChan <- lm
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// remove the file after done reading
	os.Remove(filePath)

	return nil
}

func tailCurrentFile(ctx context.Context, filePath string, tailChan chan<- LogMsg) error {

	t, err := follower.New(filePath, follower.Config{
		Whence: io.SeekStart,
		Offset: 0,
		Reopen: true,
	})

	if err != nil {
		return err
	}

	limiter := rate.NewLimiter(rate.Every(time.Minute/200), 200)

	for line := range t.Lines() {

		err := limiter.Wait(ctx)
		if err != nil {
			return err
		}

		var lm LogMsg
		err = json.Unmarshal(line.Bytes(), &lm)
		if err != nil {
			continue
		}

		tailChan <- lm

	}

	if t.Err() != nil {
		return err
	}

	return nil
}

// tailDir tails audit log directory for a relay pod
func tailDir(ctx context.Context, path string, logChan chan<- LogMsg) error {

	// filter, err := getFilter(path)
	// if err != nil {
	// 	return err
	// }

	// defer saveFilter(path, filter)

	files := func() []string {
		matches, _ := filepath.Glob(fmt.Sprintf("%s/audit*.log", path))

		sort.Strings(matches)
		return matches
	}()

	//_log.Infow("found files", "files", files)

	tailChan := make(chan LogMsg)
	for _, file := range files {
		if strings.HasSuffix(file, "audit.log") {
			go tailCurrentFile(ctx, file, tailChan)
		} else {
			go tailRotatedFile(ctx, file, tailChan)
		}
	}

tailLoop:
	for {
		select {
		case <-ctx.Done():
			break tailLoop
		case log := <-tailChan:

			// // if the log xid is seen already ignore
			// if filter.Lookup(lm.XID.Bytes()) {
			// 	filter.Delete(lm.XID.Bytes())
			// 	continue
			// }

			// filter.InsertUnique(lm.XID.Bytes())
			logChan <- log

		}
	}

	return nil
}
