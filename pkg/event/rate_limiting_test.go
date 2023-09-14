package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimitingResourceQueue(t *testing.T) {
	stop := make(chan struct{})

	numWorkers := 5
	in, out := NewRateLimitngQueue(numWorkers, stop)

	var gen1, gen2, gen3 int
	var con1, con2, con3 int32

	var wg sync.WaitGroup
	wg.Add(numWorkers + 3)

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
				time.Sleep(time.Millisecond * 1)
			}
		}
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
				time.Sleep(time.Millisecond * 1)
			}
		}
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
				time.Sleep(time.Millisecond * 1)
			}
		}
	}(&wg)

	for i := 0; i < numWorkers; i++ {
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
					time.Sleep(time.Millisecond * 1000)
				}
			}
		}(i, &wg)
	}

	time.Sleep(time.Second * 10)
	close(stop)
	wg.Wait()

	t.Log("gen1", gen1, "con1", con1, "gen2", gen2, "con2", con2, "gen3", gen3, "con3", con3)
}
