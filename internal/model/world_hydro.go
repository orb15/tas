package model

import (
	"encoding/json"
)

type WorldHydroMap map[int]*WorldHydro

type WorldHydroJSON struct {
	WorldHydroData []WorldHydro `json:"hydros"`
}

type WorldHydro struct {
	Value       int    `json:"value"`
	Percentage  string `json:"percentage"`
	Description string `json:"description"`
}

func WorldHydrosFromFile(b []byte) (WorldHydroMap, error) {

	var data WorldHydroJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldHydroMap)
	for _, d := range data.WorldHydroData {
		dataMap[d.Value] = &WorldHydro{Value: d.Value, Description: d.Description, Percentage: d.Percentage}
	}
	return dataMap, nil
}
