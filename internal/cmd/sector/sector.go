package sector

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
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
	Short: "builds an 8x10 subsector in detail",
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
	sector, err := buildSubSector(ctx, src, worldNameMgr)
	if err != nil {
		log.Error().Err(err).Msg("Sector creation failed")
		return
	}

	//fetch the argument - this is the (sub)sector name
	sectorName := args[0]
	sector.Name = sectorName

	writeSector(ctx, sector)
}

func buildSubSector(ctx *util.TASContext, worldSourceData *model.WorldSource, nameMgr *worldNameMgr) (*model.Sector, error) {

	log := ctx.Logger()
	dice := ctx.Dice()
	log.Info().Msg("Beginning sector generation...")

	sector := &model.Sector{
		Name:   "unknown",
		Worlds: make([]*model.SectorSystem, 0, 40), //40 is approx number of worlds in a subsector using the standard universe creation algorithm
	}

	//i: vertical cols on hex sector map
	//j: position/'row' in the ith column
	for col := 1; col <= 8; col++ {
		for row := 1; row <= 10; row++ {

			// handle case where odd number columns have 10 rows but
			// even number columns have only 9
			if row == 10 && col%2 == 0 {
				continue
			}

			//per rule on pg 246
			if dice.Roll() < shouldCreateWorldThreshold {
				continue
			}

			def := world.GenerateWorld(ctx)
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

			sw := &model.SectorSystem{
				WorldSummaryData: worldSummary,
				HasGasGiant:      dice.Sum(2) < gasGiantThreshold,
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

		// furthermore, create a sector csv file
		writeSectorCSV(ctx, sector, sector.ToFileName())
	}
}

func writeSectorCSV(ctx *util.TASContext, sector *model.Sector, subtree ...string) {
	log := ctx.Logger()

	//handle optional creation of deeper output dirs
	var dirpath string
	switch len(subtree) {
	case 0:
		dirpath = filepath.Join(".", h.OutputDirectoryName)
	case 1:
		dirpath = filepath.Join(".", h.OutputDirectoryName, subtree[0])
	default:
		err := fmt.Errorf("nested directories deeper than 1 level are not supported")
		log.Error().Err(err).Msg("unable to create requested output file path")
		return
	}

	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("unable to make directory")
		return
	}

	//using this approach prevents a file from being created that will overwrite an existing file
	filename := sector.Name + ".csv"
	filePath := filepath.Join(dirpath, filename)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL, h.EasyAccessFileMode)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("unable to open file")
		return
	}

	// organize the data. Must be a [][]string
	allRows := make([][]string, 0)
	var header = []string{"Loc", "Name", "UWP", "Gas Giant?", "Travel Code", "Trade Codes"}
	allRows = append(allRows, header)
	for _, w := range sector.Worlds {

		// convert non-string data to string
		hasGiant := ""
		if w.HasGasGiant {
			hasGiant = "x"
		}
		travelZone := ""
		if w.WorldSummaryData.TravelZone != "G" {
			travelZone = w.WorldSummaryData.TravelZone
		}
		tradeCodes := strings.Join(w.WorldSummaryData.TradeCodes, " ")

		//create the row and add it
		var d = []string{w.WorldSummaryData.HexLocation, w.WorldSummaryData.Name, w.WorldSummaryData.ToBareUWP(), hasGiant, travelZone, tradeCodes}
		allRows = append(allRows, d)
	}

	// create a CSV writer
	writer := csv.NewWriter(f)
	writer.Flush()

	// write the file and flush it
	if err := writer.WriteAll(allRows); err != nil { // Use WriteAll for a 2D slice
		log.Error().Err(err).Str("filename", filename).Msg("unable to write CSV file")
	}
	writer.Flush()

	// final error check - usually on flush
	if err := writer.Error(); err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("unable to flush CSV file")
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
