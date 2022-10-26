package swaprouter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey sdk.StoreKey

	gammKeeper          types.SimulationExtension
	concentratedKeeper  types.SwapI
	bankKeeper          types.BankI
	accountKeeper       types.AccountI
	communityPoolKeeper types.CommunityPoolI

	poolCreationListeners types.PoolCreationListeners

	routes map[types.PoolType]types.SwapI

	paramSpace paramtypes.Subspace
}

func NewKeeper(storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, gammKeeper types.SimulationExtension, concentratedKeeper types.SwapI, bankKeeper types.BankI, accountKeeper types.AccountI, communityPoolKeeper types.CommunityPoolI) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	routes := map[types.PoolType]types.SwapI{
		types.Balancer:     gammKeeper,
		types.Stableswap:   gammKeeper,
		types.Concentrated: concentratedKeeper,
	}

	return &Keeper{storeKey: storeKey, paramSpace: paramSpace, gammKeeper: gammKeeper, concentratedKeeper: concentratedKeeper, bankKeeper: bankKeeper, accountKeeper: accountKeeper, communityPoolKeeper: communityPoolKeeper, routes: routes}
}

// GetParams returns the total set of swaprouter parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of swaprouter parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// InitGenesis initializes the swaprouter module's state from a provided genesis
// state.
// TODO: test this
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.setNextPoolId(ctx, genState.NextPoolId)
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the swaprouter module's exported genesis.
// TODO: test this
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:     k.GetParams(ctx),
		NextPoolId: k.GetNextPoolId(ctx),
	}
}

// GetPoolAndPoke gets a pool with the given pool id.
// This method is used for simulation only.
// TODO: remove it after refactoring simulation logic.
func (k Keeper) GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.TraditionalAmmInterface, error) {
	return k.gammKeeper.GetPoolAndPoke(ctx, poolId)
}

// GetNextPoolId returns the next pool id.
// This method is used for simulation only.
// TODO: remove it after refactoring simulation logic.
// GetNextPoolId returns the next pool Id.
func (k Keeper) GetNextPoolId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPoolId := gogotypes.UInt64Value{}
	osmoutils.MustGet(store, types.KeyNextGlobalPoolId, &nextPoolId)
	return nextPoolId.Value
}

// Set the pool creation listeners.
func (k *Keeper) SetPoolCreationListeners(listeners types.PoolCreationListeners) *Keeper {
	if k.poolCreationListeners != nil {
		panic("cannot set pool creation listeners twice")
	}

	k.poolCreationListeners = listeners

	return k
}

// setNextPoolId sets next pool Id.
func (k Keeper) setNextPoolId(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextGlobalPoolId, &gogotypes.UInt64Value{Value: poolId})
}
