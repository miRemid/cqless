package types

import (
	"os"
	"path"
	"strings"
	"sync"
	"time"

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
	viper.SetDefault("network", NetworkConfig{
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
	viper.SetDefault("logger", LoggerConfig{
		SavePath:       path.Join(DEFAULT_CONFIG_PATH, defaultLogPath),
		EnableSaveFile: true,
		MaxBackups:     5,
		MaxSize:        500,
		MaxAge:         7,
	})
	viper.SetDefault("proxy", ProxyConfig{
		Timeout:             10 * time.Second,
		MaxIdleConns:        30,
		MaxIdleConnsPerHost: 30,
	})
	viper.SetDefault("gateway", GatewayConfig{
		Port:         5566,
		APIPort:      5567,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Provider:     "docker",
		Resolver: &ResolverConfig{
			Type:   "random",
			Params: map[string]string{},
		},
		EnableRateLimit: false,
		RateLimit: &RateLimitConfig{
			Limit: 10,
			Burst: 1000,
		},
		EnablePprof: true,
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

type CQLessConfig struct {
	Network *NetworkConfig `yaml:"network" mapstructure:"network"`
	Logger  *LoggerConfig  `yaml:"logger" mapstructure:"logger"`
	Proxy   *ProxyConfig   `yaml:"proxy" mapstructure:"proxy"`
	Gateway *GatewayConfig `yaml:"gateway" mapstructure:"gateway"`
	CQHTTP  *CQHTTPConfig  `yaml:"cqhttp" mapstructure:"cqhttp"`
}

type NetworkConfig struct {
	// BinaryPath CNI插件的绝对路径
	BinaryPath string `yaml:"binary_path" mapstructure:"binary_path"`
	// ConfigPath 存放CNI配置文件的路径
	ConfigPath string `yaml:"config_path" mapstructure:"config_path"`
	// ConfigFileName 配置文件名称
	ConfigFileName string `yaml:"config_file_name" mapstructure:"config_file_name"`
	// NetworkSavePath 函数容器网络配置文件保存路径
	NetworkSavePath string `yaml:"network_save_path" mapstructure:"network_save_path"`
	// NamespaceFormat 为容器创建的Namespace格式，仅支持一个字符串类型的占位符如cqless-%s
	NamespaceFormat string `yaml:"namespace_format" mapstructure:"namespace_format"`

	// 以下为生成默认配置文件字段

	// NetworkName 网络名称
	NetworkName string `yaml:"network_name" mapstructure:"network_name"`
	// BridgeName 网桥名称
	BridgeName string `yaml:"bridge_name" mapstructure:"bridge_name"`
	// SubNet 网桥子网
	SubNet string `yaml:"subnet" mapstructure:"subnet"`
	// IfPrefix 虚拟网卡的前缀
	IfPrefix string `yaml:"if_prefix" mapstructure:"if_prefix"`
}

type LoggerConfig struct {
	EnableSaveFile bool   `yaml:"enable_save_file" mapstructure:"enable_save_file"`
	SavePath       string `yaml:"save_path" mapstructure:"save_path"`

	MaxBackups int `yaml:"max_backups" mapstructure:"max_backups"`
	MaxSize    int `yaml:"max_size" mapstructure:"max_size"`
	MaxAge     int `yaml:"max_age" mapstructure:"max_age"`
}

type ProxyConfig struct {
	Timeout             time.Duration `yaml:"timeout" mapstructure:"timeout"`
	MaxIdleConns        int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" mapstructure:"max_idle_conns_per_host"`
}

type GatewayConfig struct {
	Provider     string          `yaml:"provider_type" mapstructure:"provider_type"`
	Port         int             `yaml:"port" mapstructure:"port"`
	APIPort      int             `yaml:"api_port" mapstructure:"api_port"`
	ReadTimeout  time.Duration   `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration   `yaml:"write_timeout" mapstructure:"write_timeout"`
	Resolver     *ResolverConfig `yaml:"resolver" mapstructure:"resolver"`

	EnableRateLimit bool             `yaml:"enable_rate_limit" mapstructure:"enable_rate_limit"`
	RateLimit       *RateLimitConfig `yaml:"rate_limit" mapstructure:"rate_limit"`

	EnablePprof bool `yaml:"pprof" mapstructure:"pprof"`
}

type RateLimitConfig struct {
	Limit float64 `yaml:"limit" mapstructure:"limit"`
	Burst int     `yaml:"burst" mapstructure:"burst"`
}

type ResolverConfig struct {
	Type   string            `yaml:"type" mapstructure:"type"`
	Params map[string]string `yaml:"params" mapstructure:"params"`
}

type CQHTTPConfig struct {
	AuthToken string `yaml:"auth_token" mapstructure:"auth_token"`
}
