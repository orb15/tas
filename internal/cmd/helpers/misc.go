package helpers

import (
	"encoding/json"
	"fmt"
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

type SchemeType string

const (
	StandardGeneratorScheme SchemeType = "standard"
	CustomGeneratorScheme   SchemeType = "custom"
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

// s is the struct to be marshalled to JSON, filename is the name of the file to hold that JSON
// subtree is either nil or has 1 element - the name of a folder in the output directory where
// the file should be placed. If nil, the file is placed under the ./output folder. If not nil,
// the file is placed in a subfolder of the output folder.  Sub-subfolders are not permitted
func WrappedJSONFileWriter(ctx *util.TASContext, s any, filename string, subtree ...string) {

	log := ctx.Logger()

	//handle optional creation of deeper output dirs
	var dirpath string
	switch len(subtree) {
	case 0:
		dirpath = filepath.Join(".", outputDirectoryName)
	case 1:
		dirpath = filepath.Join(".", outputDirectoryName, subtree[0])
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

func DetermineWorldGenerationSchemeFromFlagValue(fv string) (string, SchemeType, error) {
	var schemeName string
	var schemeType SchemeType
	switch fv {
	case "", "standard":
		schemeName = "standard" //allows for nice logging below
		schemeType = StandardGeneratorScheme
	case "custom":
		schemeName = "custom"
		schemeType = CustomGeneratorScheme
	default:
		err := fmt.Errorf("world generation scheme: %s is invalid", fv)
		return "", "", err
	}

	return schemeName, schemeType, nil
}
