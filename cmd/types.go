package cmd

type AlgorandAccount struct {
	Address    string
	PublicKey  []byte
	PrivateKey []byte
}

type AlgodInfo struct {
	EndPointAddress string
	Token           string
}

type AlgorandSignedTransaction struct {
	id string
	tx []byte
}

type TransactionProcessingStats struct {
	//transaction id
	txID string
	//transaction issued
	startTime int64
	//block timestamp
	endTime int64
	//round index
	blockIndex int
}
