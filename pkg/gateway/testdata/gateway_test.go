package testdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/paralus/paralus/pkg/gateway"
	"github.com/paralus/paralus/pkg/grpc"
)

type testServer struct {
}

func (s *testServer) Get(ctx context.Context, o *TestObject) (*TestObject, error) {
	fmt.Println(o)
	return o, nil
}

func runServer(stop <-chan struct{}) {

	go func() {
		lr, err := net.Listen("tcp", ":9998")
		if err != nil {
			panic(err)
		}

		s, err := grpc.NewServer()
		if err != nil {
			panic(err)
		}
		RegisterTestServer(s, &testServer{})

		if err := s.Serve(lr); err != nil {
			panic(err)
		}
	}()

	go func() {
		mux := http.NewServeMux()

		gwHandler, err := gateway.NewGateway(context.TODO(), ":9998", make([]runtime.ServeMuxOption, 0), RegisterTestHandlerFromEndpoint)

		mux.Handle("/", gwHandler)

		hs := http.Server{
			Addr:    ":9999",
			Handler: mux,
		}

		if err = hs.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	<-stop

}

func TestGateway(t *testing.T) {
	stop := make(chan struct{})

	go runServer(stop)

	defer func() {
		close(stop)
	}()

	client := http.Client{}
	resp, err := client.Get("http://localhost:9999/v2/test/project/rx8099/test/123")
	if err != nil {
		t.Error(err)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}

	var obj TestObject

	err = json.Unmarshal(b, &obj)
	if err != nil {
		t.Error(err)
		return
	}

	if obj.UrlScope != "project/rx8099" {
		t.Error("expected project/rx8099, got", obj.UrlScope)
		return
	}

}
