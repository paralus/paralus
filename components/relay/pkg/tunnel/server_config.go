package tunnel

// Dialin defines a dialin.
type Dialin struct {
	Protocol   string
	Addr       string
	ServerName string
	RootCA     []byte
	ServerCRT  []byte
	ServerKEY  []byte
	Version    string
}

// Relay defines a relay.
type Relay struct {
	Protocol   string
	Addr       string
	DialinSfx  string
	ServerName string
	RootCA     []byte
	ServerCRT  []byte
	ServerKEY  []byte
	Version    string
}

// ControllerInfo defines controller info.
type ControllerInfo struct {
	Addr         string
	PeerProbeSNI string
	RootCA       string
	ClientCRT    string
	ClientKEY    string
}

// ServerConfig is the configuration for relay server
type ServerConfig struct {
	RelayAddr  string
	Relays     map[string]*Relay
	CDRelays   map[string]*Relay
	Dialins    map[string]*Dialin
	Controller ControllerInfo
	AuditPath  string
}
