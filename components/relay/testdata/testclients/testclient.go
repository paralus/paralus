package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	//UNIXSOCKET ..
	UNIXSOCKET = "/tmp/relay-unix-"
)

func startClient(wg *sync.WaitGroup, buf []byte) {
	defer wg.Done()

	fmt.Print("stat client")
	buf3 := make([]byte, 4096)
	// connect to this socket
	socketPath := UNIXSOCKET + "kubectldialin.relay.rafay.dev"
	conn, err := net.DialTimeout("unix", socketPath, 60*time.Second)
	if err != nil {
		fmt.Println("failed to connect ", err)
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
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Number of clients: ")
	numStr, _ := reader.ReadString('\n')
	fmt.Print("Key to send: ")
	key, _ := reader.ReadString('\n')

	message := fmt.Sprintf("{\"DialinKey\": \"%s\", \"UserName\": \"namespace-admin-sa\", \"SNI\":  \"cluster1.kubectldialin.relay.rafay.dev\"}", key[:len(key)-1])
	copy(buf, message)

	fmt.Print("numStr=", numStr)
	num, err := strconv.Atoi(numStr[:len(numStr)-1])
	fmt.Println("num=", num, " err=", err)
	for i := 0; i < num; i++ {
		wg.Add(1)
		go startClient(&wg, buf)
	}

	fmt.Print("Waiting for workers")
	wg.Wait()

	fmt.Print("Completed")
}
