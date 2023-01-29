package world

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

const (

	//name of flag that holds worldgen info
	WorldGenSchemeFlagName = "scheme"
	LongformOutputFlagName = "long"

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
	worldTempFile      = "world-temp.json"

	//default world generation data
	defaultWorldName    = "UNK"
	defaultHexLocation  = "0000"
	creditsAbbreviation = "CR"
	nl                  = "\n"
	tab                 = "\t"
	sp                  = " "

	//special one-off temp values to indicate special circumstances on a table
	specialTempCodeForNoAtmo   = -1 //indicates that temperature is boiling/freezing at day/night
	specialCultureCodeForNoPop = 0  //there is no populatiobn, so there is no culture

	//table min/max bounds
	sizeMin  = 0
	sizeMax  = 10
	atmoMin  = 0
	atmoMax  = 15
	tempMin  = 2
	tempMax  = 12
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
	Short: "creates a single World",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}

		if len(args) == 1 {
			_, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("argument must be a unsigned integer. %w", err)
			}
		}

		return nil
	},

	Run: worldCmd,
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

	//determine if we want standard (as written) worldgen or want to use the custom generator
	generatorFlagVal, _ := cfg.Flags.GetString(WorldGenSchemeFlagName)
	schemeName := StandardGeneratorScheme
	switch generatorFlagVal {
	case "", "standard":
		generatorFlagVal = "standard" //allows for nice logging below
	case "custom":
		schemeName = CustomGenoratorScheme
	default:
		log.Error().Str("scheme", generatorFlagVal).Msg("invalid generator scheme name")
		return
	}
	log.Info().Str("scheme", generatorFlagVal).Msg("scheme used for world generation")

	//load the data we need to interpret & output a world
	src, err := LoadWorldSourceData(ctx)
	if err != nil {
		return
	}

	//determine number of worlds required. Ignore error here as it was validated by cobra
	numberOfWorldsToGenerate := uint64(1)
	if len(args) > 0 {
		numberOfWorldsToGenerate, _ = strconv.ParseUint(args[0], 10, 16)
		if numberOfWorldsToGenerate == 0 {
			numberOfWorldsToGenerate = 1
		}
		if numberOfWorldsToGenerate > 1000 {
			numberOfWorldsToGenerate = 1000
		}
	}

	for i := 0; uint64(i) < numberOfWorldsToGenerate; i++ {

		//generate the world
		def := GenerateWorld(ctx, schemeName)

		//summarize the world in a JSON-ready object
		summary, err := GenerateWorldSummary(ctx, def, src)
		if err != nil {
			log.Error().Err(err).Msg("unable to create world summary")
			return
		}

		log.Debug().Object("UWP", summary).Send()
		log.Trace().Msg(summary.ExtendedData.LongDescription)

		writeOutput(ctx, summary)
	}

}

func GenerateWorld(ctx *util.TASContext, schemeName SchemeType) *model.WorldDefinition {
	def := &model.WorldDefinition{}

	log := ctx.Logger()

	log.Info().Msg("generating world...")

	genScheme := generatorSchemeForName(schemeName)

	genScheme[sizeFunc](ctx, def)
	genScheme[atmosphereFunc](ctx, def)
	genScheme[temperatureFunc](ctx, def)
	genScheme[hydrographicsFunc](ctx, def)
	genScheme[populationFunc](ctx, def)
	genScheme[governmentFunc](ctx, def)
	genScheme[factionsFunc](ctx, def)
	genScheme[cultureFunc](ctx, def)
	genScheme[lawFunc](ctx, def)
	genScheme[starportFunc](ctx, def)
	genScheme[techLevelFunc](ctx, def)
	genScheme[highportFunc](ctx, def)
	genScheme[basesFunc](ctx, def)
	genScheme[travelFunc](ctx, def)
	genScheme[tradeFunc](ctx, def)

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

	//convert all numeric information in the definition to its world summary counterpart

	//starport
	esd := model.ExtendedStarportSummary{
		Starport:     src.WorldStarport[def.Starport.Value].Code,
		Quality:      src.WorldStarport[def.Starport.Value].Quality,
		Fuel:         src.WorldStarport[def.Starport.Value].Fuel,
		Facilities:   src.WorldStarport[def.Starport.Value].Facilities,
		HasHighport:  "no",
		BerthingCost: strconv.Itoa(def.Starport.BerthingCost) + creditsAbbreviation,
	}
	if def.Starport.HasHighport {
		esd.HasHighport = "yes"
	}
	summary.ExtendedData.StarportSummary = esd

	//core data
	summary.Starport = esd.Starport
	summary.Size = toHex(ctx, def.Size)
	summary.Atmosphere = toHex(ctx, def.Atmosphere)
	summary.Hydrographics = toHex(ctx, def.Hydrographics)
	summary.Population = toHex(ctx, def.Population)
	summary.Government = toHex(ctx, def.Government)
	summary.LawLevel = toHex(ctx, def.LawLevel)
	summary.TechLevel = toHex(ctx, def.TechLevel)

	//bases - use the first letter of the base type, captialized
	bases := make([]string, 0, len(def.Bases))
	for _, b := range def.Bases {
		firstAsCaps := strings.ToUpper(b[0:1])
		bases = append(bases, firstAsCaps)
	}
	summary.Bases = bases

	//trade codes use abbreviation from lookup as all caps
	codes := make([]string, 0, len(def.TradeCodes))
	for _, c := range def.TradeCodes {
		code := strings.ToUpper(src.WorldTradeCodes[c].Abbreviation)
		codes = append(codes, code)
	}
	summary.TradeCodes = codes

	//travel zone is caps of first letter of given travel zone
	summary.TravelZone = strings.ToUpper(def.TravelZone[0:1])

	//extended temperature summary
	ets := model.ExtendedTemperatureSummary{
		Classification:     src.WorldTemperatures[def.Temperature].Type,
		AverageTemperature: src.WorldTemperatures[def.Temperature].AverageTemperature,
		Description:        src.WorldTemperatures[def.Temperature].Description,
		HabitabilityZone:   def.HabitabilityZone,
	}
	summary.ExtendedData.TemperatureSummary = ets

	//extended factions summary
	if len(def.Factions) > 0 {
		factionList := make([]model.ExtendedFactionsSummary, 0, len(def.Factions))
		for _, f := range def.Factions {
			fctn := model.ExtendedFactionsSummary{
				Government:       toHex(ctx, src.WorldGov[f.GovernmentStyle].Value),
				RelativeStrength: src.WorldFactions[f.RelativeStrength].RelativeStrength,
			}
			factionList = append(factionList, fctn)
		}
		summary.ExtendedData.FactionsSummary = factionList
	}

	//culture
	ecs := model.ExtendedCultureSummary{
		Type:        src.WorldCulture[def.Culture].Type,
		Description: src.WorldCulture[def.Culture].Culture,
	}
	summary.ExtendedData.CulturDetail = ecs

	buildLongDescription(ctx, def, src, summary)

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
		worldBasesFile,
		worldCultureFile,
		worldFactionsFile,
		worldGovFile,
		worldHydroFile,
		worldLawFile,
		worldPopFile,
		worldSizeFile,
		worldStarportFile,
		worldTempFile,
		worldTradeCodeFile}

	fileData := util.IngestFiles(sourceFiles)
	if !util.AllFilesReadOk(fileData) {
		log.Error().Msg("one or more files failed to load as expected")
		for _, f := range fileData {
			if !f.Ok() {
				log.Error().Err(f.Err).Str("filename", f.Name).Send()
			}
		}
		return nil, errors.New(helpers.UnableToContinueBecauseOfErrors)
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

		case worldBasesFile:
			w, err := model.WorldBasesFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldBases = w

		case worldCultureFile:
			w, err := model.WorldCulturesFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldCulture = w

		case worldFactionsFile:
			w, err := model.WorldFactionsFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldFactions = w

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

		case worldTempFile:
			w, err := model.WorldTemperaturesFromFile(fd.Data)
			if err != nil {
				return nil, err
			}
			source.WorldTemperatures = w

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

func buildLongDescription(ctx *util.TASContext, def *model.WorldDefinition, src *model.WorldSource, summary *model.WorldSummary) {
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
	sb.WriteString("Starport")
	sb.WriteString(nl + tab + "Classification:" + sp + summary.ExtendedData.StarportSummary.Starport)
	sb.WriteString(nl + tab + "Quality:" + sp + summary.ExtendedData.StarportSummary.Quality)
	sb.WriteString(nl + tab + "Berthing Cost:" + sp + summary.ExtendedData.StarportSummary.BerthingCost)
	sb.WriteString(nl + tab + "Fuel Available:" + sp + summary.ExtendedData.StarportSummary.Fuel)
	sb.WriteString(nl + tab + "Facilities Available:" + sp + summary.ExtendedData.StarportSummary.Facilities)
	sb.WriteString(nl + tab + "Has Highport:" + sp + summary.ExtendedData.StarportSummary.HasHighport)
	sb.WriteString(nl + tab + "Travel Zone:" + sp + tzone)
	if len(summary.Bases) > 0 {
		sb.WriteString(nl + tab + "There are" + sp + toHex(ctx, len(summary.Bases)) + sp + "military bases in addition to the starport")
		for i := 0; i < len(def.Bases); i++ {
			sb.WriteString(nl + tab + tab + "Base" + sp + toHex(ctx, i+1))
			baseType := def.Bases[i]
			sb.WriteString(nl + tab + tab + tab + "Type:" + sp + baseType)
			sb.WriteString(nl + tab + tab + tab + "Description:" + sp + src.WorldBases[baseType].Description)
		}
	}

	sb.WriteString(nl)
	sb.WriteString(nl + "Size:" + sp + summary.Size)
	sb.WriteString(nl + tab + "Diameter:" + sp + src.WorldSize[def.Size].Diameter)
	sb.WriteString(nl + tab + "Gravity:" + sp + src.WorldSize[def.Size].Gravity)

	sb.WriteString(nl)
	sb.WriteString(nl + "Atmosphere:" + sp + summary.Atmosphere)
	sb.WriteString(nl + tab + "Details:" + sp + src.WorldAtmo[def.Atmosphere].Composition)
	sb.WriteString(nl + tab + "Pressure Range (PSI):" + sp + src.WorldAtmo[def.Atmosphere].Pressure)
	sb.WriteString(nl + tab + "Required Gear:" + sp + src.WorldAtmo[def.Atmosphere].GearRequired)
	sb.WriteString(nl + tab + "Temperature Zone:" + sp + summary.ExtendedData.TemperatureSummary.Classification)
	sb.WriteString(nl + tab + "Temperature Descrpition:" + sp + summary.ExtendedData.TemperatureSummary.Description)
	sb.WriteString(nl + tab + "Average Temperature:" + sp + summary.ExtendedData.TemperatureSummary.AverageTemperature)
	sb.WriteString(nl + tab + "Position within star's habitability zone:" + sp + summary.ExtendedData.TemperatureSummary.HabitabilityZone)

	sb.WriteString(nl)
	sb.WriteString(nl + "Hydrographics:" + sp + summary.Hydrographics)
	sb.WriteString(nl + tab + "Description:" + sp + src.WorldHydro[def.Hydrographics].Description)
	sb.WriteString(nl + tab + "Hydrographic percentage (liquid, solid and/or gas):" + sp + src.WorldHydro[def.Hydrographics].Percentage)

	sb.WriteString(nl)
	sb.WriteString(nl + "Population:" + sp + summary.Population)
	sb.WriteString(nl + tab + "Population is in the:" + sp + src.WorldPop[def.Population].Inhabitants)
	sb.WriteString(nl + tab + "Cultural aspect influencing society:" + sp + summary.ExtendedData.CulturDetail.Type)
	sb.WriteString(nl + tab + "How this cultural aspect influences day-to-day or business life:" + sp + summary.ExtendedData.CulturDetail.Description)

	sb.WriteString(nl)
	sb.WriteString(nl + "Government:" + sp + summary.Government)
	sb.WriteString(nl + tab + "Type:" + sp + src.WorldGov[def.Government].Type)
	sb.WriteString(nl + tab + "Description:" + sp + src.WorldGov[def.Government].Description)
	sb.WriteString(nl + tab + "Examples of this form of government:" + sp + src.WorldGov[def.Government].Example)
	sb.WriteString(nl + tab + "Items usually considered contraband by this government:" + sp + src.WorldGov[def.Government].Contraband)
	if len(summary.ExtendedData.FactionsSummary) > 0 {
		sb.WriteString(nl + tab + "There are" + sp + toHex(ctx, len(summary.ExtendedData.FactionsSummary)) + sp + "Factions opposing the Government")

		for i := range summary.ExtendedData.FactionsSummary {
			desiredGovValue := def.Factions[i].GovernmentStyle
			sb.WriteString(nl + tab + tab + "Faction" + sp + toHex(ctx, i+1))
			sb.WriteString(nl + tab + tab + tab + "Desired Government:" + sp + summary.ExtendedData.FactionsSummary[i].Government)
			sb.WriteString(nl + tab + tab + tab + "Type:" + sp + src.WorldGov[desiredGovValue].Type)
			sb.WriteString(nl + tab + tab + tab + "Description:" + sp + src.WorldGov[desiredGovValue].Description)
			sb.WriteString(nl + tab + tab + tab + "Examples of this form of government:" + sp + src.WorldGov[desiredGovValue].Example)
			sb.WriteString(nl + tab + tab + tab + "Items usually considered contraband by this government:" + sp + src.WorldGov[desiredGovValue].Contraband)
			sb.WriteString(nl + tab + tab + tab + "Influence or power relative to primary government:" + sp + summary.ExtendedData.FactionsSummary[i].RelativeStrength)
		}
	}

	sb.WriteString(nl)
	sb.WriteString(nl + "Law Level:" + sp + summary.LawLevel)
	sb.WriteString(nl + tab + "Banned Weapons:" + sp + src.WorldLaw[def.LawLevel].BannedWeapons)
	sb.WriteString(nl + tab + "Banned Armor:" + sp + src.WorldLaw[def.LawLevel].BannedArmor)

	sb.WriteString(nl)
	sb.WriteString(nl + "Tech Level:" + sp + summary.TechLevel)
	sb.WriteString(nl + tab + "Classification:" + sp + src.TechLevel[def.TechLevel].Catagory)
	sb.WriteString(nl + tab + "Description:" + sp + src.TechLevel[def.TechLevel].Description)

	if len(summary.TradeCodes) > 0 {
		codes := strings.Join(def.TradeCodes, ",")
		sb.WriteString(nl)
		sb.WriteString(nl + "Trade Codes:" + sp + codes)
	}

	summary.ExtendedData.LongDescription = sb.String()
}

func writeOutput(ctx *util.TASContext, summary *model.WorldSummary) {

	useLongform, _ := ctx.Config().Flags.GetBool(LongformOutputFlagName)

	if useLongform {
		fmt.Println(summary.ExtendedData.LongDescription)
		return
	}

	fmt.Println(summary.ToUWP())
}

func toHex(ctx *util.TASContext, i int) string {
	h, err := util.IntAsHexString(i)
	if err != nil {
		h = "error!"
		ctx.Logger().Error().Err(err).Send()
	}
	return h
}
