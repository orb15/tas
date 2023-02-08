package polish

import (
	"fmt"
	"os"
	"sort"

	"tas/internal/util"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	defaultWorldNamesPath = "./data-local/"
	defaultWorldNamesFile = "world-names.txt"
)

var PolishCmdConfig = &cobra.Command{

	Use:   "polish",
	Short: "cleans up and organizes the ./data-local/world-names.txt file",
	Run:   polishCmd,
}

func polishCmd(cmd *cobra.Command, args []string) {
	//set up logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logCfg := zerolog.ConsoleWriter{Out: os.Stdout}
	log := zerolog.New(logCfg)

	//open the file - not deferring close here b/c we want to open it later, so we will close it explicitly
	log.Info().Msg("Opening world names file...")
	fname := fmt.Sprintf("%s%s", defaultWorldNamesPath, defaultWorldNamesFile)
	rawLines, err := util.ReadWorldNamesFromFile(fname)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to read from world names file")
	}

	//dedupe and sort
	log.Info().Msg("Removing duplicates and sorting...")
	dedupe := make(map[string]struct{})
	for _, n := range rawLines {
		dedupe[n] = struct{}{}
	}
	polished := make([]string, 0, len(dedupe))
	for n := range dedupe {
		polished = append(polished, n)
	}
	sort.Strings(polished)

	//write to a temp file to prevent data loss
	log.Info().Msg("Writing tempfile...")
	tmpf, err := os.CreateTemp(defaultWorldNamesPath, "tmp-names-")
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create temp file")
	}

	first := true
	for _, s := range polished {
		if first {
			_, err = tmpf.WriteString(s)
			first = false
		} else {
			_, err = tmpf.WriteString("\n" + s)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("unable to write to temp file")
		}
	}

	err = tmpf.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to close temp file")
	}

	//rename old file (just being very careful here)
	log.Info().Msg("Cleaning up...")
	oldFileName := fmt.Sprintf("%sold-%s", defaultWorldNamesPath, defaultWorldNamesFile)
	err = os.Rename(fname, oldFileName)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to temporarily rename old world names file")
	}

	//rename the temp file
	err = os.Rename(tmpf.Name(), fname)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to rename temp file")
	}

	//remove the old world names file
	err = os.Remove(oldFileName)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to delete old world names file")
	}

	log.Info().Msg("Done!")
}
