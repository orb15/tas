package model

import (
	"encoding/json"
)

type WorldLawMap map[int]*WorldLaw

type WorldLawJSON struct {
	WorldLawData []WorldLaw `json:"laws"`
}

type WorldLaw struct {
	Value         int    `json:"value"`
	BannedWeapons string `json:"banned-weapons"`
	BannedArmor   string `json:"banned-armor"`
}

func WorldLawsFromFile(b []byte) (WorldLawMap, error) {

	var data WorldLawJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(WorldLawMap)
	for _, d := range data.WorldLawData {
		dataMap[d.Value] = &WorldLaw{Value: d.Value, BannedWeapons: d.BannedWeapons, BannedArmor: d.BannedArmor}
	}
	return dataMap, nil
}
