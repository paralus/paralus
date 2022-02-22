package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestUniqueScopedResourceQueue(t *testing.T) {
	stop := make(chan struct{})

	numCon := 5
	bufSize := 100
	in, out := NewUniqueQueue(stop, numCon, bufSize)

	var gen1, gen2, gen3 int
	var con1, con2, con3 int32

	var wg sync.WaitGroup
	wg.Add(numCon + 3)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()
	genLoop:
		for {
			select {
			case <-stop:
				break genLoop
			default:
				in <- Resource{ID: "rx28oml"}
				gen1++
				time.Sleep(time.Millisecond * 2)
			}
		}
		//t.Log("exited gen1")
	}(&wg)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()
	genLoop:
		for {
			select {
			case <-stop:
				break genLoop
			default:
				in <- Resource{ID: "4qkolkn"}
				gen2++
				time.Sleep(time.Millisecond * 2)
			}
		}
		//t.Log("exited gen2")
	}(&wg)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()
	genLoop:
		for {
			select {
			case <-stop:
				break genLoop
			default:
				in <- Resource{ID: "7w2lnkp"}
				gen3++
				time.Sleep(time.Millisecond * 2)
			}
		}
		//t.Log("exited gen3")
	}(&wg)

	for i := 0; i < numCon; i++ {
		go func(idx int, _wg *sync.WaitGroup) {
			defer _wg.Done()

		conLoop:
			for {
				select {
				case <-stop:
					break conLoop
				case item := <-out:
					switch item.ID {
					case "rx28oml":
						atomic.AddInt32(&con1, 1)
					case "4qkolkn":
						atomic.AddInt32(&con2, 1)
					case "7w2lnkp":
						atomic.AddInt32(&con3, 1)
					default:
						t.Log("unexpeced id", item.ID)
					}
					time.Sleep(time.Millisecond * 1)
				}
			}
			//t.Log("exited ", idx)
		}(i, &wg)
	}

	time.Sleep(time.Second * 5)
	close(stop)
	wg.Wait()

	t.Log("gen1", gen1, "con1", con1, "gen2", gen2, "con2", con2, "gen3", gen3, "con3", con3)

}
