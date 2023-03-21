package docker

var (
	p *DockerProvider
)

func init() {
	p = NewProvider()
	p.Init()
}
