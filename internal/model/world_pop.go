package model

import (
	"encoding/json"
)

type WorldPopMap map[int]*WorldPop

type WorldPopJSON struct {
	WorldPopData []WorldPop `json:"pops"`
}

type WorldPop struct {
	Value       int    `json:"value"`
	Inhabitants string `json:"inhabitants"`
}

func WorldPopsFromFile(b []byte) (WorldPopMap, error) {

	var data WorldPopJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldPopMap)
	for _, d := range data.WorldPopData {
		dataMap[d.Value] = &WorldPop{Value: d.Value, Inhabitants: d.Inhabitants}
	}
	return dataMap, nil
}
