package cninetwork

import "fmt"

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

func (m *CNIManager) GenerateJSON() []byte {
	return []byte(fmt.Sprintf(
		cniconf,
		m.config.NetworkName,
		m.config.BridgeName,
		m.config.SubNet,
		m.config.NetworkSavePath))
}
