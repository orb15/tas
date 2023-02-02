package world

import (
	"fmt"
	"strings"

	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

/*

What is an "average" UWP when doing RAW?

Size is 2D-2 											5		8000km, gravity .45 (bigger than mars, much smaller than earth)
Atmosphere 2D-7 + Size						5		Thin
Temperature	2D + Atmo mods				6		Temperate
Hydrographics 2D-7 + Atmo					5		Roughly 50% water
Population 2D-2										5		100,000's
Government 2D-7 + Pop							5		Feudal Technocracy
Law Level 2D-7 + Gov							5		Military weapons, machine guns and concealables prohibited. Only Jack + laser armor allowed
Starport 2D + Pop mods						7		Class C
Tech Level 1D + mods (round up)		6		Industrial: Fission power, early rocketry (1960's-ish)

The UWP is thus C55655557-6 AG LT NI (ignoring bases)
*/

const (
	MaxLoopSizeFlagName = "max"

	subsectorLoopsToRun = 40
	maxTestLoops        = 10000 //do NOT exceed this numer!
)

var WorldDebugCmdConfig = &cobra.Command{

	Use:   "debug",
	Short: "calcs world generation stats for debugging purposes",
	Run:   debugWorldGeneration,
}

// The approach here is to generate many worlds and see if there is meaningful derivation from the
// above UWP average
func debugWorldGeneration(cmd *cobra.Command, args []string) {

	//create a config to hold all data passed into this call
	cfg, err := util.NewTASConfig().
		WithArgs(args).
		WithCmd(cmd)
	if err != nil {
		fmt.Println()
		fmt.Printf("Unable to create config. This is a critical error: %s\n", err)
		fmt.Println()
		return
	}

	//create context for the calls
	log := util.NewLogger()
	ctx := util.NewContext().
		WithLogger(log).
		WithDice().
		WithConfig(cfg)

	//determine which worldgen scheme to use
	generatorFlagVal, _ := cfg.Flags.GetString(WorldGenSchemeFlagName)
	schemeName := StandardGeneratorScheme
	switch generatorFlagVal {
	case "", "standard":
		generatorFlagVal = "standard" //allows for nice printing below
	case "custom":
		schemeName = CustomGenoratorScheme
	default:
		log.Error().Str("scheme", generatorFlagVal).Msg("invalid generator scheme name")
		return
	}

	//prep data store to hold the randomized worlds
	numberOfWorldsToGenerate := subsectorLoopsToRun
	useMax, _ := cfg.Flags.GetBool(MaxLoopSizeFlagName)
	if err != nil {
		panic(1)
	}
	if useMax {
		numberOfWorldsToGenerate = maxTestLoops
	}

	if numberOfWorldsToGenerate > maxTestLoops {
		numberOfWorldsToGenerate = maxTestLoops
	}
	dataStore := make([]*model.WorldDefinition, 0, numberOfWorldsToGenerate)

	//generate the planets
	for i := 0; i < numberOfWorldsToGenerate; i++ {
		def := GenerateWorld(ctx, schemeName)
		dataStore = append(dataStore, def)
	}

	//get averages
	sizeAvg, sizeMin, sizeMax := calcStatsForAttribute("Si", dataStore)
	atmoAvg, atmoMin, atmoMax := calcStatsForAttribute("At", dataStore)
	tempAvg, tempMin, tempMax := calcStatsForAttribute("Te", dataStore)
	hydroAvg, hydroMin, hydroMax := calcStatsForAttribute("Hy", dataStore)
	popAvg, popMin, popMax := calcStatsForAttribute("Po", dataStore)
	govAvg, govMin, govMax := calcStatsForAttribute("Go", dataStore)
	lawAvg, lawMin, lawMax := calcStatsForAttribute("Ll", dataStore)
	starAvg, starMin, starMax := calcStatsForAttribute("Sp", dataStore)
	techAvg, techMin, techMax := calcStatsForAttribute("Tl", dataStore)

	var sb strings.Builder
	sb.WriteString(h.NL)
	sb.WriteString(h.NL)
	sb.WriteString("Average, Min and Max Stats for" + h.SP + fmt.Sprintf("%d debug runs", numberOfWorldsToGenerate) + h.SP + "using scheme:" + h.SP + generatorFlagVal)
	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Size" + h.TAB + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", sizeAvg, sizeMin, sizeMax))
	sb.WriteString(h.NL + "Atmosphere" + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", atmoAvg, atmoMin, atmoMax))
	sb.WriteString(h.NL + "Temperature" + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", tempAvg, tempMin, tempMax))
	sb.WriteString(h.NL + "Hydrographics" + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", hydroAvg, hydroMin, hydroMax))
	sb.WriteString(h.NL + "Population" + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", popAvg, popMin, popMax))
	sb.WriteString(h.NL + "Government" + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", govAvg, govMin, govMax))
	sb.WriteString(h.NL + "Law Level" + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", lawAvg, lawMin, lawMax))
	sb.WriteString(h.NL + "Starport" + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", starAvg, starMin, starMax))
	sb.WriteString(h.NL + "Tech Level" + h.TAB + h.TAB + fmt.Sprintf("%f\t%d\t%d", techAvg, techMin, techMax))
	sb.WriteString(h.NL)
	sb.WriteString(h.NL)

	fmt.Println(sb.String())

}

func calcStatsForAttribute(attrib string, defs []*model.WorldDefinition) (float32, int, int) {

	sum := 0
	var avg float32
	min := 15
	max := 0

	switch attrib {
	case "Si":
		for _, d := range defs {
			min = h.MinInt(min, d.Size)
			max = h.MaxInt(max, d.Size)
			sum += d.Size
		}
	case "At":
		for _, d := range defs {
			min = h.MinInt(min, d.Atmosphere)
			max = h.MaxInt(max, d.Atmosphere)
			sum += d.Atmosphere
		}
	case "Te":
		for _, d := range defs {
			min = h.MinInt(min, d.Temperature)
			max = h.MaxInt(max, d.Temperature)
			sum += d.Temperature
		}
	case "Hy":
		for _, d := range defs {
			min = h.MinInt(min, d.Hydrographics)
			max = h.MaxInt(max, d.Hydrographics)
			sum += d.Hydrographics
		}
	case "Po":
		for _, d := range defs {
			min = h.MinInt(min, d.Population)
			max = h.MaxInt(max, d.Population)
			sum += d.Population
		}
	case "Go":
		for _, d := range defs {
			min = h.MinInt(min, d.Government)
			max = h.MaxInt(max, d.Government)
			sum += d.Government
		}
	case "Ll":
		for _, d := range defs {
			min = h.MinInt(min, d.LawLevel)
			max = h.MaxInt(max, d.LawLevel)
			sum += d.LawLevel
		}
	case "Sp":
		for _, d := range defs {
			min = h.MinInt(min, d.Starport.Value)
			max = h.MaxInt(max, d.Starport.Value)
			sum += d.Starport.Value
		}
	case "Tl":
		for _, d := range defs {
			min = h.MinInt(min, d.TechLevel)
			max = h.MaxInt(max, d.TechLevel)
			sum += d.TechLevel
		}
	}

	avg = float32(sum) / float32(len(defs))

	return avg, min, max
}
