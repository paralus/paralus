package tail

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/jinzhu/inflection"
	"github.com/twmb/murmur3"

	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	"github.com/julienschmidt/httprouter"
)

var _dummyHandler = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {}

var _bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var _hashPool = sync.Pool{
	New: func() interface{} {
		return murmur3.New64()
	},
}

func getCacheKey(s string) uint64 {
	hasher := _hashPool.Get().(hash.Hash64)
	defer _hashPool.Put(hasher)

	buf := _bufPool.Get().(*bytes.Buffer)
	defer _bufPool.Put(buf)

	buf.Reset()
	buf.WriteString(s)

	hasher.Reset()
	hasher.Write(buf.Bytes())
	return hasher.Sum64()
}

// Transformer is the interface for transforming relay log message into audit message
type Transformer interface {
	Transform(lm *LogMsg, am *AuditMsg) error
}

// NewTransformer returns new relay log transformer
func NewTransformer(authzPool sentryrpc.SentryAuthorizationPool) (Transformer, error) {

	r := httprouter.New()

	r.Handle("GET", "/api", _dummyHandler)
	r.Handle("GET", "/api/:version", _dummyHandler)
	r.Handle("GET", "/api/:version/:kind0", _dummyHandler)
	r.Handle("GET", "/api/:version/:kind0/:namespace", _dummyHandler)
	r.Handle("GET", "/api/:version/:kind0/:namespace/:kind1", _dummyHandler)
	r.Handle("GET", "/api/:version/:kind0/:namespace/:kind1/:name", _dummyHandler)
	r.Handle("GET", "/apis", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version/:kind0", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version/:kind0/:namespace", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version/:kind0/:namespace/:kind1", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version/:kind0/:namespace/:kind1/:name", _dummyHandler)

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // Num keys to track frequency of (10k).
		MaxCost:     1 << 27, // Maximum cost of cache (128MB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}

	t := &transformer{
		authzPool: authzPool,
		r:         r,
		cache:     cache,
	}
	return t, nil
}

type transformer struct {
	authzPool sentryrpc.SentryAuthorizationPool
	cache     *ristretto.Cache
	r         *httprouter.Router
}

func (t *transformer) getUser(cn string) (*sentryrpc.LookupUserResponse, error) {
	key := getCacheKey(cn)
	if val, ok := t.cache.Get(key); ok {
		if usr, ok := val.(*sentryrpc.LookupUserResponse); ok {
			return usr, nil
		}
	}

	// cache miss
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	authzClient, err := t.authzPool.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer authzClient.Close()

	resp, err := authzClient.LookupUser(ctx, &sentryrpc.LookupUserRequest{
		UserCN: cn,
	})

	if err != nil {
		return nil, err
	}

	t.cache.Set(key, resp, 100)
	return resp, nil
}

func (t *transformer) getCluster(serverName string) (*sentryrpc.LookupClusterResponse, error) {
	idx := strings.Index(serverName, ".")

	if idx > 0 {
		key := getCacheKey(serverName[0:idx])
		if val, ok := t.cache.Get(key); ok {
			if cluster, ok := val.(*sentryrpc.LookupClusterResponse); ok {
				return cluster, nil
			}
		}

		// cache miss
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		authzClient, err := t.authzPool.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		defer authzClient.Close()

		resp, err := authzClient.LookupCluster(ctx, &sentryrpc.LookupClusterRequest{
			ClusterSNI: serverName,
		})

		if err != nil {
			return nil, err
		}

		t.cache.Set(key, resp, 110)
		return resp, nil
	}

	return nil, fmt.Errorf("invalid serverName %s", serverName)
}

func (t *transformer) getParams(url string) httprouter.Params {

	if h, params, _ := t.r.Lookup("GET", url); h != nil {

		return params
	}

	return nil
}

func (t *transformer) Transform(lm *LogMsg, am *AuditMsg) error {

	//_log.Infow("transforming", "logMessage", *lm)
	cluster, err := t.getCluster(lm.ServerName)
	if err != nil {
		_log.Infow("unable to lookup cluster", "error", err)
		return err
	}

	user, err := t.getUser(lm.User)
	if err != nil {
		_log.Infow("unable to lookup user", "error", err)
		return err
	}

	params := t.getParams(lm.URL)

	am.ClusterName = cluster.Name
	am.OrganizationID = user.OrganizationID
	am.PartnerID = user.PartnerID
	am.UserName = user.UserName
	am.Duration = lm.Duration
	am.Written = lm.Written
	am.Timestamp = lm.Timestamp
	am.Method = lm.Method
	am.URL = lm.URL
	am.Query = lm.Query
	am.XID = lm.XID
	am.StatusCode = lm.StatusCode
	am.RemoteAddr = lm.RemoteAddr
	am.SessionType = user.SessionType

	if params != nil {
		am.APIVersion = fmt.Sprintf("%s/%s", params.ByName("group"), params.ByName("version"))
		if params.ByName("kind1") != "" {
			am.Kind = params.ByName("kind1")
		} else {
			am.Kind = params.ByName("kind0")
		}
		am.Kind = inflection.Singular(am.Kind)
		am.Namespace = params.ByName("namespace")
		am.Name = params.ByName("name")
	}

	return nil
}
