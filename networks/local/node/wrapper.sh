#!/usr/bin/env sh

##
## Input parameters
##
ID=${ID:-0}
LOG=${LOG:-evochaind.log}

##
## Run binary with all parameters
##
export EVOCHAINDHOME="/evochaind/node${ID}/evochaind"

if [ -d "$(dirname "${EVOCHAINDHOME}"/"${LOG}")" ]; then
  evochaind --chain-id evochain-1 --home "${EVOCHAINDHOME}" "$@" | tee "${EVOCHAINDHOME}/${LOG}"
else
  evochaind --chain-id evochain-1 --home "${EVOCHAINDHOME}" "$@"
fi

