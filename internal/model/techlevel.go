package model

import (
	"fmt"

	"encoding/json"
)

type TechLevelMap map[int]*TechLevel

type TechLevelJSON struct {
	TechLevelData []TechLevel `json:"tech-levels"`
}

type TechLevel struct {
	Value       int    `json:"value"`
	Description string `json:"description"`
	Catagory    string `json:"catagory"`
}

func (tl *TechLevel) FullName() string {

	c := tl.Catagory
	return fmt.Sprintf("TL%d (%s)", tl.Value, c)
}

func (tl *TechLevel) AbbreviatedName() string {
	return fmt.Sprintf("TL%d", tl.Value)
}

func TechLevelsFromFile(b []byte) (TechLevelMap, error) {

	var data TechLevelJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(TechLevelMap)
	for _, d := range data.TechLevelData {
		dataMap[d.Value] = &TechLevel{Value: d.Value, Description: d.Description, Catagory: d.Catagory}
	}
	return dataMap, nil
}
