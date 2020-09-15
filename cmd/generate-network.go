package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateNetworkCmd)
}

var generateNetworkCmd = &cobra.Command{
	Use:   "generate-network [network template file] [output folder name]",
	Short: "Generates a network using the provided network template file.",
	Long:  "Generates a network using the provided network template JSON file.",
	Args:  cobra.MinimumNArgs(2),
	Run:   generateNetworkCmdRun,
}

func generateNetworkCmdRun(cmd *cobra.Command, args []string) {

	templateFileName := args[0]
	outputFolderName := args[1]

	bashExecutable, err := exec.LookPath("bash")
	if err != nil {
		fmt.Println(err)
	}

	createNetworkCmd := &exec.Cmd{
		Path:   bashExecutable,
		Args:   []string{bashExecutable, "./script/create-network.sh", templateFileName, outputFolderName},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	fmt.Println(createNetworkCmd.String())

	if err = createNetworkCmd.Run(); err != nil {
		fmt.Println(err)
	}

}

/***********************************************************************/
