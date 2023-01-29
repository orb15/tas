package model

import (
	"encoding/json"
)

type WorldFactionsMap map[int]*WorldFactions

type WorldFactionsJSON struct {
	WorldFactionsData []WorldFactions `json:"factions"`
}

type WorldFactions struct {
	Value            int    `json:"value"`
	RelativeStrength string `json:"relative-strength"`
}

func WorldFactionsFromFile(b []byte) (WorldFactionsMap, error) {

	var data WorldFactionsJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldFactionsMap)
	for _, d := range data.WorldFactionsData {
		dataMap[d.Value] = &WorldFactions{Value: d.Value, RelativeStrength: d.RelativeStrength}
	}
	return dataMap, nil
}
