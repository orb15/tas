package main

import (
	"tas/internal/cmd/polish"
	"tas/internal/cmd/sector"
	"tas/internal/cmd/world"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

func main() {

	//root command and flags
	var rootCmd = &cobra.Command{}
	var LogLevel string
	var ToFile bool
	rootCmd.PersistentFlags().StringVar(&LogLevel, util.LogLevelFlagName, util.LogLevelWarn, "logging level (debug, info, warn, error or off")
	rootCmd.PersistentFlags().BoolVar(&ToFile, util.ToFileFlagName, false, "set to also write output to an output file")

	//world command
	var Longform bool
	world.WorldCmdConfig.PersistentFlags().BoolVar(&Longform, world.LongformOutputFlagName, false, "set to display detailed world information rather than UWP)")
	rootCmd.AddCommand(world.WorldCmdConfig)

	//world debug command (world sub command)
	var MaxIterations bool
	world.WorldDebugCmdConfig.PersistentFlags().BoolVar(&MaxIterations, world.MaxLoopSizeFlagName, false, "set to generate max number of worlds rather than just a rough subsector count)")
	world.WorldCmdConfig.AddCommand(world.WorldDebugCmdConfig)

	//sector command
	rootCmd.AddCommand(sector.SectorCmdConfig)

	//polish command
	rootCmd.AddCommand(polish.PolishCmdConfig)

	rootCmd.Execute()
}
