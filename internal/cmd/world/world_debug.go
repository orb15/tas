package world

import (
	"fmt"
	"strings"

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
	testLoopsToRun = 10000
	maxTestLoops   = 100000 //do NOT exceed this numer!
)

var WorldDebugCmdConfig = &cobra.Command{

	Use:   "world-debug",
	Short: "calcs world generation stats for debugging purposes",
	Run:   debugWorldGeneration,
}

// The approach here is to generate many worlds and see if there is meaningful derivation from the
// above UWP average
func debugWorldGeneration(cmd *cobra.Command, args []string) {

	//prep data store to hold the randomized worlds
	numberOfWorldsToGenerate := testLoopsToRun
	if numberOfWorldsToGenerate > maxTestLoops {
		numberOfWorldsToGenerate = maxTestLoops
	}
	dataStore := make([]*model.WorldDefinition, 0, numberOfWorldsToGenerate)

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

	//generate the planets
	for i := 0; i < numberOfWorldsToGenerate; i++ {
		def := GenerateWorld(ctx, schemeName)
		dataStore = append(dataStore, def)
	}

	//get averages
	sizeAvg := calcAverageForAttribute("Si", dataStore)
	atmoAvg := calcAverageForAttribute("At", dataStore)
	tempAvg := calcAverageForAttribute("Te", dataStore)
	hydroAvg := calcAverageForAttribute("Hy", dataStore)
	popAvg := calcAverageForAttribute("Po", dataStore)
	govAvg := calcAverageForAttribute("Go", dataStore)
	lawAvg := calcAverageForAttribute("Ll", dataStore)
	starAvg := calcAverageForAttribute("Sp", dataStore)
	techAvg := calcAverageForAttribute("Tl", dataStore)

	var sb strings.Builder
	sb.WriteString(nl)
	sb.WriteString(nl)
	sb.WriteString("Stats for" + sp + fmt.Sprintf("%d debug runs", numberOfWorldsToGenerate) + sp + "using scheme:" + sp + generatorFlagVal)
	sb.WriteString(nl)
	sb.WriteString(nl + "Average Size" + tab + tab + tab + fmt.Sprintf("%f", sizeAvg))
	sb.WriteString(nl + "Average Atmosphere" + tab + tab + fmt.Sprintf("%f", atmoAvg))
	sb.WriteString(nl + "Average Temperature" + tab + tab + fmt.Sprintf("%f", tempAvg))
	sb.WriteString(nl + "Average Hydrographics" + tab + tab + fmt.Sprintf("%f", hydroAvg))
	sb.WriteString(nl + "Average Population" + tab + tab + fmt.Sprintf("%f", popAvg))
	sb.WriteString(nl + "Average Government" + tab + tab + fmt.Sprintf("%f", govAvg))
	sb.WriteString(nl + "Average Law Level" + tab + tab + fmt.Sprintf("%f", lawAvg))
	sb.WriteString(nl + "Average Starport" + tab + tab + fmt.Sprintf("%f", starAvg))
	sb.WriteString(nl + "Average Tech Level" + tab + tab + fmt.Sprintf("%f", techAvg))
	sb.WriteString(nl)
	sb.WriteString(nl)

	fmt.Println(sb.String())

}

func calcAverageForAttribute(attrib string, defs []*model.WorldDefinition) float32 {

	sum := 0
	var avg float32

	switch attrib {
	case "Si":
		for _, d := range defs {
			sum += d.Size
		}
	case "At":
		for _, d := range defs {
			sum += d.Atmosphere
		}
	case "Te":
		for _, d := range defs {
			sum += d.Temperature
		}
	case "Hy":
		for _, d := range defs {
			sum += d.Hydrographics
		}
	case "Po":
		for _, d := range defs {
			sum += d.Population
		}
	case "Go":
		for _, d := range defs {
			sum += d.Government
		}
	case "Ll":
		for _, d := range defs {
			sum += d.LawLevel
		}
	case "Sp":
		for _, d := range defs {
			sum += d.Starport.Value
		}
	case "Tl":
		for _, d := range defs {
			sum += d.TechLevel
		}
	}

	avg = float32(sum) / float32(len(defs))

	return avg
}
