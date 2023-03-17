package cninetwork

import "fmt"

type CNIConfig struct {
	BinDir       string
	ConfDir      string
	NamespaceFmt string
	DataDir      string
	ConfFileName string
	NetworkName  string
	BridgeName   string
	Subnet       string
	IfPrefix     string
}

var cniconf = `
{
    "cniVersion": "0.4.0",
    "name": "%s",
    "plugins": [
      {
        "type": "bridge",
        "bridge": "%s",
        "isGateway": true,
        "ipMasq": true,
        "ipam": {
            "type": "host-local",
            "subnet": "%s",
            "dataDir": "%s",
            "routes": [
                { "dst": "0.0.0.0/0" }
            ]
        }
      },
      {
        "type": "firewall"
      }
    ]
}
`

func (config CNIConfig) GenerateJSON() []byte {
	return []byte(fmt.Sprintf(cniconf, config.NetworkName, config.BridgeName, config.Subnet, config.DataDir))
}
