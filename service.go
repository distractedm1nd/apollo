package apollo

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/cmwaters/apollo/genesis"
	"github.com/tendermint/tendermint/types"
)

var (
	cdc = encoding.MakeConfig(app.ModuleEncodingRegisters...)
)

func Codec() encoding.Config {
	return cdc
}

type Service interface {
	Name() string
	EndpointsNeeded() []string
	EndpointsProvided() []string
	Setup(_ context.Context, dir string, pendingGenesis *types.GenesisDoc) (genesis.Modifier, error)
	Start(_ context.Context, dir string, genesis *types.GenesisDoc, inputs Endpoints) (Endpoints, error)
	Stop(context.Context) error
}

type Endpoints map[string][]string

func (e Endpoints) Add(key string, values ...string) {
	e[key] = append(e[key], values...)
}

// GetRandom returns a random active endpoint for the given key.
func (e Endpoints) GetRandom(key string) (string, error) {
	v, ok := e[key]
	if !ok || len(v) == 0 {
		return "", fmt.Errorf("no endpoints for key %s", key)
	}

	return v[rand.Intn(len(v))], nil
}

func (e Endpoints) String() string {
	var output string
	for name, endpoint := range e {
		output += fmt.Sprintf("%s: %s\t", name, endpoint)
	}
	return output
}
