package oracle

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tessornetwork/nebula/v3/x/oracle/keeper"
	"github.com/tessornetwork/nebula/v3/x/oracle/types"
)

// isPeriodLastBlock returns true if we are at the last block of the period
func isPeriodLastBlock(ctx sdk.Context, blocksPerPeriod uint64) bool {
	return (uint64(ctx.BlockHeight())+1)%blocksPerPeriod == 0
}

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper, experimental bool) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	params := k.GetParams(ctx)
	if isPeriodLastBlock(ctx, params.VotePeriod) {
		// Build claim map over all validators in active set
		validatorClaimMap := make(map[string]types.Claim)
		powerReduction := k.StakingKeeper.PowerReduction(ctx)
		for _, v := range k.StakingKeeper.GetBondedValidatorsByPower(ctx) {
			addr := v.GetOperator()
			validatorClaimMap[addr.String()] = types.NewClaim(v.GetConsensusPower(powerReduction), 0, 0, addr)
		}

		var (
			// voteTargets defines the symbol (ticker) denoms that we require votes on
			voteTargets      []string
			voteTargetDenoms []string
		)
		for _, v := range params.AcceptList {
			voteTargets = append(voteTargets, v.SymbolDenom)
			voteTargetDenoms = append(voteTargetDenoms, v.BaseDenom)
		}

		k.ClearExchangeRates(ctx)

		if isPeriodLastBlock(ctx, params.MedianPeriod) && experimental {
			k.ClearMedians(ctx)
			k.ClearMedianDeviations(ctx)
		}

		// NOTE: it filters out inactive or jailed validators
		ballotDenomSlice := k.OrganizeBallotByDenom(ctx, validatorClaimMap)

		// Iterate through ballots and update exchange rates; drop if not enough votes have been achieved.
		for _, ballotDenom := range ballotDenomSlice {
			// Get weighted median of exchange rates
			exchangeRate, err := Tally(ctx, ballotDenom.Ballot, params.RewardBand, validatorClaimMap)
			if err != nil {
				return err
			}

			// Set the exchange rate, emit ABCI event
			if err = k.SetExchangeRateWithEvent(ctx, ballotDenom.Denom, exchangeRate); err != nil {
				return err
			}

			if experimental {
				// Stamp rate every stamp period if asset is set to have historic stats tracked
				if isPeriodLastBlock(ctx, params.StampPeriod) && params.HistoricAcceptList.Contains(ballotDenom.Denom) {
					k.AddHistoricPrice(ctx, ballotDenom.Denom, exchangeRate)
				}

				// Set median price every median period if asset is set to have historic stats tracked
				if isPeriodLastBlock(ctx, params.MedianPeriod) && params.HistoricAcceptList.Contains(ballotDenom.Denom) {
					k.CalcAndSetMedian(ctx, ballotDenom.Denom)
				}
			}
		}

		// update miss counting & slashing
		voteTargetsLen := len(voteTargets)
		claimSlice := types.ClaimMapToSlice(validatorClaimMap)
		for _, claim := range claimSlice {
			// Skip valid voters
			// in MsgAggregateExchangeRateVote we filter tokens from the AcceptList.
			if int(claim.TokensVoted) == voteTargetsLen {
				continue
			}

			// Increase miss counter
			k.SetMissCounter(ctx, claim.Validator, k.GetMissCounter(ctx, claim.Validator)+1)
		}

		// Distribute rewards to ballot winners
		k.RewardBallotWinners(
			ctx,
			int64(params.VotePeriod),
			int64(params.RewardDistributionWindow),
			voteTargetDenoms,
			claimSlice,
		)

		// Clear the ballot
		k.ClearVotes(ctx, params.VotePeriod)
	}

	// Slash oracle providers who missed voting over the threshold and
	// reset miss counters of all validators at the last block of slash window
	if isPeriodLastBlock(ctx, params.SlashWindow) {
		k.SlashAndResetMissCounters(ctx)
	}

	// Prune historic prices every prune period
	if isPeriodLastBlock(ctx, params.PrunePeriod) && experimental {
		pruneBlock := uint64(ctx.BlockHeight()) - params.PrunePeriod
		for _, v := range params.HistoricAcceptList {
			k.DeleteHistoricPrice(ctx, v.String(), pruneBlock)
		}
	}

	return nil
}

// Tally calculates and returns the median. It sets the set of voters to be
// rewarded, i.e. voted within a reasonable spread from the weighted median to
// the store. Note, the ballot is sorted by ExchangeRate.
func Tally(
	ctx sdk.Context,
	ballot types.ExchangeRateBallot,
	rewardBand sdk.Dec,
	validatorClaimMap map[string]types.Claim,
) (sdk.Dec, error) {
	weightedMedian, err := ballot.WeightedMedian()
	if err != nil {
		return sdk.ZeroDec(), err
	}
	standardDeviation, err := ballot.StandardDeviation()
	if err != nil {
		return sdk.ZeroDec(), err
	}

	// rewardSpread is the MAX((weightedMedian * (rewardBand/2)), standardDeviation)
	rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2))
	rewardSpread = sdk.MaxDec(rewardSpread, standardDeviation)

	for _, tallyVote := range ballot {
		// Filter ballot winners. For voters, we filter out the tally vote iff:
		// (weightedMedian - rewardSpread) <= ExchangeRate <= (weightedMedian + rewardSpread)
		if (tallyVote.ExchangeRate.GTE(weightedMedian.Sub(rewardSpread)) &&
			tallyVote.ExchangeRate.LTE(weightedMedian.Add(rewardSpread))) ||
			!tallyVote.ExchangeRate.IsPositive() {

			key := tallyVote.Voter.String()
			claim := validatorClaimMap[key]

			claim.Weight += tallyVote.Power
			claim.TokensVoted++
			validatorClaimMap[key] = claim
		}
	}

	return weightedMedian, nil
}
