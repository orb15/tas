package helpers

import (
	"tas/internal/util"
)

type predicateType int

const (
	LT predicateType = iota
	LE
	EQ
	GE
	GT
	INR
	IS
)

func AdjustZoneDM(currentDM int, adjustDMBy int, valueToTest bool) int {
	if valueToTest {
		currentDM += adjustDMBy
	}
	return currentDM
}

func AdjustStarportDM(currentDM int, adjustDMBy int, valueToTest string, threshold string) int {
	if valueToTest == threshold {
		currentDM += adjustDMBy
	}
	return currentDM
}

func AdjustDM(ctx *util.TASContext, currentDM int, adjustDMBy int, valueToTest int, predicate predicateType, thresholds ...int) int {

	log := ctx.Logger()

	//sanity / syntax - all of these are dev errors, so fatal is appropriate here. They should _never_ occur unless I screw something up!
	requireThresholdCount := func(n int) {
		if len(thresholds) != n {
			log.Fatal().Int("threshold-count", len(thresholds)).Int("required-count", n).Msg("invalid threshold count for predicate")
		}
	}

	requireThresholdMinCount := func(n int) {
		if len(thresholds) < n {
			log.Fatal().Int("threshold-count", len(thresholds)).Int("min-count", n).Msg("invalid min threshold count for predicate")
		}
	}

	requireThresholdOrdered := func() {
		if len(thresholds) > 1 && (thresholds[0] > thresholds[1]) {
			log.Fatal().Int("low-bound", thresholds[0]).Int("up-bound", thresholds[1]).Msg("threshold range is inverted") //bigtime programmer error
		}
	}

	switch predicate {

	case LE:
		requireThresholdCount(1)
		if valueToTest <= thresholds[0] {
			return currentDM + adjustDMBy
		}
	case LT:
		requireThresholdCount(1)
		if valueToTest < thresholds[0] {
			return currentDM + adjustDMBy
		}
	case EQ:
		requireThresholdCount(1)
		if valueToTest == thresholds[0] {
			return currentDM + adjustDMBy
		}
	case GE:
		requireThresholdCount(1)
		if valueToTest >= thresholds[0] {
			return currentDM + adjustDMBy
		}
	case GT:
		requireThresholdCount(1)
		if valueToTest > thresholds[0] {
			return currentDM + adjustDMBy
		}
	case INR:
		requireThresholdCount(2)
		requireThresholdOrdered()
		if valueToTest >= thresholds[0] && valueToTest <= thresholds[1] {
			return currentDM + adjustDMBy
		}
	case IS:
		requireThresholdMinCount(2)
		shouldAdjust := false
		for _, t := range thresholds {
			if valueToTest == t {
				shouldAdjust = true
				break
			}
		}
		if shouldAdjust {
			return currentDM + adjustDMBy
		}
		return currentDM
	}
	return currentDM
}
