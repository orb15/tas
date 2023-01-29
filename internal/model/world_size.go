package model

import (
	"encoding/json"
)

type WorldSizeMap map[int]*WorldSize

type WorldSizeJSON struct {
	WorldSizeData []WorldSize `json:"sizes"`
}

type WorldSize struct {
	Value    int    `json:"value"`
	Example  string `json:"example"`
	Diameter string `json:"diameter"`
	Gravity  string `json:"gravity"`
}

func WorldSizesFromFile(b []byte) (WorldSizeMap, error) {

	var data WorldSizeJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldSizeMap)
	for _, d := range data.WorldSizeData {
		dataMap[d.Value] = &WorldSize{Value: d.Value, Example: d.Example, Diameter: d.Diameter, Gravity: d.Gravity}
	}
	return dataMap, nil
}
