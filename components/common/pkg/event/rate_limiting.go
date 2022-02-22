package event

import (
	"encoding/json"
	"time"

	"k8s.io/client-go/util/workqueue"
)

type rateLimitingQueue struct {
	q   workqueue.RateLimitingInterface
	in  <-chan Resource
	out chan<- Resource
}

func resourceToKey(r Resource) string {
	b, _ := json.Marshal(r)
	return string(b)
}

func keyToResource(k string) Resource {
	var r Resource
	json.Unmarshal([]byte(k), &r)
	return r
}

// NewRateLimitngQueue returns new rate limiting resource event queue
func NewRateLimitngQueue(numWorkers int, stop <-chan struct{}) (inChan chan<- Resource, outChan <-chan Resource) {
	in := make(chan Resource, numWorkers)
	out := make(chan Resource, numWorkers)

	q := workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
		time.Millisecond*10,
		time.Millisecond*50,
	))

	//q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	rq := &rateLimitingQueue{
		q:   q,
		in:  in,
		out: out,
	}

	go rq.run(stop)
	return in, out
}

func (rq *rateLimitingQueue) run(stop <-chan struct{}) {

	go func() {
	enqueueLoop:
		for {
			select {
			case <-stop:
				break enqueueLoop
			case r := <-rq.in:
				key := resourceToKey(r)
				//rq.q.Add(key)
				rq.q.Add(key)
				_log.Debugw("enqueued", "key", key, "len", rq.q.Len())

			}
		}

	}()

	go func() {
		_log.Debugw("running dequeue")
		for {
			_log.Debugw("queue len", "len", rq.q.Len())
			key, shutdown := rq.q.Get()
			if shutdown {
				break
			}
			_log.Debugw("got item", "item", key)

			r := keyToResource(key.(string))

			rq.q.Forget(key)
			rq.q.Done(key)
			_log.Debugw("dequeued", "resource", r)

			select {
			case rq.out <- r:
			default:
				_log.Debugw("unable to dequeue adding back to queue", "resource", r)
				rq.q.Add(key)
			}
		}
	}()

	go func() {
		<-stop
		rq.q.ShutDown()

	}()

	return
}
