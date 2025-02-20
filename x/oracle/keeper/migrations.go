package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tessornetwork/nebula/v3/x/oracle/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator creates a Migrator.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	m.keeper.SetStampPeriod(ctx, 1)
	m.keeper.SetPrunePeriod(ctx, 1)
	m.keeper.SetMedianPeriod(ctx, 1)
	m.keeper.SetHistoricAcceptList(ctx, types.DenomList{})
	return nil
}
