package model

import (
	"encoding/json"
)

type WorldStarportMap map[int]*WorldStarport

type WorldStarportJSON struct {
	WorldStarportData []WorldStarport `json:"starports"`
}

type WorldStarport struct {
	Value      int    `json:"value"`
	Code       string `json:"code"`
	Quality    string `json:"quality"`
	Fuel       string `json:"fuel"`
	Facilities string `json:"facilities"`
}

func WorldStarportsFromFile(b []byte) (WorldStarportMap, error) {

	var data WorldStarportJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldStarportMap)
	for _, d := range data.WorldStarportData {
		dataMap[d.Value] = &WorldStarport{Value: d.Value, Code: d.Code, Quality: d.Quality, Fuel: d.Fuel, Facilities: d.Facilities}
	}
	return dataMap, nil
}
