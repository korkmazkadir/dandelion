package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createTemplateCmd)
}

var createTemplateCmd = &cobra.Command{
	Use:   "create-template [number of nodes]",
	Short: "Creates a network template for a Private Algorand network.",
	Long:  `Creates a network template for a Private Algorand network. Stake is shared equally. The last wallet may have more share than other wallets.`,
	Args:  createTemplateCmdValidateArgs,
	Run:   createTemplateCmdRun,
}

func createTemplateCmdValidateArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("requires number of nodes argument")
	}

	numberOfNodes, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	if numberOfNodes > 0 || numberOfNodes < 1000 {
		return nil
	}

	return fmt.Errorf("invalid number of nodes specified: %s", args[0])
}

func createTemplateCmdRun(cmd *cobra.Command, args []string) {
	numberOfNodes, _ := strconv.Atoi(args[0])
	networkTemplate := createNetworkTemplate(numberOfNodes)
	writeToAfile(numberOfNodes, networkTemplate)
}

/***********************************************************************/

type wallet struct {
	Name   string
	Stake  float64
	Online bool
}

type genesis struct {
	NetworkName string
	Wallets     []wallet
}

type node struct {
	Name    string
	IsRelay bool
	Wallets []wallet
}

type networkTemplate struct {
	Genesis genesis
	Nodes   []node
}

func createNetworkTemplate(numberOfNodes int) networkTemplate {

	individualStake := float64(100) / float64(numberOfNodes)

	roundedIndividualStake := math.Round(individualStake*1000) / 1000

	/* Creates wallets with equal share */
	wallets := make([]wallet, numberOfNodes)
	for i := 0; i < numberOfNodes; i++ {
		wallets[i].Name = fmt.Sprintf("Wallet-%d", i)
		wallets[i].Online = true
		wallets[i].Stake = roundedIndividualStake
	}

	remainingStake := (100 - (roundedIndividualStake * float64(numberOfNodes)))
	roundedRemainingStake := math.Round(remainingStake*1000) / 1000

	/* Make sures that stakes sum up to 100 */
	wallets[numberOfNodes-1].Stake = wallets[numberOfNodes-1].Stake + roundedRemainingStake

	/* Creates nodes all is relay*/
	nodes := make([]node, numberOfNodes)
	for i := 0; i < numberOfNodes; i++ {
		nodes[i].Name = fmt.Sprintf("Node-%d", i)
		nodes[i].IsRelay = true
		nodes[i].Wallets = make([]wallet, 0)
		nodes[i].Wallets = append(nodes[i].Wallets, wallets[i])
	}

	genesisData := genesis{NetworkName: "", Wallets: wallets}

	template := networkTemplate{Genesis: genesisData, Nodes: nodes}

	return template
}

func writeToAfile(numberOfNodes int, template networkTemplate) string {

	templateJSON, err := json.MarshalIndent(template, "", "\t")
	handleErrorWithPanic(err)

	jsonFileName := fmt.Sprintf("network-template-%d.json", numberOfNodes)
	err = ioutil.WriteFile(jsonFileName, templateJSON, 0644)
	handleErrorWithPanic(err)

	return jsonFileName
}
