BIN_NAME=evochaind
#EVOHAIN_TOP=${GOPATH}/src/github.com/evoblockchain/evochain
EVOHAIN_TOP=${HOME}/evochain/
EVOHAIN_BIN=${EVOHAIN_TOP}/build
EVOHAIN_BIN=${GOPATH}/bin
EVOHAIN_NET_TOP=`pwd`
EVOHAIN_NET_CACHE=${EVOHAIN_NET_TOP}/cache
CHAIN_ID="evochain-67"


BASE_PORT_PREFIX=26600
P2P_PORT_SUFFIX=56
RPC_PORT_SUFFIX=57
REST_PORT=8545
let BASE_PORT=${BASE_PORT_PREFIX}+${P2P_PORT_SUFFIX}
let seedp2pport=${BASE_PORT_PREFIX}+${P2P_PORT_SUFFIX}
let seedrpcport=${BASE_PORT_PREFIX}+${RPC_PORT_SUFFIX}
let seedrestport=${seedrpcport}+1
