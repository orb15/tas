package util

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type TASConfig struct {
	Cmd   *cobra.Command
	Flags *pflag.FlagSet
	Args  []string
}

func NewTASConfig() *TASConfig {
	return &TASConfig{
		Args: nil,
	}
}

func (t *TASConfig) WithArgs(args []string) *TASConfig {
	t.Args = args
	return t
}

func (t *TASConfig) WithCmd(cmd *cobra.Command) (*TASConfig, error) {
	t.Cmd = cmd
	t.Flags = cmd.Flags()
	return t, nil
}
