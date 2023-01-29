package model

import (
	"encoding/json"
)

type WorldGovMap map[int]*WorldGov

type WorldGovJSON struct {
	WorldGovData []WorldGov `json:"govs"`
}

type WorldGov struct {
	Value       int    `json:"value"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example"`
	Contraband  string `json:"contraband"`
}

func WorldGovsFromFile(b []byte) (WorldGovMap, error) {

	var data WorldGovJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldGovMap)
	for _, d := range data.WorldGovData {
		dataMap[d.Value] = &WorldGov{Value: d.Value, Type: d.Type, Description: d.Description, Example: d.Example, Contraband: d.Contraband}
	}
	return dataMap, nil
}
