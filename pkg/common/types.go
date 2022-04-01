package common

type Relay struct {
	Token         string `json:"token"`
	Addr          string `json:"addr"`
	Endpoint      string `json:"endpoint"`
	Name          string `json:"name"`
	TemplateToken string `json:"templateToken"`
}

type DownloadData struct {
	ControlAddr     string
	APIAddr         string
	RelayAgentImage string
}

type CliConfigDownloadData struct {
	Profile      string `json:"profile"`
	RestEndpoint string `json:"rest_endpoint"`
	OpsEndpoint  string `json:"ops_endpoint"`
	ApiKey       string `json:"api_key"`
	ApiSecret    string `json:"api_secret"`
	Project      string `json:"project"`
	Organization string `json:"organization"`
	Partner      string `json:"partner"`
}

type contextKey struct{}
