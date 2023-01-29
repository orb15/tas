package main

import (
	"tas/internal/cmd/world"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

func main() {

	//root command and flags
	var rootCmd = &cobra.Command{}
	var LogLevel string
	rootCmd.PersistentFlags().StringVar(&LogLevel, util.LogLevelFlagName, util.LogLevelError, "logging level (debug, info, warn, error or off")

	//world command
	var GenScheme string
	var Longform bool
	world.WorldCmdConfig.PersistentFlags().StringVar(&GenScheme, world.WorldGenSchemeFlagName, "standard", "name of generator scheme (standard, custom)")
	world.WorldCmdConfig.PersistentFlags().BoolVar(&Longform, world.LongformOutputFlagName, false, "set to display detailed world information rather than UWP)")
	rootCmd.AddCommand(world.WorldCmdConfig)

	//world-debug command
	world.WorldDebugCmdConfig.PersistentFlags().StringVar(&GenScheme, world.WorldGenSchemeFlagName, "standard", "name of generator scheme (standard, custom)")
	rootCmd.AddCommand(world.WorldDebugCmdConfig)

	rootCmd.Execute()
}
