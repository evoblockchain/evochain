#!/usr/bin/env bash

NUM_NODE=4

# tackle size chronic goose deny inquiry gesture fog front sea twin raise
# acid pulse trial pill stumble toilet annual upgrade gold zone void civil
# antique onion adult slot sad dizzy sure among cement demise submit scare
# lazy cause kite fence gravity regret visa fuel tone clerk motor rent
# shellcheck disable=SC2034
HARDCODED_MNEMONIC=true

set -e
set -o errexit
set -a
set -m

set -x # activate debugging

source evo.profile
WRAPPEDTX=false
PRERUN=false
NUM_RPC=0
WHITE_LIST=0b066ca0790f27a6595560b23bf1a1193f100797,\
3813c7011932b18f27f172f0de2347871d27e852,\
6ea83a21a43c30a280a3139f6f23d737104b6975,\
bab6c32fa95f3a54ecb7d32869e32e85a25d2e08,\
testnet-node-ids


while getopts "r:isn:b:p:c:Sxwk:" opt; do
  case $opt in
  i)
    echo "EVOHAIN_INIT"
    EVOHAIN_INIT=1
    ;;
  r)
    echo "NUM_RPC=$OPTARG"
    NUM_RPC=$OPTARG
    ;;
  w)
    echo "WRAPPEDTX=$OPTARG"
    WRAPPEDTX=true
    ;;
  x)
    echo "PRERUN=$OPTARG"
    PRERUN=true
    ;;
  s)
    echo "EVOHAIN_START"
    EVOHAIN_START=1
    ;;
  k)
    echo "LOG_SERVER"
    LOG_SERVER="--log-server $OPTARG"
    ;;
  c)
    echo "Test_CASE"
    Test_CASE="--consensus-testcase $OPTARG"
    ;;
  n)
    echo "NUM_NODE=$OPTARG"
    NUM_NODE=$OPTARG
    ;;
  b)
    echo "BIN_NAME=$OPTARG"
    BIN_NAME=$OPTARG
    ;;
  S)
    STREAM_ENGINE="analysis&mysql&localhost:3306,notify&redis&localhost:6379,kline&pulsar&localhost:6650"
    echo "$STREAM_ENGINE"
    ;;
  p)
    echo "IP=$OPTARG"
    IP=$OPTARG
    ;;
  \?)
    echo "Invalid option: -$OPTARG"
    ;;
  esac
done

echorun() {
  echo "------------------------------------------------------------------------------------------------"
  echo "["$@"]"
  $@
  echo "------------------------------------------------------------------------------------------------"
}

killbyname() {
  NAME=$1
  ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2", "$8}'
  ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2}' | sh
  echo "All <$NAME> killed!"
}

init() {
  killbyname ${BIN_NAME}

  (cd ${EVOHAIN_TOP} && make install VenusHeight=1)

  rm -rf cache

  echo "=================================================="
  echo "===== Generate testnet configurations files...===="
  echorun evochaind testnet --v $1 --r $2 -o cache -l \
    --chain-id ${CHAIN_ID} \
    --starting-ip-address ${IP} \
    --base-port ${BASE_PORT} \
    --keyring-backend test
}
recover() {
  killbyname ${BIN_NAME}
  (cd ${EVOHAIN_TOP} && make install VenusHeight=1)
  rm -rf cache
  cp -rf nodecache cache
}

run() {

  index=$1
  seed_mode=$2
  p2pport=$3
  rpcport=$4
  restport=$5
  p2p_seed_opt=$6
  p2p_seed_arg=$7


  if [ "$(uname -s)" == "Darwin" ]; then
      sed -i "" 's/"enable_call": false/"enable_call": true/' cache/node${index}/evochaind/config/genesis.json
      sed -i "" 's/"enable_create": false/"enable_create": true/' cache/node${index}/evochaind/config/genesis.json
      sed -i "" 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' cache/node${index}/evochaind/config/genesis.json
  else
      sed -i 's/"enable_call": false/"enable_call": true/' cache/node${index}/evochaind/config/genesis.json
      sed -i 's/"enable_create": false/"enable_create": true/' cache/node${index}/evochaind/config/genesis.json
      sed -i 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' cache/node${index}/evochaind/config/genesis.json
  fi

  evochaind add-genesis-account 0xF99a5c14A27F0718eec993B88dBf1c82954946A9 900000000evo --home cache/node${index}/evochaind
  evochaind add-genesis-account 0x9560D2Bb9709DcCD09d41b000445aD4Bab5b96cf 900000000evo --home cache/node${index}/evochaind
  evochaind add-genesis-account 0x3BFB74d34CC459c126FcBBa45583c8953D13da07 900000000evo --home cache/node${index}/evochaind
  evochaind add-genesis-account 0xD53FBb580817fCF688648f0D167B821f9E742D29 900000000evo --home cache/node${index}/evochaind

  LOG_LEVEL=main:info,*:error,consensus:error,state:info

  echorun nohup evochaind start \
    --home cache/node${index}/evochaind \
    --p2p.seed_mode=$seed_mode \
    --p2p.allow_duplicate_ip \
    --dynamic-gp-mode=2 \
    --enable-wtx=${WRAPPEDTX} \
    --mempool.node_key_whitelist ${WHITE_LIST} \
    --p2p.pex=false \
    --p2p.addr_book_strict=false \
    $p2p_seed_opt $p2p_seed_arg \
    --p2p.laddr tcp://${IP}:${p2pport} \
    --rpc.laddr tcp://${IP}:${rpcport} \
    --log_level ${LOG_LEVEL} \
    --chain-id ${CHAIN_ID} \
    --upload-delta=false \
    --enable-gid \
    --consensus.timeout_commit 6000ms \
    --enable-blockpart-ack=false \
    --append-pid=true \
    ${LOG_SERVER} \
    --elapsed DeliverTxs=0,Round=1,CommitRound=1,Produce=1 \
    --rest.laddr tcp://localhost:$restport \
    --enable-preruntx=$PRERUN \
    --consensus-role=v$index \
    --active-view-change=true \
    ${Test_CASE} \
    --keyring-backend test >cache/val${index}.log 2>&1 &

#     --iavl-enable-async-commit \    --consensus-testcase case12.json \
#     --upload-delta \
#     --enable-preruntx \
#     --mempool.node_key_whitelist="nodeKey1,nodeKey2" \
#    --mempool.node_key_whitelist ${WHITE_LIST} \
}

function start() {
  killbyname ${BIN_NAME}
  index=0

  echo "============================================"
  echo "=========== Startup seed node...============"
  ((restport = REST_PORT)) # for evm tx
  run $index true ${seedp2pport} ${seedrpcport} $restport
  seed=$(evochaind tendermint show-node-id --home cache/node${index}/evochaind)

  echo "============================================"
  echo "======== Startup validator nodes...========="
  for ((index = 1; index < ${1}; index++)); do
    ((p2pport = BASE_PORT_PREFIX + index * 100 + P2P_PORT_SUFFIX))
    ((rpcport = BASE_PORT_PREFIX + index * 100 + RPC_PORT_SUFFIX))  # for evochaincli
    ((restport = index * 100 + REST_PORT)) # for evm tx
    run $index false ${p2pport} ${rpcport} $restport --p2p.seeds ${seed}@${IP}:${seedp2pport}
  done
  echo "start node done"
}

if [ -z ${IP} ]; then
  IP="127.0.0.1"
fi

if [ ! -z "${EVOHAIN_INIT}" ]; then
	((NUM_VAL=NUM_NODE-NUM_RPC))
  init ${NUM_VAL} ${NUM_RPC}
fi

if [ ! -z "${EVOHAIN_RECOVER}" ]; then
  recover ${NUM_NODE}
fi

if [ ! -z "${EVOHAIN_START}" ]; then
  start ${NUM_NODE}
fi
