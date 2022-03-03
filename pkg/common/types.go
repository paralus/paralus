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
