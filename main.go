package main

import (
	"tas/internal/cmd/trade"
	"tas/internal/cmd/world"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

func main() {

	//root command and flags
	var rootCmd = &cobra.Command{}
	var LogLevel string
	rootCmd.PersistentFlags().StringVar(&LogLevel, util.LogLevelFlagName, util.LogLevelWarn, "logging level (debug, info, warn, error or off")

	//world command
	var GenScheme string
	var Longform bool
	world.WorldCmdConfig.PersistentFlags().StringVar(&GenScheme, world.WorldGenSchemeFlagName, "standard", "name of generator scheme (standard, custom)")
	world.WorldCmdConfig.PersistentFlags().BoolVar(&Longform, world.LongformOutputFlagName, false, "set to display detailed world information rather than UWP)")
	rootCmd.AddCommand(world.WorldCmdConfig)

	//world debug command (world sub command)
	var MaxIterations bool
	world.WorldDebugCmdConfig.PersistentFlags().StringVar(&GenScheme, world.WorldGenSchemeFlagName, "standard", "name of generator scheme (standard, custom)")
	world.WorldDebugCmdConfig.PersistentFlags().BoolVar(&MaxIterations, world.MaxLoopSizeFlagName, false, "set to generate max number of worlds rather than just a rough subsector count)")
	world.WorldCmdConfig.AddCommand(world.WorldDebugCmdConfig)

	//trade command
	var TradeFileName string
	trade.TradeCmdConfig.PersistentFlags().StringVar(&TradeFileName, trade.TradeFileFlagName, "trade-data.json", "name of file in data-local that holds character and world trade facts")
	rootCmd.AddCommand(trade.TradeCmdConfig)

	//speculative trade command (trade sub command)
	trade.TradeCmdConfig.AddCommand(trade.SpecTradeCmdConfig)

	rootCmd.Execute()
}
