package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopNodesCmd)
}

var createTransactionCmd = &cobra.Command{
	Use:   "create-transaction [network folder] [number of transactions] [output folder name]",
	Short: "Creates algorand payment transactions",
	Args:  cobra.MinimumNArgs(3),
	Run:   createTransactionCmdRun,
}

func createTransactionCmdRun(cmd *cobra.Command, args []string) {

}

/**********************************************/



func getAccountAddressesFromGenesisFile(networkFolder string) []string, error {



	return
}

/**********************************************/

type TransactionBody struct {
	amt  int
	fee  int
	fv   int
	gen  string
	gh   string
	lv   int
	note string
	rcv  string
	snd  string
	typ  string `json:"type"`
}

type AlgorandTransaction struct {
	txn TransactionBody
}

func createTransaction(numberOfTransactions int, accountAddresses []string) []AlgorandTransaction {

}
