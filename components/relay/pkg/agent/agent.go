package agent

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/controller/apply"
	clientutil "github.com/RafaySystems/rcloud-base/components/common/pkg/controller/client"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/register"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/cleanup"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/proxy"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/relaylogger"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/tunnel"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	podNameEnv             = "POD_NAME"
	podNamespaceEnv        = "POD_NAMESPACE"
	maxDialsEnv            = "MAX_DIALS"
	dialoutProxyEnv        = "DIALOUT_PROXY"
	dialoutProxyAuthEnv    = "DIALOUT_PROXY_AUTHENTICATION" // user:password
	configNameEnv          = "RELAY_CONFIGMAP_NAME"
	configNameCD           = "cd-agent-configmap"
	cdConfigPath           = "/etc/config/relayConfigData"
	relayRegisterTokenName = "relays"
	relayClusterIDName     = "clusterID"
	relayAgentIDName       = "agentID"
)

var (
	podName            string
	podNamespace       string
	relayRegisterToken string
	relayClusterID     string
	relayAgentID       string
	applier            client.Client
	lastHash           string
	log                *relaylogger.RelayLog
	configName         string
)

const (
	rafaySystemNamespace = "rafay-system"
)

func processRelayConfigData(cfgData string) error {
	if err := json.Unmarshal([]byte(cfgData), &utils.RelayNetworks); err != nil {
		log.Error(
			err,
			"failed to unmashal",
		)
		return err
	}
	return nil
}

func getConfigMap(ctx context.Context, c client.Client, name string) (*corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{}

	err := c.Get(ctx, client.ObjectKey{
		Namespace: rafaySystemNamespace,
		Name:      name,
	}, &cm)
	if err != nil {
		return nil, err
	}

	cm.TypeMeta = metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	}

	return &cm, nil
}

func getConfigHash(ctx context.Context, configMap string) (string, error) {

	cm, err := getConfigMap(ctx, applier, configMap)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(cm.Data)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(b)

	//return cm.Annotations[hash.ObjectHash], nil
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func getRelayAgentConfig(ctx context.Context) error {
	var ok bool

	cm, err := getConfigMap(ctx, applier, configName)
	if err != nil {
		return err
	}

	if relayClusterID, ok = cm.Data[relayClusterIDName]; !ok {
		return fmt.Errorf("failed to get relayClusterID")
	}
	if relayRegisterToken, ok = cm.Data[relayRegisterTokenName]; !ok {
		return fmt.Errorf("failed to get relayRegisterToken")
	}

	if err := processRelayConfigData(relayRegisterToken); err != nil {
		return fmt.Errorf("failed to process relayRegisterToken %s", relayRegisterToken)
	}
	if len(utils.RelayNetworks) <= 0 {
		return fmt.Errorf("empty relaynetwork configmap %s", relayRegisterToken)
	}

	log.Info(
		"relay agent config values",
		"podName", podName,
		"podNamespace", podNamespace,
	)

	for _, item := range utils.RelayNetworks {
		log.Info(
			"relay network info",
			"Token", item.Token,
			"Addr", item.Addr,
			"Domain", item.Domain,
			"Name", item.Name,
		)
	}

	utils.ClusterID = relayClusterID
	return nil
}

func setupclient(ctx context.Context, log *relaylogger.RelayLog) error {
	configName = "relay-agent-config"
	viper.SetDefault(podNameEnv, "relay-agent")
	viper.SetDefault(podNamespaceEnv, "rafay-system")
	viper.BindEnv(podNameEnv)
	viper.BindEnv(podNamespaceEnv)

	podName = viper.GetString(podNameEnv)
	podNamespace = viper.GetString(podNamespaceEnv)

	utils.PODNAME = podName

	viper.SetDefault(maxDialsEnv, "10")
	viper.BindEnv(maxDialsEnv)
	utils.MaxDials = viper.GetInt(maxDialsEnv)

	viper.SetDefault(dialoutProxyEnv, "")
	viper.BindEnv(dialoutProxyEnv)

	if u, err := url.Parse(viper.GetString(dialoutProxyEnv)); err == nil {
		utils.DialoutProxy = u.Host
	} else {
		utils.DialoutProxy = viper.GetString(dialoutProxyEnv)
	}

	viper.SetDefault(dialoutProxyAuthEnv, "")
	viper.BindEnv(dialoutProxyAuthEnv)
	proxyAuth := viper.GetString(dialoutProxyAuthEnv)
	utils.DialoutProxyAuth = proxyAuth

	if utils.MaxDials < utils.MinDials {
		utils.MaxDials = utils.MinDials
	}

	viper.SetDefault(configNameEnv, "")
	viper.BindEnv(configNameEnv)
	customeRelayCfgMap := viper.GetString(configNameEnv)
	if customeRelayCfgMap != "" {
		configName = customeRelayCfgMap
	}

	err := getRelayAgentConfig(ctx)
	if err != nil {
		log.Error(
			err,
			"failed to get relay config",
		)
		return err
	}
	return nil
}

func setupCDClient(ctx context.Context, log *relaylogger.RelayLog) error {

	viper.SetDefault(podNameEnv, "relay-cdagent")
	viper.SetDefault(podNamespaceEnv, "rafay-system")
	viper.BindEnv(podNameEnv)
	viper.BindEnv(podNamespaceEnv)

	podName = viper.GetString(podNameEnv)
	podNamespace = viper.GetString(podNamespaceEnv)

	utils.PODNAME = podName

	viper.SetDefault(maxDialsEnv, "2")
	viper.BindEnv(maxDialsEnv)
	// utils.MaxDials = viper.GetInt(maxDialsEnv)

	viper.SetDefault(dialoutProxyEnv, "")
	viper.BindEnv(dialoutProxyEnv)
	// utils.DialoutProxy = viper.GetString(dialoutProxyEnv)

	viper.SetDefault(dialoutProxyAuthEnv, "")
	viper.BindEnv(dialoutProxyAuthEnv)
	// proxyAuth := viper.GetString(dialoutProxyAuthEnv)
	// utils.DialoutProxyAuth = base64.StdEncoding.EncodeToString([]byte(proxyAuth))

	// if utils.MaxDials < 1 {
	// 	utils.MaxDials = 2
	// }

	err := getRelayCDAgentConfig(ctx)
	if err != nil {
		log.Error(
			err,
			"failed to get relay config",
		)
		return err
	}
	return nil
}

func getRelayCDAgentConfig(ctx context.Context) error {
	var ok bool

	configBytes, err := ioutil.ReadFile(cdConfigPath)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return err
	}

	if relayAgentID, ok = config[relayAgentIDName].(string); !ok {
		return fmt.Errorf("failed to get cd-relayClusterID")
	}

	if maxDialsStr, ok := config["maxDials"].(string); ok {
		maxDials, err := strconv.Atoi(maxDialsStr)
		if err != nil {
			maxDials = 2
		}
		utils.MaxDials = maxDials
	}

	if utils.MaxDials < utils.MinDials {
		utils.MaxDials = utils.MinDials
	}

	var dialoutProxy string
	if dp, ok := config["dialoutProxy"]; ok {
		if dialoutProxy, ok = dp.(string); ok {
		}
	}

	if dialoutProxy == "" {
		if u, err := url.Parse(viper.GetString(dialoutProxyEnv)); err == nil {
			dialoutProxy = u.Host
		} else {
			dialoutProxy = viper.GetString(dialoutProxyEnv)
		}
	}

	utils.DialoutProxy = dialoutProxy

	var dialoutProxyAuth string
	if dpa, ok := config["dialoutProxyAuthentication"]; ok {
		if dialoutProxyAuth, ok = dpa.(string); ok {
		}
	}

	if dialoutProxyAuth == "" {
		dialoutProxyAuth = viper.GetString(dialoutProxyAuthEnv)
	}

	utils.DialoutProxyAuth = dialoutProxyAuth

	if relays, ok := config[relayRegisterTokenName].(interface{}); ok {
		relayRegisterTokenBytes, err := json.Marshal(relays)
		if err != nil {
			return fmt.Errorf("failed to get cd-relayRegisterToken: %s", err.Error())
		}
		relayRegisterToken = string(relayRegisterTokenBytes)
	} else {
		return fmt.Errorf("failed to get cd-relayRegisterToken")
	}

	if err := processRelayConfigData(relayRegisterToken); err != nil {
		return fmt.Errorf("failed to process cd-relayRegisterToken %s", relayRegisterToken)
	}
	if len(utils.RelayNetworks) <= 0 {
		return fmt.Errorf("empty cd-relaynetwork configmap %s", relayRegisterToken)
	}

	log.Info(
		"relay cd agent config values",
		"podName", podName,
		"podNamespace", podNamespace,
	)

	for _, item := range utils.RelayNetworks {
		log.Info(
			"relay network info",
			"Token", item.Token,
			"Addr", item.Addr,
			"Domain", item.Domain,
			"Name", item.Name,
		)
	}

	utils.AgentID = relayAgentID
	return nil
}

// registerRelayAgent will register with rafay-core-relay-connector template token
// registration fetches client-certificate/root-ca to connect to relay server
func registerRelayAgent(ctx context.Context, rn utils.Relaynetwork) error {
	cfg := &register.Config{
		TemplateToken: rn.TemplateToken,
		Addr:          rn.Addr,
		ClientID:      rn.Token,
		ClientIP:      utils.GetRelayIP(),
		Name:          podName,
	}

	if utils.IsHTTPS(rn.Addr) {
		cfg.Scheme = "https"
	}

	log.Info("config:", cfg)

	if err := register.Register(ctx, cfg); err != nil {
		log.Error(
			err,
			"failed to register relay agent",
		)
		return err
	}

	log.Info(
		"certificate:",
		string(cfg.Certificate),
	)

	rc := utils.RelayNetworkConfig{}
	rc.Network = rn
	rc.RelayAgentCACert = cfg.CACertificate
	rc.RelayAgentCert = cfg.Certificate
	rc.RelayAgentKey = cfg.PrivateKey

	utils.RelayAgentConfig[rn.Name] = rc

	return nil
}

func checkConfigMapChange(configMap string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	h, err := getConfigHash(ctx, configMap)
	if err != nil {
		log.Error(
			err,
			"failed to get relay configmap hash",
		)
		return false
	}

	if lastHash != "" {
		if lastHash != h {
			//hash changed
			log.Info("hash of configmap data changed", "prev", lastHash, "curr", h)
			lastHash = h
			return true
		}
	} else {
		log.Info("setting hash of configmap data", "hash", h)
		lastHash = h
	}

	return false
}

// RunRelayCDAgent entry to the relay client
func RunRelayCDAgent(ctx context.Context, logLvl int) {
	lastHash = ""
	utils.RelayNetworks = nil
	utils.RelayAgentConfig = make(map[string]utils.RelayNetworkConfig)

	log = relaylogger.NewLogger(logLvl).WithName("Relay Agent")

	c, err := clientutil.New()
	if err != nil {
		log.Error(
			err,
			"failed in clientutil new",
		)
		utils.ExitChan <- true
		return
	}

	if err := proxy.InitUnixCacheRoundTripper(); err != nil {
		log.Error(
			err,
			"failed to init unix cached round tripper",
		)
		utils.ExitChan <- true
		return
	}

	applier = apply.NewApplier(c)

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := setupCDClient(rctx, log); err != nil {
		log.Error(
			err,
			"relay agent exiting",
		)
		utils.ExitChan <- true
		return
	}

	for _, rn := range utils.RelayNetworks {
		go handleRelayNetworks(rctx, rn)
	}

	//watch config changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err, "unable to create watcher for config file")
		utils.ExitChan <- true
		return
	}
	defer watcher.Close()
	err = watcher.Add(cdConfigPath)
	if err != nil {
		log.Error(err, "unable to add config file to watcher")
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if ok {
				log.Info("event:", event, "modified file:", event.Name)
			}
			if event.Op != fsnotify.Chmod {
				utils.ExitChan <- true
				return
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				utils.ExitChan <- true
				return
			}
			log.Info("error:", err)
		case <-ctx.Done():
			log.Info(
				"Relay agent exiting",
			)
			return
		}
	}

}

// RunRelayKubeCTLAgent entry to the relay client
func RunRelayKubeCTLAgent(ctx context.Context, logLvl int) {
	lastHash = ""
	utils.RelayNetworks = nil
	utils.RelayAgentConfig = make(map[string]utils.RelayNetworkConfig)

	log = relaylogger.NewLogger(logLvl).WithName("Relay Agent")

	c, err := clientutil.New()
	if err != nil {
		log.Error(
			err,
			"failed in clientutil new",
		)
		utils.TerminateChan <- true
		return
	}

	if err := proxy.InitUnixCacheRoundTripper(); err != nil {
		log.Error(
			err,
			"failed to init unix cached round tripper",
		)
		utils.ExitChan <- true
		return
	}

	applier = apply.NewApplier(c)

	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := setupclient(rctx, log); err != nil {
		log.Error(
			err,
			"relay agent exiting",
		)
		utils.ExitChan <- true
		return
	}

	go cleanup.StaleAuthz(rctx)

	for _, rn := range utils.RelayNetworks {
		go handleRelayNetworks(rctx, rn)
	}

	//ticker configmap changes
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			if checkConfigMapChange(configName) {
				// cancel the ctx
				ticker.Stop()
				utils.ExitChan <- true
				return
			}
		case <-ctx.Done():
			log.Info(
				"Relay agent exiting",
			)
			return
		}
	}

}

func handleRelayNetworks(ctx context.Context, rn utils.Relaynetwork) {
	rnExitChan := make(chan bool)
restartRegister:
	ticker := time.NewTicker(5 * time.Second)
	if err := registerRelayAgent(ctx, rn); err != nil {
		//wait until config is available
		for {
			select {
			case <-ticker.C:
				goto restartRegister
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}
	ticker.Stop()

restartNetwork:
	go tunnel.StartClient(ctx, log, "", utils.RelayAgentConfig[rn.Name], rnExitChan)

	for {
		select {
		case <-rnExitChan:
			log.Info(
				"relay network stopped, restart in 5 secs",
				"name", rn.Name,
			)
			time.Sleep(2 * time.Second)
			goto restartNetwork
		case <-ctx.Done():
			log.Info(
				"Relay network exiting",
				"name", rn.Name,
			)
			return
		}
	}

}
