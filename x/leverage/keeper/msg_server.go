package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/tessornetwork/nebula/v3/x/leverage/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/leverage
// module.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

func (s msgServer) Supply(
	goCtx context.Context,
	msg *types.MsgSupply,
) (*types.MsgSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}
	received, err := s.keeper.Supply(ctx, supplierAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets supplied",
		"supplier", msg.Supplier,
		"supplied", msg.Asset.String(),
		"received", received.String(),
	)
	err = ctx.EventManager().EmitTypedEvent(&types.EventSupply{
		Supplier: msg.Supplier,
		Asset:    msg.Asset,
		Utoken:   received,
	})
	return &types.MsgSupplyResponse{
		Received: received,
	}, err
}

func (s msgServer) Withdraw(
	goCtx context.Context,
	msg *types.MsgWithdraw,
) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}
	received, err := s.keeper.Withdraw(ctx, supplierAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	err = s.logWithdrawal(ctx, msg.Supplier, msg.Asset, received, "supplied assets withdrawn")
	return &types.MsgWithdrawResponse{
		Received: received,
	}, err
}

func (s msgServer) MaxWithdraw(
	goCtx context.Context,
	msg *types.MsgMaxWithdraw,
) (*types.MsgMaxWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}

	uToken, err := s.keeper.maxWithdraw(ctx, supplierAddr, msg.Denom)
	if err != nil {
		return nil, err
	}

	received, err := s.keeper.Withdraw(ctx, supplierAddr, uToken)
	if err != nil {
		return nil, err
	}

	err = s.logWithdrawal(ctx, msg.Supplier, uToken, received, "maximum supplied assets withdrawn")
	return &types.MsgMaxWithdrawResponse{
		Withdrawn: uToken,
		Received:  received,
	}, err
}

func (s msgServer) logWithdrawal(ctx sdk.Context, supplier string, redeemed, received sdk.Coin, desc string) error {
	s.keeper.Logger(ctx).Debug(
		desc,
		"supplier", supplier,
		"redeemed", redeemed.String(),
		"received", received.String(),
	)
	err := ctx.EventManager().EmitTypedEvent(&types.EventWithdraw{
		Supplier: supplier,
		Utoken:   redeemed,
		Asset:    received,
	})
	return err
}

func (s msgServer) Collateralize(
	goCtx context.Context,
	msg *types.MsgCollateralize,
) (*types.MsgCollateralizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	if err := s.keeper.Collateralize(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral added",
		"borrower", msg.Borrower,
		"amount", msg.Asset.String(),
	)
	err = ctx.EventManager().EmitTypedEvent(&types.EventCollaterize{
		Borrower: msg.Borrower,
		Utoken:   msg.Asset,
	})
	return &types.MsgCollateralizeResponse{}, err
}

func (s msgServer) SupplyCollateral(
	goCtx context.Context,
	msg *types.MsgSupplyCollateral,
) (*types.MsgSupplyCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}
	uToken, err := s.keeper.Supply(ctx, supplierAddr, msg.Asset)
	if err != nil {
		return nil, err
	}
	if err = s.keeper.Collateralize(ctx, supplierAddr, uToken); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets supplied",
		"supplier", msg.Supplier,
		"supplied", msg.Asset.String(),
		"received", uToken.String(),
	)
	if err = ctx.EventManager().EmitTypedEvent(&types.EventSupply{
		Supplier: msg.Supplier,
		Asset:    msg.Asset,
		Utoken:   uToken,
	}); err != nil {
		return nil, err
	}
	s.keeper.Logger(ctx).Debug(
		"collateral added",
		"borrower", msg.Supplier,
		"amount", uToken.String(),
	)
	err = ctx.EventManager().EmitTypedEvent(&types.EventCollaterize{
		Borrower: msg.Supplier,
		Utoken:   uToken,
	})
	return &types.MsgSupplyCollateralResponse{
		Collateralized: uToken,
	}, err
}

func (s msgServer) Decollateralize(
	goCtx context.Context,
	msg *types.MsgDecollateralize,
) (*types.MsgDecollateralizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	if err := s.keeper.Decollateralize(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral removed",
		"borrower", msg.Borrower,
		"amount", msg.Asset.String(),
	)
	err = ctx.EventManager().EmitTypedEvent(&types.EventDecollaterize{
		Borrower: msg.Borrower,
		Utoken:   msg.Asset,
	})
	return &types.MsgDecollateralizeResponse{}, err
}

func (s msgServer) Borrow(
	goCtx context.Context,
	msg *types.MsgBorrow,
) (*types.MsgBorrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	if err := s.keeper.Borrow(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets borrowed",
		"borrower", msg.Borrower,
		"amount", msg.Asset.String(),
	)
	err = ctx.EventManager().EmitTypedEvent(&types.EventBorrow{
		Borrower: msg.Borrower,
		Asset:    msg.Asset,
	})
	return &types.MsgBorrowResponse{}, err
}

func (s msgServer) Repay(
	goCtx context.Context,
	msg *types.MsgRepay,
) (*types.MsgRepayResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	repaid, err := s.keeper.Repay(ctx, borrowerAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"borrowed assets repaid",
		"borrower", msg.Borrower,
		"attempted", msg.Asset.String(),
		"repaid", repaid.String(),
	)
	err = ctx.EventManager().EmitTypedEvent(&types.EventRepay{
		Borrower: msg.Borrower,
		Repaid:   repaid,
	})
	return &types.MsgRepayResponse{
		Repaid: repaid,
	}, err
}

func (s msgServer) Liquidate(
	goCtx context.Context,
	msg *types.MsgLiquidate,
) (*types.MsgLiquidateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidator, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return nil, err
	}
	borrower, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}
	repaid, liquidated, reward, err := s.keeper.Liquidate(ctx, liquidator, borrower, msg.Repayment, msg.RewardDenom)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"unhealthy borrower liquidated",
		"liquidator", msg.Liquidator,
		"borrower", msg.Borrower,
		"attempted", msg.Repayment.String(),
		"repaid", repaid.String(),
		"liquidated", liquidated.String(),
		"reward", reward.String(),
	)
	err = ctx.EventManager().EmitTypedEvent(&types.EventLiquidate{
		Liquidator: msg.Liquidator,
		Borrower:   msg.Borrower,
		Liquidated: liquidated,
	})
	return &types.MsgLiquidateResponse{
		Repaid:     repaid,
		Collateral: liquidated,
		Reward:     reward,
	}, err
}

// GovUpdateRegistry updates existing tokens with new settings
// or adds the new tokens to registry.
func (s msgServer) GovUpdateRegistry(
	goCtx context.Context,
	msg *types.MsgGovUpdateRegistry,
) (*types.MsgGovUpdateRegistryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// checking req msg authority is the gov module address
	if s.keeper.authority != msg.Authority {
		return &types.MsgGovUpdateRegistryResponse{},
			govtypes.ErrInvalidSigner.Wrapf(
				"invalid authority: expected %s, got %s",
				s.keeper.authority, msg.Authority,
			)
	}

	registeredTokens := s.keeper.GetAllRegisteredTokens(ctx)
	registeredTokenDenoms := make(map[string]bool)

	for _, token := range registeredTokens {
		registeredTokenDenoms[token.BaseDenom] = true
	}

	// update the token settings
	err := s.keeper.SaveOrUpdateTokenSettingsToRegistry(ctx, msg.Authority, msg.UpdateTokens, registeredTokenDenoms, true)
	if err != nil {
		return &types.MsgGovUpdateRegistryResponse{}, err
	}

	// adds  the new token settings
	err = s.keeper.SaveOrUpdateTokenSettingsToRegistry(ctx, msg.Authority, msg.AddTokens, registeredTokenDenoms, false)
	if err != nil {
		return &types.MsgGovUpdateRegistryResponse{}, err
	}

	return &types.MsgGovUpdateRegistryResponse{}, nil
}
