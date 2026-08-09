package names

import (
	"github.com/spf13/cobra"
)

const (
	defaultWorldNamesPath = "./data-local/"
	defaultWorldNamesFile = "default-world-names.txt"
)

var NamesCmdConfig = &cobra.Command{

	Use:   "names",
	Short: "a no-op command to serve as a top-level command for naming-related functions",
	Run:   namesCmd,
}

func namesCmd(cmd *cobra.Command, args []string) {
	cmd.Usage()
}
