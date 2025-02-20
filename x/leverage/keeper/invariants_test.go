package keeper_test

import (
	appparams "github.com/tessornetwork/nebula/v3/app/params"
	"github.com/tessornetwork/nebula/v3/x/leverage/keeper"
	"github.com/tessornetwork/nebula/v3/x/leverage/types"
)

func (s *IntegrationTestSuite) TestReserveAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// artificially set reserves
	s.setReserves(coin(appparams.BondDenom, 300_000000))

	// check invariants
	_, broken := keeper.ReserveAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}

func (s *IntegrationTestSuite) TestCollateralAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 NEBULA
	addr := s.newAccount(coin(nebulaDenom, 1000_000000))
	s.supply(addr, coin(nebulaDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+nebulaDenom, 1000_000000))

	// check invariant
	_, broken := keeper.InefficientCollateralAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)

	uTokenDenom := types.ToUTokenDenom(appparams.BondDenom)

	// withdraw the supplied nebula in the initBorrowScenario
	s.withdraw(addr, coin(uTokenDenom, 1000_000000))

	// check invariant
	_, broken = keeper.InefficientCollateralAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}

func (s *IntegrationTestSuite) TestBorrowAmountInvariant() {
	app, ctx, require := s.app, s.ctx, s.Require()

	// creates account which has supplied and collateralized 1000 NEBULA
	addr := s.newAccount(coin(nebulaDenom, 1000_000000))
	s.supply(addr, coin(nebulaDenom, 1000_000000))
	s.collateralize(addr, coin("u/"+nebulaDenom, 1000_000000))

	// user borrows 20 nebula
	s.borrow(addr, coin(appparams.BondDenom, 20_000000))

	// check invariant
	_, broken := keeper.InefficientBorrowAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)

	// user repays 30 nebula, actually only 20 because is the min between
	// the amount borrowed and the amount repaid
	_, err := app.LeverageKeeper.Repay(ctx, addr, coin(appparams.BondDenom, 30_000000))
	require.NoError(err)

	// check invariant
	_, broken = keeper.InefficientBorrowAmountInvariant(app.LeverageKeeper)(ctx)
	require.False(broken)
}
