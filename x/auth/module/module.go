package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/compat"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	types2 "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

var (
	_ app.TypeProvider = Module{}
	_ app.Provisioner  = Module{}
)

type Inputs struct {
	dig.In

	Codec            codec.Codec
	KeyProvider      app.KVStoreKeyProvider
	SubspaceProvider types2.SubspaceProvider
}

type Outputs struct {
	dig.Out

	Handler    app.Handler `group:"app.handler"`
	ViewKeeper types.ViewKeeper
	Keeper     types.Keeper `security-role:"admin"`
}

type CLIOutputs struct {
	dig.Out

	TxCmd    *cobra.Command   `group:"cosmos.tx.v1.Command"`
	QueryCmd []*cobra.Command `group:"cosmos.query.v1.Command,flatten"`
}

func (m Module) RegisterTypes(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (m Module) Provision(key app.ModuleKey, registrar container.Registrar) error {
	// provide AccountRetriever
	err := registrar.Provide(func() client.AccountRetriever {
		return types.AccountRetriever{}
	})
	if err != nil {
		return err
	}

	// provide CLI handlers
	err = registrar.Provide(func() CLIOutputs {
		am := auth.AppModuleBasic{}
		return CLIOutputs{
			TxCmd: am.GetTxCmd(),
			QueryCmd: []*cobra.Command{
				am.GetQueryCmd(),
				authcmd.GetAccountCmd(),
			},
		}
	})
	if err != nil {
		return err
	}

	// provide app handlers
	return registrar.Provide(
		func(inputs Inputs) (Outputs, error) {
			var accCtr types.AccountConstructor
			if m.AccountConstructor != nil {
				err := inputs.Codec.UnpackAny(m.AccountConstructor, &accCtr)
				if err != nil {
					return Outputs{}, err
				}
			} else {
				accCtr = DefaultAccountConstructor{}
			}

			perms := map[string][]string{}
			for _, perm := range m.Permissions {
				perms[perm.Address] = perm.Permissions
			}

			var randomGenesisAccountsProvider types.RandomGenesisAccountsProvider
			if m.RandomGenesisAccountsProvider != nil {
				err := inputs.Codec.UnpackAny(m.RandomGenesisAccountsProvider, &randomGenesisAccountsProvider)
				if err != nil {
					return Outputs{}, err
				}
			} else {
				randomGenesisAccountsProvider = DefaultRandomGenesisAccountsProvider{}
			}

			keeper := authkeeper.NewAccountKeeper(
				inputs.Codec,
				inputs.KeyProvider(key),
				inputs.SubspaceProvider(key),
				func() types.AccountI {
					return accCtr.NewAccount()
				},
				perms,
			)
			appMod := auth.NewAppModule(inputs.Codec, keeper, func(simState *module.SimulationState) types.GenesisAccounts {
				return randomGenesisAccountsProvider.RandomGenesisAccounts(simState)
			})

			return Outputs{
				ViewKeeper: viewOnlyKeeper{keeper},
				Keeper:     keeper,
				Handler:    compat.AppModuleHandler(key.ID(), appMod),
			}, nil
		},
	)
}

func (m DefaultAccountConstructor) NewAccount() types.AccountI {
	return &types.BaseAccount{}
}

func (m DefaultRandomGenesisAccountsProvider) RandomGenesisAccounts(simState *module.SimulationState) types.GenesisAccounts {
	return simulation.RandomGenesisAccounts(simState)
}