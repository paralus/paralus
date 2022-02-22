package tunnel

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"golang.org/x/net/http2"
)

var _dplog *relaylogger.RelayLog

type connPair struct {
	conn       net.Conn
	clientConn *http2.ClientConn
}

type dialinConnector struct {
	// this is list of dialin peer ids
	connKeys []string
	lbIndex  int
}
type dialinPool struct {
	t     *http2.Transport
	conns map[string]connPair // key is host:port
	free  func(string)
	mu    sync.RWMutex
	//map of connector IDs with dialin SNI key
	dialinConnectors map[string]*dialinConnector
}

func newDialinPool(t *http2.Transport, f func(string), log *relaylogger.RelayLog) *dialinPool {
	return &dialinPool{
		t:                t,
		free:             f,
		conns:            make(map[string]connPair),
		dialinConnectors: make(map[string]*dialinConnector),
	}
}

func (p *dialinPool) URL(key string) string {
	return fmt.Sprint("https://", key)
}

//GetClientConn get connector
func (p *dialinPool) GetClientConn(req *http.Request, addr string) (*http2.ClientConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if cp, ok := p.conns[addr]; ok && cp.clientConn.CanTakeNewRequest() {
		return cp.clientConn, nil
	}

	return nil, errClientNotConnected
}

// CheckDialinKeyExist check cached key's still exist in pool
func (p *dialinPool) CheckDialinKeyExist(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.conns[key]
	return ok
}

func (p *dialinPool) MarkDead(c *http2.ClientConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for addr, cp := range p.conns {
		if cp.clientConn == c {
			_dplog.Debug("MarkDead", addr, p)
			p.close(cp, addr)
			s := strings.SplitAfter(addr, utils.JoinString)
			if len(s) == 3 {
				sni := strings.Trim(s[0], utils.JoinString)
				addr := s[1] + s[2]
				if !p.deleteDialinConnectorKey(sni, addr) {
					_dplog.Error(nil, "error in dialin MarkDead connector key delete did not find key ", addr)
				}
			}
			return
		}
	}
}

func (p *dialinPool) AddConn(conn net.Conn, identifier string, sni string, remoteAddr string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// append peer-cert-id with remoteAddr
	addr := p.addr(identifier) + utils.JoinString + remoteAddr

	// prepend dialin sni to addr to construct the key
	key := sni + utils.JoinString + addr

	if cp, ok := p.conns[key]; ok {
		// check conns has the key
		if err := p.ping(cp); err != nil {
			p.close(cp, key)
		} else {
			return "", fmt.Errorf("connector key %s already in dialin pool ", key)
		}
	}

	// create transport new connection
	c, err := p.t.NewClientConn(conn)
	if err != nil {
		return "", err
	}

	// add conn to conns
	p.conns[key] = connPair{
		conn:       conn,
		clientConn: c,
	}

	go sendHeartBeats(p.conns[key])

	// set dialinConnectors map using key as sni
	if item, ok := p.dialinConnectors[sni]; ok {
		itemCount := len(item.connKeys)
		if itemCount <= 0 {
			p.dialinConnectors[sni].connKeys = append(p.dialinConnectors[sni].connKeys, addr)
			sort.Strings(p.dialinConnectors[sni].connKeys)
		} else {
			index := sort.SearchStrings(p.dialinConnectors[sni].connKeys, addr)
			if index < itemCount && p.dialinConnectors[sni].connKeys[index] == addr {
				_dplog.Info("address already exist")
			} else {
				//item not found
				p.dialinConnectors[sni].connKeys = append(p.dialinConnectors[sni].connKeys, addr)
				sort.Strings(p.dialinConnectors[sni].connKeys)
			}
		}
	} else {
		p.dialinConnectors[sni] = &dialinConnector{
			lbIndex: 0,
		}
		p.dialinConnectors[sni].connKeys = append(p.dialinConnectors[sni].connKeys, addr)
		sort.Strings(p.dialinConnectors[sni].connKeys)
	}

	_dplog.Info(
		"Added dialin connection",
		"addr", addr,
		"key", key,
	)
	return key, nil
}

//GetDialinConnectorKey get connector key
func (p *dialinPool) GetDialinConnectorCount(sni string) (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_dplog.Debug(
		"GetDialinConnectorCount",
		"sni", sni,
		"pool", p,
	)

	_dplog.Info("GetDialinConnectorCount", p.dialinConnectors)
	if item, ok := p.dialinConnectors[sni]; ok {
		itemCount := len(item.connKeys)
		if itemCount > 0 {
			return itemCount, nil
		}
	}

	return 0, fmt.Errorf("Empty dialin pool.dialinConnectors for sni %s ", sni)
}

func (p *dialinPool) getConnKey(sni string, item *dialinConnector, count int) (string, error) {
	for i := 0; i < count; i++ {
		key := sni + utils.JoinString + item.connKeys[i]
		if _, ok := p.conns[key]; ok {
			return key, nil
		}
	}

	return "", fmt.Errorf("Empty dialin pool.dialinConnectors for sni %s ", sni)
}

//GetDialinConnectorKey get connector key
func (p *dialinPool) GetDialinConnectorKey(sni string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	_dplog.Info(
		"GetDialinConnectorKey",
		"sni", sni,
	)

	if item, ok := p.dialinConnectors[sni]; ok {
		itemCount := len(item.connKeys)

		if itemCount <= 0 {
			return "", fmt.Errorf("Empty dialin pool.dialinConnectors for sni %s ", sni)
		}

		//simple round robin selection
		if item.lbIndex >= itemCount {
			item.lbIndex = 0
		}

		key := sni + utils.JoinString + item.connKeys[item.lbIndex]
		item.lbIndex++

		// check key
		if _, ok := p.conns[key]; !ok {
			// get any key
			return p.getConnKey(sni, item, itemCount)
		}

		// key is valid
		return key, nil
	}

	return "", fmt.Errorf("Empty dialin pool.dialinConnectors for sni %s ", sni)
}

func (p *dialinPool) deleteDialinConnectorKey(sni, addr string) bool {
	if item, ok := p.dialinConnectors[sni]; ok {
		itemCount := len(item.connKeys)
		index := sort.SearchStrings(p.dialinConnectors[sni].connKeys, addr)
		if index < itemCount {
			// Not expecting lots of connectors. So using slower delete opeartion
			// by shifting elements to preserve the sorted orders
			p.dialinConnectors[sni].connKeys = append(item.connKeys[:index], item.connKeys[index+1:]...)
			_dplog.Info(
				"Deleted Connnection from dialin pool",
				"addr", addr,
			)
			if len(p.dialinConnectors[sni].connKeys) <= 0 {
				_dplog.Info(
					"Deleted last connection from dialin pool",
					"sni", sni,
					"addr", addr,
				)
				delete(p.dialinConnectors, sni)
			}
			return true
		}
	}
	return false
}

func (p *dialinPool) DeleteConn(identifier string, sni string, remoteAddr string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// append peer-cert-id with remoteAddr
	addr := identifier + utils.JoinString + remoteAddr

	// prepend dialin sni to addr to construct the key
	key := sni + utils.JoinString + addr
	if cp, ok := p.conns[key]; ok {
		p.close(cp, key)
	} else {
		_dplog.Error(
			nil,
			"did not find key in pool.Conns",
			"key", key,
			"pool", p,
		)
	}

	if p.deleteDialinConnectorKey(sni, addr) {
		return
	}

	_dplog.Info(
		"DeleteConn did not find the coonection",
		"sni", sni,
		"addr", addr,
		"key", key,
		"pool", p,
	)

}

func (p *dialinPool) Ping(identifier string) (time.Duration, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if cp, ok := p.conns[identifier]; ok {
		start := time.Now()
		err := p.ping(cp)
		return time.Since(start), err
	}

	return 0, errClientNotConnected
}

// heart beats to keep dialins alive to avoid idletimeout
func sendHeartBeats(cp connPair) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), utils.DefaultPingTimeout)
		if err := cp.clientConn.Ping(ctx); err != nil {
			cancel()
			_dplog.Debug("dialin keep-alive ping err", err)
			return
		}
		_dplog.Debug("dialin keep-alive ping success")
		cancel()
		time.Sleep(30 * time.Second)
	}
}

func (p *dialinPool) ping(cp connPair) error {
	ctx, cancel := context.WithTimeout(context.Background(), utils.DefaultPingTimeout)
	defer cancel()

	return cp.clientConn.Ping(ctx)
}

func (p *dialinPool) close(cp connPair, addr string) {
	cp.conn.Close()
	_dplog.Info("dialin close", addr)
	delete(p.conns, addr)
	if p.free != nil {
		p.free(addr)
	}
}

func (p *dialinPool) addr(identifier string) string {
	return identifier
}

//StartDialinPoolMgr starting dialin connection manager
func StartDialinPoolMgr(ctx context.Context, log *relaylogger.RelayLog, exitChan chan<- bool) {
	_dplog = log.WithName("DialinPool")

	for {
		select {
		case <-ctx.Done():
			slog.Error(
				ctx.Err(),
				"Stopping dialin pool manager",
			)
			return
		}
	}
}

//GetDialinMetrics get connector key
func (p *dialinPool) GetDialinMetrics(w http.ResponseWriter) {
	var clusterCnt, connCnt int
	clusterCnt, connCnt = 0, 0

	p.mu.RLock()
	defer p.mu.RUnlock()

	_dplog.Info(
		"GetDialinMetrics",
	)

	fmt.Fprintf(w, "{\"dialinmetrics\": [")
	for sni, item := range p.dialinConnectors {
		clusterCnt++
		connCnt += len(item.connKeys)
		itemCount := strconv.Itoa(len(item.connKeys))

		fmt.Fprintf(w, "{\"cluster\": \""+sni+"\", \"connections\": \""+itemCount+"\"},")

	}

	fmt.Fprintf(w, "], \"totalclusters\": "+strconv.Itoa(clusterCnt))
	fmt.Fprintf(w, ", \"totalconnections\": "+strconv.Itoa(connCnt))
	fmt.Fprintf(w, ", \"podname\": "+utils.PODNAME)

	fmt.Fprintf(w, " }")
}
