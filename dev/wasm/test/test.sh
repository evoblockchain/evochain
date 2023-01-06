#!/bin/bash
set -o errexit -o nounset -o pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

evochaincli keys add fred
echo "0-----------------------"
evochaincli tx send account1 $(evochaincli keys show fred -a) 1000evo --fees 0.001evo -y -b block

echo "1-----------------------"
echo "## Add new CosmWasm contract"
RESP=$(evochaincli tx wasm store "$DIR/../../../x/wasm/keeper/testdata/hackatom.wasm" \
  --from account1 --fees 0.001evo --gas 1500000 -y --node=http://localhost:26657 -b block -o json)

CODE_ID=$(echo "$RESP" | jq -r '.logs[0].events[1].attributes[-1].value')
echo "* Code id: $CODE_ID"
echo "* Download code"
TMPDIR=$(mktemp -t evochaincliXXXXXX)
evochaincli q wasm code "$CODE_ID" "$TMPDIR"
rm -f "$TMPDIR"
echo "-----------------------"
echo "## List code"
evochaincli query wasm list-code --node=http://localhost:26657 -o json | jq

echo "2-----------------------"
echo "## Create new contract instance"
INIT="{\"verifier\":\"$(evochaincli keys show account1 -a)\", \"beneficiary\":\"$(evochaincli keys show fred -a)\"}"
evochaincli tx wasm instantiate "$CODE_ID" "$INIT" --admin="$(evochaincli keys show account1 -a)" \
  --from account1  --fees 0.001evo --amount="100evo" --label "local0.1.0" \
  --gas 1000000 -y -b block -o json | jq

CONTRACT=$(evochaincli query wasm list-contract-by-code "$CODE_ID" -o json | jq -r '.contracts[-1]')
echo "* Contract address: $CONTRACT"
echo "### Query all"
RESP=$(evochaincli query wasm contract-state all "$CONTRACT" -o json)
echo "$RESP" | jq
echo "### Query smart"
evochaincli query wasm contract-state smart "$CONTRACT" '{"verifier":{}}' -o json | jq
echo "### Query raw"
KEY=$(echo "$RESP" | jq -r ".models[0].key")
evochaincli query wasm contract-state raw "$CONTRACT" "$KEY" -o json | jq

echo "3-----------------------"
echo "## Execute contract $CONTRACT"
MSG='{"release":{}}'
evochaincli tx wasm execute "$CONTRACT" "$MSG" \
  --from account1 \
  --gas 1000000 --fees 0.001evo -y  -b block -o json | jq

echo "4-----------------------"
echo "## Set new admin"
echo "### Query old admin: $(evochaincli q wasm contract "$CONTRACT" -o json | jq -r '.contract_info.admin')"
echo "### Update contract"
evochaincli tx wasm set-contract-admin "$CONTRACT" "$(evochaincli keys show fred -a)" \
  --from account1 --fees 0.001evo -y -b block -o json | jq
echo "### Query new admin: $(evochaincli q wasm contract "$CONTRACT" -o json | jq -r '.contract_info.admin')"

echo "5-----------------------"
echo "## Migrate contract"
echo "### Upload new code"
RESP=$(evochaincli tx wasm store "$DIR/../../../x/wasm/keeper/testdata/burner.wasm" \
  --from account1 --fees 0.001evo --gas 1000000 -y --node=http://localhost:26657 -b block -o json)

BURNER_CODE_ID=$(echo "$RESP" | jq -r '.logs[0].events[1].attributes[-1].value')
echo "### Migrate to code id: $BURNER_CODE_ID"

DEST_ACCOUNT=$(evochaincli keys show fred -a)
evochaincli tx wasm migrate "$CONTRACT" "$BURNER_CODE_ID" "{\"payout\": \"$DEST_ACCOUNT\"}" --from fred  --fees 0.001evo \
 -b block -y -o json | jq

echo "### Query destination account: $BURNER_CODE_ID"
evochaincli q account "$DEST_ACCOUNT" -o json | jq
echo "### Query contract meta data: $CONTRACT"
evochaincli q wasm contract "$CONTRACT" -o json | jq

echo "### Query contract meta history: $CONTRACT"
evochaincli q wasm contract-history "$CONTRACT" -o json | jq

echo "6-----------------------"
echo "## Clear contract admin"
echo "### Query old admin: $(evochaincli q wasm contract "$CONTRACT" -o json | jq -r '.contract_info.admin')"
echo "### Update contract"
evochaincli tx wasm clear-contract-admin "$CONTRACT" --fees 0.001evo \
  --from fred -y -b block -o json | jq
echo "### Query new admin: $(evochaincli q wasm contract "$CONTRACT" -o json | jq -r '.contract_info.admin')"
