#!/bin/bash -eu

# tinker the mainnet genesis by swapping polychain -> priv_validator_key

# USAGE: ./tinker-mainnet-genesis.sh

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

NEBULAMAINNET_GENESIS_PATH="${NEBULAMAINNET_GENESIS_PATH:-$CWD/mainnet_tinkered_genesis.json}"
MAINNET_EXPORTED_GENESIS_URL="${MAINNET_EXPORTED_GENESIS_URL:-"https://storage.googleapis.com/nebudropzone/mainnet-3269830-genesis.json.gz"}"
EXPORTED_GENESIS_UNPROCESSED="${EXPORTED_GENESIS_UNPROCESSED:-$CWD/nebulamainnet.genesis.json}"
EXPORTED_GENESIS_UNZIPED="${EXPORTED_GENESIS_UNZIPED:-$CWD/nebulamainnet.genesis.json.gz}"

# Checks for the tinkered genesis file
if [ ! -f "$NEBULAMAINNET_GENESIS_PATH" ]; then
  echo "$NEBULAMAINNET_GENESIS_PATH doesn't exist"

  if [ ! -f "$EXPORTED_GENESIS_UNPROCESSED" ]; then

    if [ ! -f $EXPORTED_GENESIS_UNZIPED ]; then
      echo "$EXPORTED_GENESIS_UNZIPED doesn't exist, we need to curl it"
      curl $MAINNET_EXPORTED_GENESIS_URL > $EXPORTED_GENESIS_UNZIPED
    fi

    echo "$EXPORTED_GENESIS_UNPROCESSED doesn't exist, we need to unpack"
    gunzip -k $EXPORTED_GENESIS_UNZIPED
  fi

  echo "$EXPORTED_GENESIS_UNPROCESSED exists and ready to be processed"
  EXPORTED_GENESIS_UNPROCESSED_PATH=$EXPORTED_GENESIS_UNPROCESSED COSMOS_GENESIS_TINKERER_SCRIPT=nebulamainnet-fork.py EXPORTED_GENESIS_PROCESSED_PATH=$NEBULAMAINNET_GENESIS_PATH $CWD/process_genesis.sh
fi
