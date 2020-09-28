package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"../dbconnector"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/transaction"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createTransactionCmd)
}

var createTransactionCmd = &cobra.Command{
	Use:   "create-transaction [number of transactions] [note size in bytes]",
	Short: "Creates algorand payment transactions with specified note size",
	Args:  cobra.MinimumNArgs(2),
	Run:   createTransactionCmdRun,
}

func createTransactionCmdRun(cmd *cobra.Command, args []string) {

	dbConnector := getDBConnector()
	defer dbConnector.Close()

	accounts := getAlgorandAccountsFromDB(dbConnector)

	algodInfos := getAlgodInfoFromDB(dbConnector)

	if len(algodInfos) == 0 {
		panic("No algod info available!")
	}

	algodAddress := fmt.Sprintf("http://%s", algodInfos[0].EndPointAddress)
	algodToken := algodInfos[0].Token

	fmt.Println("******Using Client******")
	fmt.Println(fmt.Sprintf("\t Endpoint Address:%s", algodAddress))
	fmt.Println(fmt.Sprintf("\t Algod Token     :%s", algodToken))
	fmt.Println("***********************")

	algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		panic(fmt.Errorf("Issue with creating algod client: %s", err))
	}

	numberOfTransactions, err := strconv.Atoi(string(args[0]))
	if err != nil {
		panic(fmt.Errorf("[number of transactions] Could not convert string to int %s", err))
	}

	sizeOfNoteInBytes, err := strconv.Atoi(string(args[1]))
	if err != nil {
		panic(fmt.Errorf("[note size in bytes] Could not convert string to int %s", err))
	}

	signedTransactions := createSignedTransactions(numberOfTransactions, sizeOfNoteInBytes, accounts, algodClient)
	submitTransactions(signedTransactions, algodClient)

}

//Algorand example: https://github.com/algorand/docs/blob/master/examples/start_building/v2/go/yourFirstTransaction.go

func getAlgorandAccountsFromDB(dbConnector dbconnector.DBConnector) []AlgorandAccount {

	accountData, err := dbConnector.GetWithPrefix(ExperimentAccountPrefix)
	if err != nil {
		panic(fmt.Errorf("Error occured during getting accounts from DB: %s", err))
	}

	var accounts []AlgorandAccount

	//fmt.Println("------------Accounts------------")
	for _, accountJSON := range accountData {
		//fmt.Println(string(accountJSON))
		var algorandAccount AlgorandAccount
		json.Unmarshal(accountJSON, &algorandAccount)
		accounts = append(accounts, algorandAccount)
	}

	//fmt.Println(accounts)

	return accounts
}

func getAlgodInfoFromDB(dbConnector dbconnector.DBConnector) []AlgodInfo {

	algodInfoData, err := dbConnector.GetWithPrefix(ExperimentAlgodInfoPrefix)
	if err != nil {
		panic(fmt.Errorf("Error occured during getting algod info from DB: %s", err))
	}

	var algodInfos []AlgodInfo

	//fmt.Println("------------Algod Infos------------")
	for _, algodInfoJSON := range algodInfoData {
		//fmt.Println(string(algodInfoJSON))
		var algodInfo AlgodInfo
		json.Unmarshal(algodInfoJSON, &algodInfo)
		algodInfos = append(algodInfos, algodInfo)
	}

	//fmt.Println(algodInfos)

	return algodInfos
}

func getAlgodClient(algodAddress string, algodToken string) *algod.Client {
	algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		panic(fmt.Errorf("Issue with creating algod client: %s", err))
	}

	return algodClient
}

func createSignedTransactions(numberOfTransactions int, sizeOfNoteInBytes int, accounts []AlgorandAccount, algodClient *algod.Client) []AlgorandSignedTransaction {

	numberOfAccounts := len(accounts)

	s1 := rand.NewSource(time.Now().UnixNano())
	seededRandom := rand.New(s1)

	var signedTransactions []AlgorandSignedTransaction

	for index := 0; index < numberOfTransactions; index++ {
		randomAccountIndexFrom := seededRandom.Intn(numberOfAccounts)
		fromAccount := accounts[randomAccountIndexFrom]
		fmt.Println(fmt.Sprintf("[%d]From account:%s", randomAccountIndexFrom, fromAccount.Address))

		randomAccountIndexTo := seededRandom.Intn(numberOfAccounts)
		toAccount := accounts[randomAccountIndexTo]
		fmt.Println(fmt.Sprintf("[%d]To account:%s", randomAccountIndexTo, toAccount.Address))

		txParams, err := algodClient.SuggestedParams().Do(context.Background())
		if err != nil {
			panic(fmt.Errorf("Error getting suggested tx params: %s", err))
		}

		var amount uint64 = 1
		var minFee uint64 = 1000

		note := make([]byte, sizeOfNoteInBytes)
		_, err = seededRandom.Read(note)
		if err != nil {
			panic(fmt.Errorf("could not read random bytes to construct transaction %s", err))
		}

		genID := txParams.GenesisID
		genHash := txParams.GenesisHash
		firstValidRound := uint64(txParams.FirstRoundValid)
		lastValidRound := uint64(txParams.LastRoundValid)

		txn, err := transaction.MakePaymentTxnWithFlatFee(fromAccount.Address, toAccount.Address, minFee, amount, firstValidRound, lastValidRound, note, "", genID, genHash)
		if err != nil {
			panic(fmt.Errorf("Error creating transaction: %s", err))
		}

		// Sign the transaction
		txID, signedTxn, err := crypto.SignTransaction(fromAccount.PrivateKey, txn)
		if err != nil {
			panic(fmt.Errorf("Failed to sign transaction: %s", err))
		}
		//fmt.Printf("Signed txid: %s\n", txID)

		signedTX := AlgorandSignedTransaction{id: txID, tx: signedTxn}
		signedTransactions = append(signedTransactions, signedTX)
	}

	return signedTransactions
}

func submitTransactions(signedTransactions []AlgorandSignedTransaction, algodClient *algod.Client) {

	for _, signedTX := range signedTransactions {

		fmt.Println("Submitting transaction: ", signedTX.id)

		sendResponse, err := algodClient.SendRawTransaction(signedTX.tx).Do(context.Background())
		if err != nil {
			panic(fmt.Errorf("failed to send transaction: %s", err))
		}
		fmt.Printf("Transaction successfully submitted: %s size: %d bytes\n", sendResponse, len(signedTX.tx))

	}

}
