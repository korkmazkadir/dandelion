package cmd

import (
	"encoding/json"
	"fmt"

	"../dbconnector"
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

	dbConnector := getDBConnector()
	defer dbConnector.Close()

	getAlgorandAccountsFromDB(dbConnector)

}

//Algorand example: https://github.com/algorand/docs/blob/master/examples/start_building/v2/go/yourFirstTransaction.go

func getAlgorandAccountsFromDB(dbConnector dbconnector.DBConnector) ([]AlgorandAccount, error) {

	accountData, err := dbConnector.GetWithPrefix(ExperimentAccountPrefix)
	if err != nil {
		panic(fmt.Errorf("Error occured during getting accounts from DB: %s", err))
	}

	var accounts []AlgorandAccount

	for _, accountJSON := range accountData {
		var algorandAccount AlgorandAccount
		json.Unmarshal(accountJSON, &algorandAccount)
		accounts = append(accounts, algorandAccount)
	}

	fmt.Println(accounts)

	return nil, nil
}

/**********************************************/
