package v15

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icqkeeper "github.com/strangelove-ventures/async-icq/v4/keeper"
	icqtypes "github.com/strangelove-ventures/async-icq/v4/types"

	"github.com/osmosis-labs/osmosis/v14/wasmbinding"
	ibcratelimit "github.com/osmosis-labs/osmosis/v14/x/ibc-rate-limit"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/v14/app/keepers"
	appParams "github.com/osmosis-labs/osmosis/v14/app/params"
	"github.com/osmosis-labs/osmosis/v14/app/upgrades"
	gammkeeper "github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
	packetforwardtypes "github.com/strangelove-ventures/packet-forward-middleware/v4/router/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		poolmanagerParams := poolmanagertypes.NewParams(keepers.GAMMKeeper.GetParams(ctx).PoolCreationFee)

		keepers.PoolManagerKeeper.SetParams(ctx, poolmanagerParams)
		keepers.PacketForwardKeeper.SetParams(ctx, packetforwardtypes.DefaultParams())
		setICQParams(ctx, keepers.ICQKeeper)

		// N.B: pool id in gamm is to be deprecated in the future
		// Instead,it is moved to poolmanager.
		migrateNextPoolId(ctx, keepers.GAMMKeeper, keepers.PoolManagerKeeper)

		//  N.B.: this is done to avoid initializing genesis for poolmanager module.
		// Otherwise, it would overwrite migrations with InitGenesis().
		// See RunMigrations() for details.
		fromVM[poolmanagertypes.ModuleName] = 0

		// Metadata for uosmo and uion were missing prior to this upgrade.
		// They are added in this upgrade.
		registerOsmoIonMetadata(ctx, keepers.BankKeeper)

		// Why do E2E tests pass when there isn't a rate limiting contract set?
		setRateLimits(ctx, keepers.AccountKeeper, keepers.RateLimitingICS4Wrapper, keepers.WasmKeeper)

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func setICQParams(ctx sdk.Context, icqKeeper *icqkeeper.Keeper) {
	icqparams := icqtypes.DefaultParams()
	icqparams.AllowQueries = wasmbinding.GetStargateWhitelistedPaths()
	// Adding SmartContractState query to allowlist
	icqparams.AllowQueries = append(icqparams.AllowQueries, "/cosmwasm.wasm.v1.Query/SmartContractState")
	icqKeeper.SetParams(ctx, icqparams)
}

func setRateLimits(ctx sdk.Context, accountKeeper *authkeeper.AccountKeeper, rateLimitingICS4Wrapper *ibcratelimit.ICS4Wrapper, wasmKeeper *wasmkeeper.Keeper) {
	govModule := accountKeeper.GetModuleAddress(govtypes.ModuleName)
	contractKeeper := wasmkeeper.NewGovPermissionKeeper(wasmKeeper)

	paths := []string{
		`{"add_path": {"channel_id": "any", "denom": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
          "quotas":
            [
              {"name":"ATOM-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"ATOM-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
		`{"add_path": {"channel_id": "any", "denom": "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
          "quotas":
            [
              {"name":"USDC-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"USDC-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
		`{"add_path": {"channel_id": "any", "denom": "ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F",
          "quotas":
            [
              {"name":"WBTC-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"WBTC-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
		`{"add_path": {"channel_id": "any", "denom": "ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5",
          "quotas":
            [
              {"name":"WETH-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"WETH-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
		`{"add_path": {"channel_id": "any", "denom": "ibc/6AE98883D4D5D5FF9E50D7130F1305DA2FFA0C652D1DD9C123657C6B4EB2DF8A",
          "quotas":
            [
              {"name":"EVMOS-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"EVMOS-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
		`{"add_path": {"channel_id": "any", "denom": "ibc/987C17B11ABC2B20019178ACE62929FE9840202CE79498E29FE8E5CB02B7C0A4",
          "quotas":
            [
              {"name":"STARS-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"STARS-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
		`{"add_path": {"channel_id": "any", "denom": "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7",
          "quotas":
            [
              {"name":"DAI-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"DAI-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
		`{"add_path": {"channel_id": "any", "denom": "ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED",
          "quotas":
            [
              {"name":"JUNO-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"JUNO-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
		`{"add_path": {"channel_id": "any", "denom": "ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1",
          "quotas":
            [
              {"name":"CRO-DAY","duration":86400,"send_recv":[30,30]},
              {"name":"CRO-WEEK","duration":604800,"send_recv":[60,60]}
            ]
          }}`,
	}

	contract := rateLimitingICS4Wrapper.GetParams(ctx)
	fmt.Println("UPGRADE: rate limiting contract", contract)
	if contract == "" {
		panic("rate limiting contract not set")
	}
	rateLimitingContract, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		panic("contract address improperly formatted")
	}
	for _, denom := range paths {
		fmt.Println("UPGRADE: setting path for", denom)
		_, err := contractKeeper.Execute(ctx, rateLimitingContract, govModule, []byte(denom), nil)
		if err != nil {
			panic(err)
		}
	}
}

func migrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanager.Keeper) {
	// N.B: pool id in gamm is to be deprecated in the future
	// Instead,it is moved to poolmanager.
	// nolint: staticcheck
	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	poolmanagerKeeper.SetNextPoolId(ctx, nextPoolId)

	for poolId := uint64(1); poolId < nextPoolId; poolId++ {
		poolType, err := gammKeeper.GetPoolType(ctx, poolId)
		if err != nil {
			panic(err)
		}

		poolmanagerKeeper.SetPoolRoute(ctx, poolId, poolType)
	}
}

func registerOsmoIonMetadata(ctx sdk.Context, bankKeeper bankkeeper.Keeper) {
	uosmoMetadata := banktypes.Metadata{
		Description: "The native token of Osmosis",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    appParams.BaseCoinUnit,
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    appParams.HumanCoinUnit,
				Exponent: appParams.OsmoExponent,
				Aliases:  nil,
			},
		},
		Base:    appParams.BaseCoinUnit,
		Display: appParams.HumanCoinUnit,
	}

	uionMetadata := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uion",
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    "ion",
				Exponent: 6,
				Aliases:  nil,
			},
		},
		Base:    "uion",
		Display: "ion",
	}

	bankKeeper.SetDenomMetaData(ctx, uosmoMetadata)
	bankKeeper.SetDenomMetaData(ctx, uionMetadata)
}