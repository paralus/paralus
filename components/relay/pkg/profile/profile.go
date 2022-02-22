package profile

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

// RunProfiler runs pprof on the given port
func RunProfiler(port int, stop <-chan struct{}) error {

	s := http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	defer func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		s.Shutdown(ctx)
	}()

	err := s.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
