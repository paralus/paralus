package tunnel

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

const (
	//UNIXSOCKET ..
	UNIXSOCKET = "/tmp/relay-unix-"
)

func startClient(wg *sync.WaitGroup, buf []byte, t *testing.T) {
	defer wg.Done()

	fmt.Print("starting client")
	buf3 := make([]byte, 4096)
	// connect to this socket
	socketPath := UNIXSOCKET + "kubectldialin.relay.rafay.dev"
	conn, err := net.DialTimeout("unix", socketPath, 60*time.Second)
	if err != nil {
		t.Error("failed to connect ", err)
		return
	}

	fmt.Println("Connected to ", socketPath)

	for {
		conn.Write(buf)
		str := "GET /apis/rbac.authorization.k8s.io/v1?timeout=32s HTTP/1.1\r\nHost: 192.168.56.103:6443\r\nUser-Agent: curl/7.54.0\r\nAccept: */*\r\nX-Rafay-User:namespace-admin-sa\r\nX-Rafay-Namespace: default\r\n\r\n"
		conn.Write([]byte(str))

		conn.SetReadDeadline(time.Now().Add(10 * (time.Second)))

		for {
			nr, err := conn.Read(buf3)
			if err != nil {
				fmt.Println("done reading ", err)
				return
			}
			data := buf3[0:nr]
			println("Client got:", string(data), nr)
		}
	}
}

func TestRelayClient(t *testing.T) {
	var wg sync.WaitGroup

	buf := make([]byte, 1024)
	num := 1
	key := "cluster1.kubectldialin.relay.rafay.dev-ABCD"

	message := fmt.Sprintf("{\"DialinKey\": \"%s\", \"UserName\": \"namespace-admin-sa\", \"SNI\":  \"cluster1.kubectldialin.relay.rafay.dev\"}", key[:len(key)-1])
	copy(buf, message)

	for i := 0; i < num; i++ {
		wg.Add(1)
		fmt.Println("starting client")
		go startClient(&wg, buf, t)
	}

	fmt.Print("Waiting for workers")
	wg.Wait()

	fmt.Print("Completed")
}
