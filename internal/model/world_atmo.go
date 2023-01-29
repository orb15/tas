package model

import (
	"encoding/json"
)

type WorldAtmoMap map[int]*WorldAtmo

type WorldAtmoJSON struct {
	WorldAtmoData []WorldAtmo `json:"atmos"`
}

type WorldAtmo struct {
	Value        int    `json:"value"`
	Composition  string `json:"composition"`
	Pressure     string `json:"pressure"`
	GearRequired string `json:"gear-required"`
}

func WorldAtmoFromFile(b []byte) (WorldAtmoMap, error) {

	var data WorldAtmoJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldAtmoMap)
	for _, d := range data.WorldAtmoData {
		dataMap[d.Value] = &WorldAtmo{Value: d.Value, Composition: d.Composition, Pressure: d.Pressure, GearRequired: d.GearRequired}
	}
	return dataMap, nil
}
