package leaderelection

import (
	"fmt"
	"testing"
	"time"
)

func TestLeaderElectionRun(t *testing.T) {
	lock1, err := NewLock("test-lock", "default", "client-1")
	if err != nil {
		t.Error(err)
		return
	}
	lock2, err := NewLock("test-lock", "default", "client-2")
	if err != nil {
		t.Error(err)
		return
	}
	stop1 := make(chan struct{})
	stop2 := make(chan struct{})

	go func() {
		Run(lock1, func(stop <-chan struct{}) {
			fmt.Println(lock1.Identity(), " became leader")
		loop:
			for {
				select {
				case <-stop:
					fmt.Println("stopping ", lock1.Identity())
					break loop
				}
			}
		}, stop1)
	}()

	go func() {
		Run(lock2, func(stop <-chan struct{}) {
			fmt.Println(lock2.Identity(), " became leader")
		loop:
			for {
				select {
				case <-stop:
					fmt.Println("stopping ", lock2.Identity())
					break loop
				}
			}
		}, stop2)
	}()

	go func() {
		time.Sleep(time.Second * 20)
		close(stop1)
	}()

	go func() {
		time.Sleep(time.Second * 20)
		close(stop2)
	}()

	<-stop1
	<-stop2

	time.Sleep(time.Second * 1)
}

func TestLeaderElectionConfigMapRun(t *testing.T) {
	lock1, err := NewConfigMapLock("test-lock", "default", "client-1")
	if err != nil {
		t.Error(err)
		return
	}
	lock2, err := NewConfigMapLock("test-lock", "default", "client-2")
	if err != nil {
		t.Error(err)
		return
	}
	stop1 := make(chan struct{})
	stop2 := make(chan struct{})

	go func() {
		Run(lock1, func(stop <-chan struct{}) {
			fmt.Println(lock1.Identity(), " became leader")
		loop:
			for {
				select {
				case <-stop:
					fmt.Println("stopping ", lock1.Identity())
					break loop
				}
			}
		}, stop1)
	}()
	go func() {
		Run(lock2, func(stop <-chan struct{}) {
			fmt.Println(lock2.Identity(), " became leader")
		loop:
			for {
				select {
				case <-stop:
					fmt.Println("stopping ", lock2.Identity())
					break loop
				}
			}
		}, stop2)
	}()

	time.Sleep(time.Second * 20)
	close(stop2)

	time.Sleep(time.Second * 20)
	close(stop1)

}
