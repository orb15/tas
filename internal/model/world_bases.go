package model

import (
	"encoding/json"
)

type WorldBaseMap map[string]*WorldBase

type WorldBaseJSON struct {
	WorldBaseData []WorldBase `json:"bases"`
}

type WorldBase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func WorldBasesFromFile(b []byte) (WorldBaseMap, error) {

	var data WorldBaseJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldBaseMap)
	for _, d := range data.WorldBaseData {
		dataMap[d.Name] = &WorldBase{Name: d.Name, Description: d.Description}
	}
	return dataMap, nil
}
