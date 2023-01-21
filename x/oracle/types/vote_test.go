package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestAggregateExchangeRatePrevoteString(t *testing.T) {
	addr := sdk.ValAddress(sdk.AccAddress([]byte("addr1_______________")))
	aggregateVoteHash := GetAggregateVoteHash("salt", "NEBULA:100,ATOM:100", addr)
	aggregateExchangeRatePreVote := NewAggregateExchangeRatePrevote(
		aggregateVoteHash,
		addr,
		100,
	)

	require.Equal(t, "hash: 19c30cf9ea8aa0e0b03904162cadec0f2024a76d\nvoter: nebulavaloper1v9jxgu33ta047h6lta047h6lta047h6l5ltnvg\nsubmit_block: 100\n", aggregateExchangeRatePreVote.String())
}

func TestAggregateExchangeRateVoteString(t *testing.T) {
	aggregateExchangeRatePreVote := NewAggregateExchangeRateVote(
		ExchangeRateTuples{
			NewExchangeRateTuple(NebulaDenom, sdk.OneDec()),
		},
		sdk.ValAddress(sdk.AccAddress([]byte("addr1_______________"))),
	)

	require.Equal(t, "exchange_rate_tuples:\n    - denom: unebula\n      exchange_rate: \"1.000000000000000000\"\nvoter: nebulavaloper1v9jxgu33ta047h6lta047h6lta047h6l5ltnvg\n", aggregateExchangeRatePreVote.String())
}

func TestExchangeRateTuplesString(t *testing.T) {
	exchangeRateTuple := NewExchangeRateTuple(NebulaDenom, sdk.OneDec())
	require.Equal(t, exchangeRateTuple.String(), "denom: unebula\nexchange_rate: \"1.000000000000000000\"\n")

	exchangeRateTuples := ExchangeRateTuples{
		exchangeRateTuple,
		NewExchangeRateTuple(IbcDenomAtom, sdk.SmallestDec()),
	}
	require.Equal(t, "- denom: unebula\n  exchange_rate: \"1.000000000000000000\"\n- denom: ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2\n  exchange_rate: \"0.000000000000000001\"\n", exchangeRateTuples.String())
}

func TestParseExchangeRateTuples(t *testing.T) {
	valid := "unebula:123.0,uatom:123.123"
	_, err := ParseExchangeRateTuples(valid)
	require.NoError(t, err)

	duplicatedDenom := "unebula:100.0,uatom:123.123,uatom:121233.123"
	_, err = ParseExchangeRateTuples(duplicatedDenom)
	require.Error(t, err)

	invalidCoins := "123.123"
	_, err = ParseExchangeRateTuples(invalidCoins)
	require.Error(t, err)

	invalidCoinsWithValid := "unebula:123.0,123.1"
	_, err = ParseExchangeRateTuples(invalidCoinsWithValid)
	require.Error(t, err)

	zeroCoinsWithValid := "unebula:0.0,uatom:123.1"
	_, err = ParseExchangeRateTuples(zeroCoinsWithValid)
	require.Error(t, err)

	negativeCoinsWithValid := "unebula:-1234.5,uatom:123.1"
	_, err = ParseExchangeRateTuples(negativeCoinsWithValid)
	require.Error(t, err)

	multiplePricesPerRate := "unebula:123: unebula:456,uusdc:789"
	_, err = ParseExchangeRateTuples(multiplePricesPerRate)
	require.Error(t, err)

	res, err := ParseExchangeRateTuples("")
	require.Nil(t, err)
	require.Nil(t, res)
}
