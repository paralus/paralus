package main

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

var min = 10
var max = 50

func worker1(ctx context.Context, index int) {
	var errCnt = 0
	var success = 0
	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case <-ctx.Done():
			time.Sleep(10 * time.Second)
			fmt.Println("worker1 ", index, " done: success", success, "errcnt ", errCnt)
			return
		default:
			//out, err := exec.Command("/usr/local/bin/kubectl", "--kubeconfig", "../kubectlconfigs/kubeconfigproxy-123abcd.user.relay.rafay.local", "get", "pods", "-o", "yaml").Output()
			//out, err := exec.Command("/usr/local/bin/kubectl", "--kubeconfig", "/Users/stephanbenny/WorkSpace/kubeconfig-c-5", "get", "all", "-A").Output()
			out, err := exec.Command("/usr/local/bin/kubectl", "--kubeconfig", "/Users/stephanbenny/WorkSpace/kubeconfig-c-5", "get", "pods").Output()
			//out, err := exec.Command("/usr/local/bin/kubectl", "--kubeconfig", "/Users/stephanbenny/WorkSpace/kubeconfig-c-5", "get", "pods", "-o", "yaml").Output()
			if err != nil {
				errCnt++
			} else {
				success++
			}
			intvl := time.Duration(rand.Intn(max-min+1)+min) * time.Millisecond
			time.Sleep(intvl)
			fmt.Printf("worker1[%d][%d][%d]: %s\n", index, success, errCnt, out)
		}
	}
}

func worker2(ctx context.Context, index int) {
	var errCnt = 0
	var success = 0
	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case <-ctx.Done():
			time.Sleep(10 * time.Second)
			fmt.Println("worker2 ", index, " done: success", success, "errcnt ", errCnt)
			return
		default:
			//out, err := exec.Command("/usr/local/bin/kubectl", "--kubeconfig", "../kubectlconfigs/kubeconfigproxy-123abcd.user.relay.rafay.local", "get", "pods", "-o", "yaml").Output()
			//out, err := exec.Command("/usr/local/bin/kubectl", "--kubeconfig", "/Users/stephanbenny/WorkSpace/kubeconfig-c-6", "get", "all", "-A").Output()
			out, err := exec.Command("/usr/local/bin/kubectl", "--kubeconfig", "/Users/stephanbenny/WorkSpace/kubeconfig-c-6", "get", "pods").Output()
			//out, err := exec.Command("/usr/local/bin/kubectl", "--kubeconfig", "/Users/stephanbenny/WorkSpace/kubeconfig-c-5", "get", "pods", "-o", "yaml").Output()
			if err != nil {
				errCnt++
			} else {
				success++
			}
			intvl := time.Duration(rand.Intn(max-min+1)+min) * time.Millisecond
			time.Sleep(intvl)
			fmt.Printf("worker2[%d][%d][%d]: %s\n", index, success, errCnt, out)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	for i := 1; i <= 10; i++ {
		//go worker2(ctx, i)
		go worker1(ctx, i)

	}

	time.Sleep(120 * time.Minute)

	cancel()

	time.Sleep(15 * time.Second)

}
