package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	defaultExpectedFileSize = 250
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

func IngestFiles(folder string, filesToRead []string) map[string]*IngestResult {

	results := make(map[string]*IngestResult)

	for _, f := range filesToRead {

		path := fmt.Sprintf("%s%s", folder, f)

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

func ReadWorldNamesFromFile(fname string) ([]string, error) {

	rawLines := make([]string, 0, defaultExpectedFileSize)

	//open the file
	wnf, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer func() {
		wnf.Close()
	}()

	scanner := bufio.NewScanner(wnf)
	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)
		if len(text) == 0 {
			continue
		}
		rawLines = append(rawLines, text)
	}

	//ensure no errors when reading file
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rawLines, nil
}
