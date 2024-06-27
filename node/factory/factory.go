package factory

import (
	"strconv"

	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/cmwaters/apollo"
	"github.com/cmwaters/apollo/node/bridge"
	"github.com/cmwaters/apollo/node/light"
)

type NodeType int

const (
	Consensus NodeType = iota
	Bridge
	Light
)

func CreateServices(lightCount int, bridgeCount int) []apollo.Service {
	var services []apollo.Service
	rpcPort := 26658
	for i := 0; i < bridgeCount; i++ {
		cfg := nodebuilder.DefaultConfig(node.Bridge)
		cfg.RPC.Port = strconv.Itoa(rpcPort)
		services = append(services, bridge.New(cfg))
		rpcPort++
	}
	for i := 0; i < lightCount; i++ {
		cfg := nodebuilder.DefaultConfig(node.Light)
		cfg.RPC.Port = strconv.Itoa(rpcPort)
		services = append(services, light.New(cfg))
		rpcPort++
	}
	return services
}
