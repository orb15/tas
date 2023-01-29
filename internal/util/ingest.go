package util

import (
	"fmt"
	"os"
)

type IngestResult struct {
	Name string
	Err  error
	Data []byte
}

func (i *IngestResult) Ok() bool {
	return i.Err == nil
}

func AllFilesReadOk(data map[string]*IngestResult) bool {
	for _, v := range data {
		if !v.Ok() {
			return false
		}
	}
	return true
}

func IngestFiles(filesToRead []string) map[string]*IngestResult {

	results := make(map[string]*IngestResult)

	for _, f := range filesToRead {

		path := fmt.Sprintf("data/%s", f)

		ir := &IngestResult{
			Name: f,
			Err:  nil,
			Data: nil,
		}

		data, err := os.ReadFile(path)
		if err != nil {
			ir.Err = fmt.Errorf("unable to read file %s. Underlying error: %w", path, err)
		} else {
			ir.Data = data
		}

		results[f] = ir
	}

	return results
}
