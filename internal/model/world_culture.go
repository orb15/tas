package model

import (
	"encoding/json"
)

type WorldCultureMap map[int]*WorldCulture

type WorldCultureJSON struct {
	WorldCultureData []WorldCulture `json:"cultures"`
}

type WorldCulture struct {
	Value   int    `json:"value"`
	Type    string `json:"type"`
	Culture string `json:"culture"`
}

func WorldCulturesFromFile(b []byte) (WorldCultureMap, error) {

	var data WorldCultureJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldCultureMap)
	for _, d := range data.WorldCultureData {
		dataMap[d.Value] = &WorldCulture{Value: d.Value, Type: d.Type, Culture: d.Culture}
	}
	return dataMap, nil
}
