package util

import (
	"math/rand"
	"time"
)

const (
	d6 = 6
	d3 = 3
)

type Dice interface {
	Roll(mods ...int) int
	Sum(n int, mods ...int) int
	D66() int
	D3(mods ...int) int
	Dx(sides int) int
}

type dice struct {
	randgen *rand.Rand
}

func NewDice() Dice {
	return &dice{
		randgen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (d *dice) Roll(mods ...int) int {
	r := d.randgen.Intn(d6) + 1
	for _, m := range mods {
		r += m
	}

	return r
}

func (d *dice) Sum(n int, mods ...int) int {
	sum := 0
	for i := 1; i <= n; i++ {
		sum += d.Roll()
	}
	for _, m := range mods {
		sum += m
	}
	return sum
}

func (d *dice) D66() int {
	tens := d.Roll()
	ones := d.Roll()
	return (tens * 10) + ones
}

func (d *dice) D3(mods ...int) int {
	r := d.randgen.Intn(d3) + 1
	for _, m := range mods {
		r += m
	}

	return r
}

func (d *dice) Dx(sides int) int {
	r := d.randgen.Intn(sides) + 1
	return r
}
