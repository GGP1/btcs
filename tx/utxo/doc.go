// Package utxo implements operations related to unspent transaction outputs (UTXO),
// which allow us to check for the balance of every wallet of the network. When we look
// for our balance, we only need to find those that can be unlocked with our key.
//
// "Unspent outputs" means that they haven't been referenced in any input, and thus, still
// belong to the person that received them.
//
// The UTXO set is a cache that is built from all blockchain transactions, so we have to
// iterate over all of them just once to find an unspent output or an address balance.
package utxo
