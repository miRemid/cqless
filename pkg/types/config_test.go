package types

import (
	"testing"

	"github.com/spf13/viper"
	"gotest.tools/v3/assert"
)

func Test_readConfig(t *testing.T) {
	config := CQLessConfig{}
	assert.NilError(t, viper.ReadInConfig())
	assert.NilError(t, viper.Unmarshal(&config))
	t.Log(config.Network.BinaryPath)
	t.Log(config.Network.NetworkSavePath)
}
