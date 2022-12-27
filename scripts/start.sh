#!/bin/bash

KEY="account1"
CHAINID="evochain-11"
MONIKER="evo"
CURDIR=`dirname $0`
HOME_SERVER=$CURDIR/"node_data"

set -e
set -o errexit
set -a
set -m


killbyname() {
  NAME=$1
  ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2", "$8}'
  ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2}' | sh
  echo "All <$NAME> killed!"
}


run() {
    LOG_LEVEL=main:info,iavl:info,*:error,state:info,provider:info

#    evochaind start --pruning=nothing --rpc.unsafe \
    evochaind start --rpc.unsafe \
      --local-rpc-port 26657 \
      --log_level $LOG_LEVEL \
      --log_file json \
      --dynamic-gp-mode=2 \
      --consensus.timeout_commit 2000ms \
      --enable-preruntx=1 \
      --iavl-enable-async-commit \
      --enable-gid \
      --append-pid=true \
      --iavl-output-modules evm=0,acc=0 \
      --commit-gap-height 3 \
      --trace --home $HOME_SERVER --chain-id $CHAINID \
      --elapsed Round=1,CommitRound=1,Produce=1 \
      --rest.laddr "tcp://localhost:8545" > evo.log 2>&1 &

# --iavl-commit-interval-height \
# --iavl-enable-async-commit \
#      --iavl-cache-size int                              Max size of iavl cache (default 1000000)
#      --iavl-commit-interval-height int                  Max interval to commit node cache into leveldb (default 100)
#      --iavl-debug int                                   Enable iavl project debug
#      --iavl-enable-async-commit                         Enable async commit
#      --iavl-enable-pruning-history-state                Enable pruning history state
#      --iavl-height-orphans-cache-size int               Max orphan version to cache in memory (default 8)
#      --iavl-max-committed-height-num int                Max committed version to cache in memory (default 8)
#      --iavl-min-commit-item-count int                   Min nodes num to triggle node cache commit (default 500000)
#      --iavl-output-modules
    exit
}


killbyname evochaind
killbyname evochaincli

set -x # activate debugging

# run

# remove existing daemon and client
rm -rf ~/.evochain*
rm -rf $HOME_SERVER

(cd .. && make install Venus1Height=1 Venus2Height=1 EarthHeight=1)

# Set up config for CLI
evochaincli config chain-id $CHAINID
evochaincli config output json
evochaincli config indent true
evochaincli config trust-node true
evochaincli config keyring-backend test

# if $KEY exists it should be deleted
#

evochaincli keys add --recover account1 -m "cattle dance unaware certain design axis wedding worry another frost treat park" -y

evochaincli keys add --recover account2 -m "also voice sock art correct select yellow super grief ozone spot chuckle" -y

evochaincli keys add --recover account3 -m "shrug master soup parade audit combine end arena tail print fun company" -y

evochaincli keys add --recover account4 -m "need pig cloud recipe hub chicken lizard either raccoon ramp leisure noble" -y

# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
evochaind init $MONIKER --chain-id $CHAINID --home $HOME_SERVER

# Change parameter token denominations to evo
cat $HOME_SERVER/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="evo"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="evo"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="evo"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="evo"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json

# Enable EVM

if [ "$(uname -s)" == "Darwin" ]; then
    sed -i "" 's/"enable_call": false/"enable_call": true/' $HOME_SERVER/config/genesis.json
    sed -i "" 's/"enable_create": false/"enable_create": true/' $HOME_SERVER/config/genesis.json
    sed -i "" 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' $HOME_SERVER/config/genesis.json
else 
    sed -i 's/"enable_call": false/"enable_call": true/' $HOME_SERVER/config/genesis.json
    sed -i 's/"enable_create": false/"enable_create": true/' $HOME_SERVER/config/genesis.json
    sed -i 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' $HOME_SERVER/config/genesis.json
fi

# Allocate genesis accounts (cosmos formatted addresses)
evochaind add-genesis-account $(evochaincli keys show $KEY    -a) 500000000evo --home $HOME_SERVER
evochaind add-genesis-account $(evochaincli keys show account2 -a) 500000000evo --home $HOME_SERVER
evochaind add-genesis-account $(evochaincli keys show account3 -a) 500000000evo --home $HOME_SERVER
evochaind add-genesis-account $(evochaincli keys show account4 -a) 500000000evo --home $HOME_SERVER

# Sign genesis transaction
evochaind gentx --name $KEY --keyring-backend test --home $HOME_SERVER

# Collect genesis tx
evochaind collect-gentxs --home $HOME_SERVER

# Run this to ensure everything worked and that the genesis file is setup correctly
evochaind validate-genesis --home $HOME_SERVER
evochaincli config keyring-backend test

run

# evochaincli tx send account1 0x1136dc175C2B5C18778738fdF6EB209f7A5C4Bda 1evo --fees 1evo -b block -y
