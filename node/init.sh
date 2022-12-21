#!/bin/sh

ACCOUNT_NAME="satoshi"
MNEMONIC=$(btcs wallet create | grep "Mnemonic" | tail -c +11)
MINER="false"
DEBUG="false"

# Export the account name for use in other scripts
export $ACCOUNT_NAME

echo -n $MNEMONIC > mnemonic.txt

while getopts ':dm' flag; do
  case "${flag}" in
	d) DEBUG="true" ;;
    m) MINER="true" ;;
  esac
done

btcs wallet createaccount $ACCOUNT_NAME
btcs startnode $ACCOUNT_NAME -m=$MINER --debug=$DEBUG