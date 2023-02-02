package helpers

const (
	UnableToContinueBecauseOfErrors = "errors prevent further execution"

	CreditsAbbreviation = "CR"
	NL                  = "\n"
	TAB                 = "\t"
	SP                  = " "
)

func MaxInt(i int, j int) int {
	if i >= j {
		return i
	}
	return j
}

func MinInt(i int, j int) int {
	if i <= j {
		return i
	}
	return j
}
