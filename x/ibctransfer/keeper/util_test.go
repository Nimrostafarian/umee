package keeper_test

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	nebulaapp "github.com/tessornetwork/nebula/v3/app"
)

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	encConfig := nebulaapp.MakeEncodingConfig()
	app := nebulaapp.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		nebulaapp.DefaultNodeHome,
		5,
		encConfig,
		nebulaapp.EmptyAppOptions{},
		nebulaapp.GetWasmEnabledProposals(),
		nebulaapp.EmptyWasmOpts,
	)
	genesisState := nebulaapp.NewDefaultGenesisState(encConfig.Codec)

	return app, genesisState
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

func AddressFromString(address string) string {
	return sdk.AccAddress(crypto.AddressHash([]byte(address))).String()
}
