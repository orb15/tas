package world

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

const (

	//name of flag that holds worldgen info
	WorldGenSchemeFlagName = "worldscheme"
	LongformOutputFlagName = "long"

	maxNumberOfWorldsToGenerate = 1000

	//file-specific constants
	techLevelFile      = "techlevel.json"
	worldAtmoFile      = "world-atmo.json"
	worldBasesFile     = "world-bases.json"
	worldCultureFile   = "world-culture.json"
	worldFactionsFile  = "world-factions.json"
	worldGovFile       = "world-gov.json"
	worldHydroFile     = "world-hydro.json"
	worldLawFile       = "world-law.json"
	worldPopFile       = "world-pop.json"
	worldSizeFile      = "world-size.json"
	worldStarportFile  = "world-starport.json"
	worldTradeCodeFile = "world-trade-codes.json"

	//default world generation data
	defaultWorldName   = "UNK"
	defaultHexLocation = "0000"

	//special one-off temp values to indicate special circumstances on a table
	specialTempCodeForNoAtmo   = -1 //indicates that temperature is boiling/freezing at day/night
	specialCultureCodeForNoPop = 0  //there is no population, so there is no culture

	//table min/max bounds
	sizeMin  = 0
	sizeMax  = 10
	atmoMin  = 0
	atmoMax  = 15
	hydroMin = 0
	hydroMax = 10
	popMin   = 0
	popMax   = 12
	govMin   = 0
	govMax   = 15
	lawMin   = 0
	lawMax   = 9
	starMin  = 2
	starMax  = 11
	techMin  = 0
	techMax  = 15
)

var WorldCmdConfig = &cobra.Command{

	Use:   "world",
	Short: "creates and displays Worlds",
	Run:   worldCmd,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}

		//number of worlds to generate
		if len(args) == 1 {
			_, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("argument must be a unsigned integer. %w", err)
			}
		}

		return nil
	},
}

func worldCmd(cmd *cobra.Command, args []string) {

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

	//build a context to make all data easily available between calls
	loglevel, _ := cfg.Flags.GetString(util.LogLevelFlagName)
	log := util.NewLogger(loglevel)
	ctx := util.NewContext().
		WithLogger(log).
		WithDice().
		WithConfig(cfg)

	//load the data we need to interpret & output a world
	src, err := LoadWorldSourceData(ctx)
	if err != nil {
		return
	}

	//determine number of worlds required. Ignore error here as it was validated by cobra
	numberOfWorldsToGenerate := uint64(1)
	if len(args) > 0 {
		numberOfWorldsToGenerate, _ = strconv.ParseUint(args[0], 10, 16) // <- note 16-bit limit on this param!
		if numberOfWorldsToGenerate == 0 {
			numberOfWorldsToGenerate = 1
		}
		if numberOfWorldsToGenerate > maxNumberOfWorldsToGenerate {
			numberOfWorldsToGenerate = maxNumberOfWorldsToGenerate
		}
	}

	for i := 0; uint64(i) < numberOfWorldsToGenerate; i++ {

		//generate the world
		def := GenerateWorld(ctx)

		//summarize the world in a JSON-ready object
		summary, err := GenerateWorldSummary(ctx, def, src)
		if err != nil {
			log.Error().Err(err).Msg("unable to create world summary")
			return
		}

		//add the long description to the summary
		BuildLongDescription(ctx, summary)

		log.Debug().Object("UWP", summary).Send()
		writeOutput(ctx, summary)
	}

}

func GenerateWorld(ctx *util.TASContext) *model.WorldDefinition {
	def := &model.WorldDefinition{}

	log := ctx.Logger()

	log.Info().Msg("generating world...")

	//generate a world using the raw functions
	rawGenerateSize(ctx, def)
	rawGenerateAtmosphere(ctx, def)
	rawGenerateHydrographics(ctx, def)
	rawGeneratePopulation(ctx, def)
	rawGenerateGovernment(ctx, def)
	rawGenerateLawLevel(ctx, def)
	rawGenerateStarport(ctx, def)
	rawGenerateTechLevel(ctx, def)

	rawGenerateHighport(ctx, def)
	rawGenerateTradeCodes(ctx, def)
	rawGenerateTravelCode(ctx, def)

	log.Info().Msg("world generation complete")
	return def
}

func GenerateWorldSummary(ctx *util.TASContext, def *model.WorldDefinition, src *model.WorldSource) (*model.WorldSummary, error) {

	log := ctx.Logger()

	log.Info().Msg("generating world summary...")

	summary := &model.WorldSummary{
		Name:         defaultWorldName,
		HexLocation:  defaultHexLocation,
		ExtendedData: model.ExtendedWorldSummary{},
	}

	//----------------------------------------
	//core data - forms the UWP
	summary.Starport = src.WorldStarport[def.Starport.Value].Code
	summary.Size = toHex(ctx, def.Size)
	summary.Atmosphere = toHex(ctx, def.Atmosphere)
	summary.Hydrographics = toHex(ctx, def.Hydrographics)
	summary.Population = toHex(ctx, def.Population)
	summary.Government = toHex(ctx, def.Government)
	summary.LawLevel = toHex(ctx, def.LawLevel)
	summary.TechLevel = toHex(ctx, def.TechLevel)

	//trade codes use abbreviation from lookup as all caps
	codes := make([]string, 0, len(def.TradeCodes))
	for _, c := range def.TradeCodes {
		code := strings.ToUpper(src.WorldTradeCodes[c].Abbreviation)
		codes = append(codes, code)
	}
	summary.TradeCodes = codes

	//travel zone is caps of first letter of given travel zone
	summary.TravelZone = strings.ToUpper(def.TravelZone[0:1])

	//----------------------------------------
	//extended data - full info on each element in UWP
	//plus other data not part of the UWP

	// extended starport
	esps := model.ExtendedStarportSummary{
		Quality:      src.WorldStarport[def.Starport.Value].Quality,
		Fuel:         src.WorldStarport[def.Starport.Value].Fuel,
		Facilities:   src.WorldStarport[def.Starport.Value].Facilities,
		HasHighport:  "no",
		BerthingCost: strconv.Itoa(def.Starport.BerthingCost) + h.CreditsAbbreviation,
	}
	if def.Starport.HasHighport {
		esps.HasHighport = "yes"
	}
	summary.ExtendedData.StarportDetails = esps

	//extended size
	ess := model.ExtendedSizeSummary{
		Diameter: src.WorldSize[def.Size].Diameter,
		Gravity:  src.WorldSize[def.Size].Gravity,
	}
	summary.ExtendedData.SizeDetails = ess

	//extended atmosphere
	eas := model.ExetendedAtmosphereSummary{
		Composition:  src.WorldAtmo[def.Atmosphere].Composition,
		Pressure:     src.WorldAtmo[def.Atmosphere].Pressure,
		GearRequired: src.WorldAtmo[def.Atmosphere].GearRequired,
	}
	summary.ExtendedData.AtmosphereDetails = eas

	//extended hydrographics
	ehs := model.ExtendedHydrographicsSummary{
		Percentage:  src.WorldHydro[def.Hydrographics].Percentage,
		Description: src.WorldHydro[def.Hydrographics].Description,
	}
	summary.ExtendedData.HydrographicsDetails = ehs

	//extended population
	eps := model.ExtendedPopulationSummary{
		Inhabitants: src.WorldPop[def.Population].Inhabitants,
	}
	summary.ExtendedData.PopulationDetails = eps

	//extended government
	egs := model.ExtendedGovernmentSummary{
		Type:        src.WorldGov[def.Government].Type,
		Description: src.WorldGov[def.Government].Description,
		Example:     src.WorldGov[def.Government].Example,
		Contraband:  src.WorldGov[def.Government].Contraband,
	}
	summary.ExtendedData.GovernmentDetails = egs

	//extended law level
	els := model.ExtendedLawSummary{
		BannedWeapons: src.WorldLaw[def.LawLevel].BannedWeapons,
		BannedArmor:   src.WorldLaw[def.LawLevel].BannedArmor,
	}
	summary.ExtendedData.LawDetails = els

	//extended tech level
	ets := model.ExtendedTechLevelSummary{
		Catagory:    src.TechLevel[def.TechLevel].Catagory,
		Description: src.TechLevel[def.TechLevel].Description,
	}
	summary.ExtendedData.TechDetails = ets

	log.Info().Msg("world summary complete")

	return summary, nil
}

func LoadWorldSourceData(ctx *util.TASContext) (*model.WorldSource, error) {

	log := ctx.Logger()

	// load source data files
	log.Info().Msg("loading world source files...")

	var sourceFiles = []string{
		techLevelFile,
		worldAtmoFile,
		worldGovFile,
		worldHydroFile,
		worldLawFile,
		worldPopFile,
		worldSizeFile,
		worldStarportFile,
		worldTradeCodeFile}

	fileData := util.IngestFiles("data/", sourceFiles)
	if !util.AllFilesReadOk(fileData) {
		log.Error().Msg("one or more files failed to load as expected")
		for _, f := range fileData {
			if !f.Ok() {
				log.Error().Err(f.Err).Str("filename", f.Name).Send()
			}
		}
		return nil, errors.New(h.UnableToContinueBecauseOfErrors)
	}
	log.Info().Msg("source data files load complete")

	//parse core data
	log.Info().Msg("parsing world source files...")
	source := &model.WorldSource{}

	for filename, fd := range fileData {

		switch filename {

		case techLevelFile:
			tl, err := model.TechLevelsFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.TechLevel = tl

		case worldAtmoFile:
			w, err := model.WorldAtmoFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldAtmo = w

		case worldGovFile:
			w, err := model.WorldGovsFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldGov = w

		case worldHydroFile:
			w, err := model.WorldHydrosFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldHydro = w

		case worldLawFile:
			w, err := model.WorldLawsFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldLaw = w

		case worldPopFile:
			w, err := model.WorldPopsFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldPop = w

		case worldSizeFile:
			w, err := model.WorldSizesFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldSize = w

		case worldStarportFile:
			w, err := model.WorldStarportsFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldStarport = w

		case worldTradeCodeFile:
			w, err := model.WorldTradeCodesFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldTradeCodes = w
		}
	}
	log.Info().Msg("parsing world source files complete")
	return source, nil
}

func BuildLongDescription(ctx *util.TASContext, summary *model.WorldSummary) {
	var sb strings.Builder

	var tzone string
	switch summary.TravelZone {
	case "A":
		tzone = "Amber"
	case "R":
		tzone = "Red"
	default:
		tzone = "Green"
	}

	sb.WriteString("UWP:" + h.SP + summary.ToUWP())
	sb.WriteString(h.NL + "Starport")
	sb.WriteString(h.NL + h.TAB + "Classification:" + h.SP + summary.Starport)
	sb.WriteString(h.NL + h.TAB + "Quality:" + h.SP + summary.ExtendedData.StarportDetails.Quality)
	sb.WriteString(h.NL + h.TAB + "Berthing Cost:" + h.SP + summary.ExtendedData.StarportDetails.BerthingCost)
	sb.WriteString(h.NL + h.TAB + "Fuel Available:" + h.SP + summary.ExtendedData.StarportDetails.Fuel)
	sb.WriteString(h.NL + h.TAB + "Facilities Available:" + h.SP + summary.ExtendedData.StarportDetails.Facilities)
	sb.WriteString(h.NL + h.TAB + "Has Highport:" + h.SP + summary.ExtendedData.StarportDetails.HasHighport)
	sb.WriteString(h.NL + h.TAB + "Travel Zone:" + h.SP + tzone)

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Size:" + h.SP + summary.Size)
	sb.WriteString(h.NL + h.TAB + "Diameter:" + h.SP + summary.ExtendedData.SizeDetails.Diameter)
	sb.WriteString(h.NL + h.TAB + "Gravity:" + h.SP + summary.ExtendedData.SizeDetails.Gravity)

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Atmosphere:" + h.SP + summary.Atmosphere)
	sb.WriteString(h.NL + h.TAB + "Details:" + h.SP + summary.ExtendedData.AtmosphereDetails.Composition)
	sb.WriteString(h.NL + h.TAB + "Pressure Range (PSI):" + h.SP + summary.ExtendedData.AtmosphereDetails.Pressure)
	sb.WriteString(h.NL + h.TAB + "Required Gear:" + h.SP + summary.ExtendedData.AtmosphereDetails.GearRequired)

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Hydrographics:" + h.SP + summary.Hydrographics)
	sb.WriteString(h.NL + h.TAB + "Description:" + h.SP + summary.ExtendedData.HydrographicsDetails.Description)
	sb.WriteString(h.NL + h.TAB + "Hydrographic percentage (liquid, solid and/or gas):" + h.SP + summary.ExtendedData.HydrographicsDetails.Percentage)

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Population:" + h.SP + summary.Population)
	sb.WriteString(h.NL + h.TAB + "Population is in the:" + h.SP + summary.ExtendedData.PopulationDetails.Inhabitants)

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Government:" + h.SP + summary.Government)
	sb.WriteString(h.NL + h.TAB + "Type:" + h.SP + summary.ExtendedData.GovernmentDetails.Type)
	sb.WriteString(h.NL + h.TAB + "Description:" + h.SP + summary.ExtendedData.GovernmentDetails.Description)
	sb.WriteString(h.NL + h.TAB + "Examples of this form of government:" + h.SP + summary.ExtendedData.GovernmentDetails.Example)
	sb.WriteString(h.NL + h.TAB + "Items usually considered contraband by this government:" + h.SP + summary.ExtendedData.GovernmentDetails.Contraband)

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Law Level:" + h.SP + summary.LawLevel)
	sb.WriteString(h.NL + h.TAB + "Banned Weapons:" + h.SP + summary.ExtendedData.LawDetails.BannedWeapons)
	sb.WriteString(h.NL + h.TAB + "Banned Armor:" + h.SP + summary.ExtendedData.LawDetails.BannedArmor)

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Tech Level:" + h.SP + summary.TechLevel)
	sb.WriteString(h.NL + h.TAB + "Classification:" + h.SP + summary.ExtendedData.TechDetails.Catagory)
	sb.WriteString(h.NL + h.TAB + "Description:" + h.SP + summary.ExtendedData.TechDetails.Description)

	if len(summary.TradeCodes) > 0 {
		codes := strings.Join(summary.TradeCodes, ",")
		sb.WriteString(h.NL)
		sb.WriteString(h.NL + "Trade Codes:" + h.SP + codes)
	}

	summary.UWP = summary.ToUWP()
	summary.ExtendedData.LongDescription = sb.String()
}

func writeOutput(ctx *util.TASContext, summary *model.WorldSummary) {

	//get flags
	useLongform, _ := ctx.Config().Flags.GetBool(LongformOutputFlagName)
	writeToFile, _ := ctx.Config().Flags.GetBool(util.ToFileFlagName)

	if useLongform {
		fmt.Println(summary.ExtendedData.LongDescription)
		return
	} else {
		fmt.Println(summary.ToUWP())
	}

	if writeToFile {
		h.WrappedJSONFileWriter(ctx, summary, summary.ToFileName())
	}
}

func toHex(ctx *util.TASContext, i int) string {
	h, err := util.IntAsHexString(i)
	if err != nil {
		h = "error!"
		ctx.Logger().Error().Err(err).Send()
	}
	return h
}
