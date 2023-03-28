package types

import (
	"os"
	"path"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const (
	defaultSavePath = ".local/share/cqless"
	defaultCNIPath  = "cni"
	defaultLogPath  = "log"
)

var (
	home              = os.Getenv("HOME")
	defaultConfigPath = path.Join(home, defaultSavePath)
	config            = new(CQLessConfig)
)

func GetConfig() *CQLessConfig {
	return config
}

func init() {
	if err := os.MkdirAll(defaultConfigPath, 0775); err != nil {
		panic(err)
	}
	viper.SetConfigName("cqless")
	viper.SetConfigType("yaml")
	viper.SetDefault("network", Network{
		BinaryPath:      "/opt/cni/bin",
		ConfigPath:      path.Join(defaultConfigPath, defaultCNIPath),
		ConfigFileName:  "10-cqless.conflist",
		NetworkSavePath: path.Join(defaultConfigPath, defaultCNIPath),
		NamespaceFormat: "cqless-%s",
		IfPrefix:        "cqeth",
		NetworkName:     "cqless-cni-bridge",
		BridgeName:      "cqless0",
		SubNet:          "10.72.0.0/16",
	})
	viper.SetDefault("logger", Logger{
		Debug:    true,
		SavePath: path.Join(defaultConfigPath, defaultLogPath),
	})
	viper.SetDefault("proxy", Proxy{
		Timeout:             10 * time.Second,
		MaxIdleConns:        30,
		MaxIdleConnsPerHost: 30,
	})
	viper.SetDefault("gateway", Gateway{
		Port:         5566,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Provider:     "docker",
	})
	viper.AddConfigPath(defaultConfigPath)
	if err := viper.SafeWriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			panic(err)
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(config, viper.DecoderConfigOption(func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "yaml"
	})); err != nil {
		panic(err)
	}
}

type CQLessConfig struct {
	Network *Network `yaml:"network" mapstructure:"network"`
	Logger  *Logger  `yaml:"logger" mapstructure:"logger"`
	Proxy   *Proxy   `yaml:"proxy" mapstructure:"proxy"`
	Gateway *Gateway `yaml:"gateway" mapstructure:"gateway"`
}

type Network struct {
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

type Logger struct {
	SavePath string `yaml:"save_path" mapstructure:"save_path"`
	Debug    bool   `yaml:"debug_mode" mapstructure:"debug_mode"`
}

type Proxy struct {
	Timeout             time.Duration `yaml:"timeout" mapstructure:"timeout"`
	MaxIdleConns        int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" mapstructure:"max_idle_conns_per_host"`
}

type Gateway struct {
	Provider     string        `yaml:"provider_type" mapstructure:"provider_type"`
	Port         int           `yaml:"port" mapstructure:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
}
