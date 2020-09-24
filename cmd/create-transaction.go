package cmd

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createTransactionCmd)
}

var createTransactionCmd = &cobra.Command{
	Use:   "create-transaction [network folder] [number of transactions] [output folder name]",
	Short: "Creates algorand payment transactions",
	//Args:  cobra.MinimumNArgs(3),
	Run: createTransactionCmdRun,
}

func createTransactionCmdRun(cmd *cobra.Command, args []string) {

	createAlgorandAccount()
}

/**********************************************/

func createAlgorandAccount() {

	fmt.Println("------------Creating algorand accounts-----")
	account := crypto.GenerateAccount()

	fmt.Println("Private Key: ", account.PrivateKey)
	fmt.Println("Publick Key: ", account.PublicKey)
	fmt.Println("Account Add: ", account.Address)

}

/*
func getAccountAddressesFromGenesisFile(networkFolder string) []string, error {



	return nil, nil
}
*/

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

/*
func createTransaction(numberOfTransactions int, accountAddresses []string) []AlgorandTransaction {

}
*/
