package light

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/cmwaters/apollo"
	"github.com/cmwaters/apollo/genesis"
	"github.com/cmwaters/apollo/node/bridge"
	"github.com/cmwaters/apollo/node/consensus"
	"github.com/cmwaters/apollo/node/util"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/tendermint/tendermint/types"
)

var _ apollo.Service = &Service{}

const (
	LightServiceName  = "light-node"
	RPCEndpointLabel  = "light-rpc"
	DocsEndpointLabel = "light-api-docs"
	DocsEndpoint      = "https://node-rpc-docs.celestia.org"
)

type Service struct {
	uuid    string
	node    *nodebuilder.Node
	store   nodebuilder.Store
	chainID string
	config  *nodebuilder.Config
}

func New(config *nodebuilder.Config) *Service {
	return &Service{
		uuid:   config.RPC.Port,
		config: config,
	}
}

func (s *Service) Name() string {
	return LightServiceName + "-" + s.uuid
}

func (s *Service) EndpointsNeeded() []string {
	return []string{consensus.RPCEndpointLabel, consensus.GRPCEndpointLabel, bridge.P2PEndpointLabel}
}

func (s *Service) EndpointsProvided() []string {
	return []string{RPCEndpointLabel, DocsEndpointLabel}
}

// TODO: We should automatically fund the light client account so that they can
// start submitting blobs straight away
func (s *Service) Setup(ctx context.Context, dir string, pendingGenesis *types.GenesisDoc) (genesis.Modifier, error) {
	return nil, nodebuilder.Init(*s.config, dir, node.Light)
}

func (s *Service) Start(ctx context.Context, dir string, genesis *types.GenesisDoc, inputs apollo.Endpoints) (apollo.Endpoints, error) {
	s.chainID = genesis.ChainID

	_, err := util.ConfigureRandomConsensusEndpoint(ctx, inputs, s.config)
	if err != nil {
		return nil, err
	}

	var bridgeAddrInfo peer.AddrInfo
	peerId, err := inputs.GetRandom(bridge.P2PEndpointLabel)
	if err != nil {
		return nil, fmt.Errorf("failed to get bridge peer ID: %w", err)
	}
	if err := bridgeAddrInfo.UnmarshalJSON([]byte(peerId)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bridge addr info: %w", err)
	}

	bridgeAddrs, err := peer.AddrInfoToP2pAddrs(&bridgeAddrInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bridge addr info to multiaddrs: %w", err)
	}

	// set the trusted peers
	s.config.Header.TrustedPeers = []string{bridgeAddrs[0].String()}

	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)

	keysPath := filepath.Join(dir, "keys")
	ring, err := keyring.New(app.Name, s.config.State.KeyringBackend, keysPath, os.Stdin, encConf.Codec)
	if err != nil {
		return nil, err
	}

	s.store, err = nodebuilder.OpenStore(dir, ring)
	if err != nil {
		return nil, err
	}

	s.node, err = nodebuilder.NewWithConfig(node.Light, p2p.Network(s.chainID), s.store, s.config)
	if err != nil {
		return nil, err
	}

	if err := s.node.Host.Connect(ctx, bridgeAddrInfo); err != nil {
		return nil, fmt.Errorf("failed to connect to bridge node: %w", err)
	}

	var endpoints apollo.Endpoints
	endpoints = make(map[string][]string)
	endpoints.Add(RPCEndpointLabel, fmt.Sprintf("http://localhost:%s", s.config.RPC.Port))
	endpoints.Add(DocsEndpointLabel, DocsEndpoint)

	return endpoints, s.node.Start(ctx)
}

func (s *Service) Stop(ctx context.Context) error {
	if err := s.node.Stop(ctx); err != nil {
		return err
	}
	return s.store.Close()
}
