package notify

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/paralus/paralus/pkg/log"
	"github.com/paralus/paralus/pkg/match"
	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/service"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

var (
	_log      = log.GetLogger()
	_notifier Notifier
	// ErrNotInitialized is returned when notifier is not initialized
	ErrNotInitialized = errors.New("notifier not initialized")
	once              = sync.Once{}
)

const (
	maxNotifyWorkers = 6
)

// Notifier is the interface for notifying cluster changes
type Notifier interface {
	Start(stop <-chan struct{})
	AddListener(c chan<- infrav3.Cluster, opts ...query.Option) error
	RemoveListener(c chan<- infrav3.Cluster)
}

// New returns new notifier
func New(cs service.ClusterService) Notifier {
	return &notifier{
		ClusterService: cs,
		listeners:      make(map[chan<- infrav3.Cluster]match.Matcher),
	}
}

func KeyFromMeta(meta *commonv3.Metadata) string {
	return fmt.Sprintf("%s/%s/%s/%s", meta.Partner, meta.Organization, meta.Project, meta.Name)
}

func MetaFromKey(key string) (meta commonv3.Metadata) {
	items := strings.Split(key, "/")
	if len(items) != 4 {
		return
	}

	meta.Name = items[3]
	meta.Partner = items[0]
	meta.Organization = items[1]
	meta.Project = items[2]

	return
}

type notifier struct {
	sync.RWMutex
	service.ClusterService
	listeners map[chan<- infrav3.Cluster]match.Matcher
}

var _ Notifier = (*notifier)(nil)

func (n *notifier) Start(stop <-chan struct{}) {

	mChan := make(chan commonv3.Metadata, maxNotifyWorkers)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-stop
		defer cancel()
		//defer wq.ShutDown()
	}()

	// start cluster service listener
	go n.ListenClusters(ctx, mChan)

	for i := 0; i < maxNotifyWorkers; i++ {
		go func() {
		notifyLoop:
			for {
				select {
				case <-stop:
					break notifyLoop
				case m := <-mChan:

					notify := func(meta commonv3.Metadata) {
						nctx, cancel := context.WithTimeout(ctx, time.Second*1)
						defer cancel()
						c, err := n.Get(nctx, query.WithMeta(&m))
						if err != nil {
							_log.Infow("invalid cluster meta for notify", "meta", m)
							return
						}
						n.notifyListeners(*c)
					}

					notify(m)
				}
			}
		}()
	}

	<-stop

}

func (n *notifier) AddListener(c chan<- infrav3.Cluster, opts ...query.Option) error {

	matcher, err := match.New(opts...)
	if err != nil {
		return err
	}

	n.Lock()
	defer n.Unlock()

	n.listeners[c] = matcher
	_log.Debugw("notify listerners", "number", len(n.listeners))

	return nil
}

func (n *notifier) RemoveListener(c chan<- infrav3.Cluster) {
	n.Lock()
	delete(n.listeners, c)
	n.Unlock()
	_log.Debugw("notify listerners", "number", len(n.listeners))
}

func (n *notifier) notifyListeners(c infrav3.Cluster) {
	n.RLock()
	for lChan, matcher := range n.listeners {
		if matcher.Match(*c.Metadata) {
			// only send if channel is ready to accept
			select {
			case lChan <- c:
			default:
			}

			// if len(lChan) < 1 {
			// 	lChan <- c
			// }
		}
	}
	n.RUnlock()
}

// Init initializes the notifier at package level
func Init(cs service.ClusterService) {
	_notifier = New(cs)
}

// Start starts the notifier at package lvel
func Start(stop <-chan struct{}) error {

	if _notifier == nil {
		return ErrNotInitialized
	}

	once.Do(func() {
		go _notifier.Start(stop)
	})

	return nil
}

// AddListener adds listerner to the notifier
func AddListener(c chan<- infrav3.Cluster, opts ...query.Option) error {
	if _notifier == nil {
		return ErrNotInitialized
	}

	return _notifier.AddListener(c, opts...)
}

// RemoveListener removes listener from notifier
func RemoveListener(c chan<- infrav3.Cluster) error {
	if _notifier == nil {
		return ErrNotInitialized
	}

	_notifier.RemoveListener(c)
	return nil
}
