package model

import (
	"encoding/json"
)

type WorldTemperatureMap map[int]*WorldTemperature

type WorldTemperatureJSON struct {
	WorldTemperatureData []WorldTemperature `json:"temps"`
}

type WorldTemperature struct {
	Value              int    `json:"value"`
	Type               string `json:"type"`
	AverageTemperature string `json:"avg-temp"`
	Description        string `json:"description"`
}

func WorldTemperaturesFromFile(b []byte) (WorldTemperatureMap, error) {

	var data WorldTemperatureJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldTemperatureMap)
	for _, d := range data.WorldTemperatureData {
		dataMap[d.Value] = &WorldTemperature{Value: d.Value, Type: d.Type, AverageTemperature: d.AverageTemperature, Description: d.Description}
	}
	return dataMap, nil
}
