package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/client/queryproto"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient queryproto.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = queryproto.NewQueryClient(suite.QueryHelper)
}

func (suite *KeeperTestSuite) prepareCustomBalancerPool(
	balances sdk.Coins,
	poolAssets []balancertypes.PoolAsset,
	poolParams balancer.PoolParams,
) uint64 {
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, balances)
	}

	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], poolParams, poolAssets, ""),
	)
	suite.Require().NoError(err)

	return poolID
}
