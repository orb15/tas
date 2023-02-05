package helpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"tas/internal/util"
)

const (
	UnableToContinueBecauseOfErrors = "errors prevent further execution"

	CreditsAbbreviation = "CR"
	NL                  = "\n"
	TAB                 = "\t"
	SP                  = " "

	easyAccessFileMode  = 0755
	outputDirectoryName = "output"
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

func WrappedJSONFileWriter(ctx *util.TASContext, s any, filename string) {

	log := ctx.Logger()
	dirpath := filepath.Join(".", outputDirectoryName)
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("unable to make directory")
		return
	}
	bytes, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("unable to marshal data to JSON")
		return
	}

	//using this approach prevents a file from being created that will overwrite an existing file
	filePath := filepath.Join(dirpath, filename)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL, easyAccessFileMode)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("unable to open file")
		return
	}
	defer (func() {
		err := f.Close()
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("failed to close file")
			return
		}
	})()

	bwrit, err := f.Write(bytes)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("unable to write to file")
		return
	}
	if len(bytes) != bwrit {
		log.Error().Err(err).Int("expected", len(bytes)).Int("wrote", bwrit).Str("filename", filename).Msg("did not write the number of expected bytes to file")
		return
	}

}
