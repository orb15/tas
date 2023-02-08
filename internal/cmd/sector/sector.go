package sector

import (
	"fmt"
	"strings"

	h "tas/internal/cmd/helpers"
	"tas/internal/cmd/world"

	"tas/internal/model"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

const (
	gasGiantThreshold          = 10
	shouldCreateWorldThreshold = 4
)

var sectorMapIDStrings = map[int]string{1: "01", 2: "02", 3: "03", 4: "04", 5: "05", 6: "06", 7: "07", 8: "08", 9: "09", 10: "10"}

var SectorCmdConfig = &cobra.Command{

	Use:   "sector",
	Short: "determines trade modifiers and other trade-related information",
	Run:   sectorCmd,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactly 1 arguments required - the name of the sector")
		}
		return nil
	},
}

func sectorCmd(cmd *cobra.Command, args []string) {
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
	flagVal, _ := cfg.Flags.GetString(world.WorldGenSchemeFlagName)
	schemeAsString, schemeType, err := h.DetermineWorldGenerationSchemeFromFlagValue(flagVal)
	if err != nil {
		log.Error().Err(err).Msg("invalid flag value for world generation scheme")
		return
	}
	log.Info().Str("scheme", schemeAsString).Msg("scheme used for world generation")

	//load the data we need to interpret & output a world
	src, err := world.LoadWorldSourceData(ctx)
	if err != nil {
		return
	}

	//prepare a world namer
	worldNameMgr, err := newWorldNames(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to prepare world name data")
		return
	}

	//build the subsector
	sector, err := buildSubSector(ctx, schemeType, src, worldNameMgr)
	if err != nil {
		log.Error().Err(err).Msg("Sector creation failed")
		return
	}

	//fetch the argument - thisis the secotr name
	sectorName := args[0]
	sector.Name = sectorName

	writeSector(ctx, sector)
}

func buildSubSector(ctx *util.TASContext, worldGenScheme h.SchemeType, worldSourceData *model.WorldSource, nameMgr *worldNameMgr) (*model.Sector, error) {

	log := ctx.Logger()
	dice := ctx.Dice()
	log.Info().Msg("Beginning sector generation...")

	sector := &model.Sector{
		Name:   "unknown",
		Worlds: make([]*model.SectorWorld, 0, 40), //40 is approx number of worlds in a subsector using the standard universe creation algorithm
	}

	//i: vertical cols on hex sector map
	//j: position/'row' in the ith column
	for col := 1; col <= 8; col++ {
		for row := 1; row <= 10; row++ {

			//per rule on pg 246
			if dice.Roll() < shouldCreateWorldThreshold {
				continue
			}

			def := world.GenerateWorld(ctx, worldGenScheme)
			worldSummary, err := world.GenerateWorldSummary(ctx, def, worldSourceData)
			if err != nil {
				log.Error().Err(err).Msg("unable to generate world")
				return nil, err
			}

			//add some data and recalc UWP then do the summary's long desc
			worldSummary.HexLocation = sectorMapIDStrings[col] + sectorMapIDStrings[row]
			worldSummary.Name = nameMgr.Get()
			worldSummary.UWP = worldSummary.ToUWP()
			world.BuildLongDescription(ctx, worldSummary)

			sw := &model.SectorWorld{
				WorldSummaryData: worldSummary,
				HasGasGiant:      dice.Roll() < gasGiantThreshold,
			}

			sector.Worlds = append(sector.Worlds, sw)
			log.Info().Str("UWP", worldSummary.UWP).Send()
		}
	}

	log.Info().Int("worlds-generated", len(sector.Worlds)).Msg("Sector generation complere")
	return sector, nil

}

func writeSector(ctx *util.TASContext, sector *model.Sector) {

	var sb strings.Builder

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + fmt.Sprintf("Sector: %s (%d worlds)", sector.Name, len(sector.Worlds)))
	sb.WriteString(h.NL + "=====================================")
	for _, w := range sector.Worlds {
		sb.WriteString(h.NL + w.WorldSummaryData.ToUWP())
	}
	fmt.Println(sb.String())

	//also write to file if requested
	writeToFile, _ := ctx.Config().Flags.GetBool(util.ToFileFlagName)
	if writeToFile {
		sectorName := sector.ToFileName()
		for _, w := range sector.Worlds {
			h.WrappedJSONFileWriter(ctx, w, w.WorldSummaryData.ToLongFileName(), sectorName)
		}
	}
}

type worldNameMgr struct {
	availNames map[int]string
	dice       util.Dice
}

func newWorldNames(ctx *util.TASContext) (*worldNameMgr, error) {

	defaultWorldNamesPath := "./data-local/"
	defaultWorldNamesFile := "world-names.txt"
	fname := fmt.Sprintf("%s%s", defaultWorldNamesPath, defaultWorldNamesFile)

	rawNames, err := util.ReadWorldNamesFromFile(fname)
	if err != nil {
		return nil, err
	}

	nameMap := make(map[int]string)
	for i, n := range rawNames {
		nameMap[i] = n
	}

	return &worldNameMgr{
		availNames: nameMap,
		dice:       ctx.Dice(),
	}, nil

}

func (w *worldNameMgr) Get() string {

	cap := len(w.availNames)

	found := false
	var name string
	var ok bool
	for !found {
		pull := w.dice.Dx(cap)
		name, ok = w.availNames[pull]
		if !ok {
			continue
		}
		delete(w.availNames, pull)
		found = true
	}
	return name
}
