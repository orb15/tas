package model

import (
	"encoding/json"
)

type WorldTradeCodeMap map[string]*WorldTradeCode

type WorldTradeCodeJSON struct {
	WorldTradeCodeData []WorldTradeCode `json:"codes"`
}

type WorldTradeCode struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

func WorldTradeCodesFromFile(b []byte) (WorldTradeCodeMap, error) {

	var data WorldTradeCodeJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldTradeCodeMap)
	for _, d := range data.WorldTradeCodeData {
		dataMap[d.Name] = &WorldTradeCode{Name: d.Name, Abbreviation: d.Abbreviation}
	}
	return dataMap, nil
}
