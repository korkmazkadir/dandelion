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
