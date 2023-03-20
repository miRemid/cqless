package docker

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

var (
	p *DockerProvider
)

func init() {
	p = NewProvider()
	p.Init()
}

func Test_Inspect(t *testing.T) {

	id := "1cc6a820e3695d9da7d85861a1e4e7e9b269678bdb60f42179e9f867ffb62b00"

	data, err := p.Inspect(context.Background(), id)
	assert.NilError(t, err)
	log.Info(data.Config.Env)

}
