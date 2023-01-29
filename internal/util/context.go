package util

import (
	"context"

	"github.com/rs/zerolog"
)

type keyType string

const (
	keyLogger keyType = "logger"
	keyDice   keyType = "dice"
	keyConfig keyType = "config"
)

type TASContext struct {
	ctx context.Context
}

func NewContext() *TASContext {
	return &TASContext{ctx: context.Background()}
}

func (t *TASContext) WithLogger(l *zerolog.Logger) *TASContext {
	t.ctx = context.WithValue(t.ctx, keyLogger, l)
	return t
}

func (t *TASContext) Logger() *zerolog.Logger {
	return t.ctx.Value(keyLogger).(*zerolog.Logger)
}

func (t *TASContext) WithDice() *TASContext {
	t.ctx = context.WithValue(t.ctx, keyDice, NewDice())
	return t
}

func (t *TASContext) Dice() Dice {
	return t.ctx.Value(keyDice).(Dice)
}

func (t *TASContext) WithConfig(cfg *TASConfig) *TASContext {
	t.ctx = context.WithValue(t.ctx, keyConfig, cfg)
	return t
}

func (t *TASContext) Config() *TASConfig {
	return t.ctx.Value(keyConfig).(*TASConfig)
}
