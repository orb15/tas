package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiceRoll(t *testing.T) {

	iterations := 100000
	sum := 0
	d := NewDice()
	for i := 0; i < iterations; i++ {
		sum += d.Roll()
	}

	avg := float32(sum) / float32(iterations)

	assert.InDelta(t, avg, 3.5, .01, "base 1D roll is not generating the expected average")
}

func TestDiceRollWithMods(t *testing.T) {

	iterations := 100000
	sum := 0
	d := NewDice()
	for i := 0; i < iterations; i++ {
		sum += d.Roll(2, 2, -1, -1, -2, -8, 8)
	}

	avg := float32(sum) / float32(iterations)

	assert.InDelta(t, avg, 3.5, .01, "base 1D roll (with modifiers) is not generating the expected average")
}

func TestDiceSum(t *testing.T) {

	iterations := 1000000
	sum := 0
	d := NewDice()
	for i := 0; i < iterations; i++ {
		sum += d.Sum(6)
	}

	avg := float32(sum) / float32(iterations)

	assert.InDelta(t, avg, 21, .01, "base 6D roll is not generating the expected average")
}

func TestDiceSumWithMods(t *testing.T) {

	iterations := 1000000
	sum := 0
	d := NewDice()
	for i := 0; i < iterations; i++ {
		sum += d.Sum(6, 2, 2, -1, -1, -2, -8, 8)
	}

	avg := float32(sum) / float32(iterations)

	assert.InDelta(t, avg, 21, .01, "base 6D roll (with modifiers) is not generating the expected average")
}
