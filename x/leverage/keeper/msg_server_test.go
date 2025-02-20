package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tessornetwork/nebula/v3/x/leverage/fixtures"
	"github.com/tessornetwork/nebula/v3/x/leverage/types"
)

func (s *IntegrationTestSuite) TestAddTokensToRegistry() {
	govAccAddr := s.app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()
	registeredNebula := fixtures.Token("unebula", "NEBULA", 6)
	newTokens := fixtures.Token("uabcd", "ABCD", 6)

	testCases := []struct {
		name      string
		req       *types.MsgGovUpdateRegistry
		expectErr bool
		errMsg    string
	}{
		{
			"invalid token data",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				AddTokens: []types.Token{
					fixtures.Token("uosmo", "", 6), // empty denom is invalid
				},
			},
			true,
			"invalid denom",
		},
		{
			"unauthorized authority address",
			&types.MsgGovUpdateRegistry{
				Authority:   s.addrs[0].String(),
				Title:       "test",
				Description: "test",
				AddTokens: []types.Token{
					registeredNebula,
				},
			},
			true,
			"expected gov account as only signer for proposal message",
		},
		{
			"already registered token",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				AddTokens: []types.Token{
					registeredNebula,
				},
			},
			true,
			fmt.Sprintf("token %s is already registered", registeredNebula.BaseDenom),
		},
		{
			"valid authority and valid token for registry",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				AddTokens: []types.Token{
					newTokens,
				},
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgSrvr.GovUpdateRegistry(s.ctx, tc.req)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)
				// no tokens should have been deleted
				tokens := s.app.LeverageKeeper.GetAllRegisteredTokens(s.ctx)
				s.Require().Len(tokens, 4)

				token, err := s.app.LeverageKeeper.GetTokenSettings(s.ctx, "uabcd")
				s.Require().NoError(err)
				s.Require().Equal(token.BaseDenom, "uabcd")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateRegistry() {
	govAccAddr := s.app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()
	modifiedNebula := fixtures.Token("unebula", "NEBULA", 6)
	modifiedNebula.ReserveFactor = sdk.MustNewDecFromStr("0.69")

	testCases := []struct {
		name      string
		req       *types.MsgGovUpdateRegistry
		expectErr bool
		errMsg    string
	}{
		{
			"invalid token data",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				UpdateTokens: []types.Token{
					fixtures.Token("uosmo", "", 6), // empty denom is invalid
				},
			},
			true,
			"invalid denom",
		},
		{
			"unauthorized authority address",
			&types.MsgGovUpdateRegistry{
				Authority:   s.addrs[0].String(),
				Title:       "test",
				Description: "test",
				UpdateTokens: []types.Token{
					fixtures.Token("uosmo", "", 6), // empty denom is invalid
				},
			},
			true,
			"expected gov account as only signer for proposal message",
		},
		{
			"valid authority and valid update token registry",
			&types.MsgGovUpdateRegistry{
				Authority:   govAccAddr,
				Title:       "test",
				Description: "test",
				UpdateTokens: []types.Token{
					modifiedNebula,
				},
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgSrvr.GovUpdateRegistry(s.ctx, tc.req)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)
				// no tokens should have been deleted
				tokens := s.app.LeverageKeeper.GetAllRegisteredTokens(s.ctx)
				s.Require().Len(tokens, 3)

				token, err := s.app.LeverageKeeper.GetTokenSettings(s.ctx, "unebula")
				s.Require().NoError(err)
				s.Require().Equal("0.690000000000000000", token.ReserveFactor.String(),
					"reserve factor is correctly set")

				_, err = s.app.LeverageKeeper.GetTokenSettings(s.ctx, fixtures.AtomDenom)
				s.Require().NoError(err)
			}
		})
	}
}
