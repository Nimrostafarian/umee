package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tessornetwork/nebula/v3/x/oracle/types"
)

func (s *IntegrationTestSuite) TestBallot_OrganizeBallotByDenom() {
	require := s.Require()
	s.app.OracleKeeper.SetExchangeRate(s.ctx, displayDenom, sdk.OneDec())
	claimMap := make(map[string]types.Claim)

	// Empty Map
	res := s.app.OracleKeeper.OrganizeBallotByDenom(s.ctx, claimMap)
	require.Empty(res)

	s.app.OracleKeeper.SetAggregateExchangeRateVote(
		s.ctx, valAddr, types.AggregateExchangeRateVote{
			ExchangeRateTuples: types.ExchangeRateTuples{
				types.ExchangeRateTuple{
					Denom:        "NEBULA",
					ExchangeRate: sdk.OneDec(),
				},
			},
			Voter: valAddr.String(),
		},
	)

	claimMap[valAddr.String()] = types.Claim{
		Power:       1,
		Weight:      1,
		TokensVoted: 1,
		Validator:   valAddr,
	}
	res = s.app.OracleKeeper.OrganizeBallotByDenom(s.ctx, claimMap)
	require.Equal([]types.BallotDenom{
		{
			Ballot: types.ExchangeRateBallot{types.NewVoteForTally(sdk.OneDec(), "NEBULA", valAddr, 1)},
			Denom:  "NEBULA"},
	}, res)
}

func (s *IntegrationTestSuite) TestBallot_ClearBallots() {
	prevote := types.AggregateExchangeRatePrevote{
		Hash:        "hash",
		Voter:       addr.String(),
		SubmitBlock: 0,
	}
	s.app.OracleKeeper.SetAggregateExchangeRatePrevote(s.ctx, valAddr, prevote)
	prevoteRes, err := s.app.OracleKeeper.GetAggregateExchangeRatePrevote(s.ctx, valAddr)
	s.Require().NoError(err)
	s.Require().Equal(prevoteRes, prevote)

	var tuples types.ExchangeRateTuples
	tuples = append(tuples, types.ExchangeRateTuple{
		Denom:        "NEBULA",
		ExchangeRate: sdk.ZeroDec(),
	})
	vote := types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}
	s.app.OracleKeeper.SetAggregateExchangeRateVote(s.ctx, valAddr, vote)
	voteRes, err := s.app.OracleKeeper.GetAggregateExchangeRateVote(s.ctx, valAddr)
	s.Require().NoError(err)
	s.Require().Equal(voteRes, vote)

	s.app.OracleKeeper.ClearVotes(s.ctx, 0)
	_, err = s.app.OracleKeeper.GetAggregateExchangeRatePrevote(s.ctx, valAddr)
	s.Require().Error(err)
	_, err = s.app.OracleKeeper.GetAggregateExchangeRateVote(s.ctx, valAddr)
	s.Require().Error(err)
}
