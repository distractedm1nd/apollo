package util

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/celestiaorg/celestia-node/libs/utils"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/cmwaters/apollo"
	"github.com/cmwaters/apollo/node/consensus"
	rpcclient "github.com/tendermint/tendermint/rpc/client/http"
)

func GetTrustedHash(ctx context.Context, rpcEndpoint string) (string, error) {
	client, err := rpcclient.New(rpcEndpoint, "/websocket")
	if err != nil {
		return "", fmt.Errorf("failed to create RPC client: %w", err)
	}
	firstHeight := int64(1)
	header, err := client.Header(ctx, &firstHeight)
	if err != nil {
		return "", fmt.Errorf("failed to query header at height 1: %w", err)
	}

	return header.Header.Hash().String(), nil
}

func ParsePort(endpoint string) (string, error) {
	split := strings.Split(endpoint, ":")
	if len(split) == 0 {
		return "", fmt.Errorf("failed to parse port from endpoint: %s", endpoint)
	}
	port := strings.Split(split[len(split)-1], "/")[0]

	if _, err := strconv.Atoi(port); err != nil {
		return "", fmt.Errorf("failed to parse port from endpoint: %s", endpoint)
	}

	return port, nil
}

// ConfigureRandomConsensusEndpoint configures the node's consensus endpoint to a random endpoint.
func ConfigureRandomConsensusEndpoint(ctx context.Context, inputs apollo.Endpoints, cfg *nodebuilder.Config) (*nodebuilder.Config, error) {
	consensusRPCEndpoint, err := inputs.GetRandom(consensus.RPCEndpointLabel)
	if err != nil {
		return nil, fmt.Errorf("failed to get consensus RPC endpoint: %w", err)
	}

	consensusGRPCEndpoint, err := inputs.GetRandom(consensus.GRPCEndpointLabel)
	if err != nil {
		return nil, fmt.Errorf("failed to get consensus GRPC endpoint: %w", err)
	}

	headerHash, err := GetTrustedHash(ctx, consensusRPCEndpoint)
	if err != nil {
		return nil, err
	}
	cfg.Header.TrustedHash = headerHash

	consensusIP, err := utils.ValidateAddr(consensusRPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse consensus RPC endpoint: %w", err)
	}
	cfg.Core.IP = consensusIP

	rpcPort, err := ParsePort(consensusRPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse consensus RPC endpoint: %w", err)
	}
	cfg.Core.RPCPort = rpcPort

	grpcPort, err := ParsePort(consensusGRPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse consensus GRPC endpoint: %w", err)
	}
	cfg.Core.GRPCPort = grpcPort

	return cfg, nil
}
