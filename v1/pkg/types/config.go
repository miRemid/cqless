package types

import (
	"os"
	"path"
	"strings"
	"sync"
	"time"

	ntypes "github.com/miRemid/cqless/v1/pkg/cninetwork/types"
	ctypes "github.com/miRemid/cqless/v1/pkg/cqhttp/types"
	gtypes "github.com/miRemid/cqless/v1/pkg/gateway/types"
	ltypes "github.com/miRemid/cqless/v1/pkg/logger/types"
	ptypes "github.com/miRemid/cqless/v1/pkg/provider/types"
	proxyTypes "github.com/miRemid/cqless/v1/pkg/proxy/types"
	rtypes "github.com/miRemid/cqless/v1/pkg/resolver/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const (
	defaultCNIPath = "cni"
	defaultLogPath = "log"
)

var (
	home                              = os.Getenv("HOME")
	DEFAULT_CONFIG_PATH               = path.Join(home, DEFAULT_SAVE_PATH)
	config              *CQLessConfig = nil
	mutex                             = sync.Mutex{}
	DEBUG               string
)

type CQLessConfig struct {
	Network  *ntypes.NetworkOption   `yaml:"network" mapstructure:"network"`
	Logger   *ltypes.LoggerOption    `yaml:"logger" mapstructure:"logger"`
	Gateway  *gtypes.GatewayOption   `yaml:"gateway" mapstructure:"gateway"`
	Proxy    *proxyTypes.ProxyOption `yaml:"proxy" mapstructure:"proxy"`
	Resolver *rtypes.ResolverOption  `yaml:"resolver" mapstructure:"resolver"`
	Provider *ptypes.ProviderOption  `yaml:"provider" mapstructure:"provider"`
	CQHTTP   *ctypes.CQHTTPOption    `yaml:"cqhttp" mapstructure:"cqhttp"`
}

func GetConfig() *CQLessConfig {
	if config == nil {
		mutex.Lock()
		if config == nil {
			initConfig()
		}
		mutex.Unlock()
	}
	return config
}

func initConfig() {
	cfg := new(CQLessConfig)
	if err := os.MkdirAll(DEFAULT_CONFIG_PATH, 0775); err != nil {
		panic(err)
	}
	viper.SetConfigName("cqless")
	viper.SetConfigType("yaml")
	viper.SetDefault("network", ntypes.NetworkOption{
		BinaryPath:      "/opt/cni/bin",
		ConfigPath:      path.Join(DEFAULT_CONFIG_PATH, defaultCNIPath),
		ConfigFileName:  "10-cqless.conflist",
		NetworkSavePath: path.Join(DEFAULT_CONFIG_PATH, defaultCNIPath),
		NamespaceFormat: "cqless-%s",
		IfPrefix:        "cqeth",
		NetworkName:     "cqless-cni-bridge",
		BridgeName:      "cqless0",
		SubNet:          "10.72.0.0/16",
	})
	viper.SetDefault("logger", ltypes.LoggerOption{
		SavePath:       path.Join(DEFAULT_CONFIG_PATH, defaultLogPath),
		EnableSaveFile: true,
		MaxBackups:     5,
		MaxSize:        500,
		MaxAge:         7,
	})
	viper.SetDefault("gateway", gtypes.GatewayOption{
		Address:         "127.0.0.1:5565",
		HTTPAddress:     "127.0.0.1:5566",
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    60 * time.Second,
		EnableRateLimit: false,
		RateLimit: &gtypes.RateLimitOption{
			Limit: 10,
			Burst: 1000,
		},
		EnablePprof: true,
	})
	viper.SetDefault("proxy", proxyTypes.ProxyOption{
		Address:             "127.0.0.1:5567",
		Timeout:             10 * time.Second,
		MaxIdleConns:        30,
		MaxIdleConnsPerHost: 30,
	})
	viper.SetDefault("cqhttp", ctypes.CQHTTPOption{
		AuthToken: "",
	})
	viper.SetDefault("resolver", rtypes.ResolverOption{
		StorageOption: &rtypes.StorageOption{
			Strategy:     rtypes.STORAGE_LOCAL,
			DBPath:       DEFAULT_CONFIG_PATH,
			RpcEndpoints: []string{"127.0.0.1:2379"},
			DialTimeout:  2 * time.Second,
		},
		SelectorOption: &rtypes.SelectorOption{
			Strategy: rtypes.SELECTOR_RANDOM,
		},
	})
	viper.SetDefault("provider", ptypes.ProviderOption{
		Strategy: ptypes.PROVIDER_DOCKER,
	})
	viper.AddConfigPath(DEFAULT_CONFIG_PATH)
	if err := viper.SafeWriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			panic(err)
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(cfg, viper.DecoderConfigOption(func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "yaml"
	})); err != nil {
		panic(err)
	}
	config = cfg
	d, ok := os.LookupEnv("DEBUG")
	if ok {
		DEBUG = strings.ToUpper(d)
	} else {
		DEBUG = "TRUE"
	}
}
