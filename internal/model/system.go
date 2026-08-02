package model

import (
	"encoding/json"
	"fmt"
)

type SystemOrbital struct {
	IsPrimary         bool        `json:"is-primary"`
	IsGasGiant        bool        `json:"is-gas-giant"`
	IsAsteroid        bool        `json:"is-asteroid"`
	UWP               string      `json:"uwp"`
	OrbitalNumber     int         `json:"orbital-number"`
	OrbitalDistanceAU float32     `json:"orbital-distance-au"`
	Size              int         `json:"size"`
	Moons             []Moon      `json:"moons"`
	Tags              []SystemTag `json:"tags"`
}

type Moon struct {
	Size int         `json:"size"`
	Tags []SystemTag `json:"tags"`
}

type SolarSystem struct {
	Name              string           `json:"name"`
	TotalOrbitalCount int              `json:"total-orbital-count"`
	AsteroidCount     int              `json:"asteroid-count"`
	GasGiantCount     int              `json:"gas-giant-count"`
	RockyCount        int              `json:"rocky-planet-count"`
	Orbitals          []*SystemOrbital `json:"orbitals"`
	HasGasGiants      bool             `json:"has-gas-giants"`
}

func (s *SolarSystem) ToFilename() string {
	return fmt.Sprintf("solar_system_%s.json", s.Name)
}
func (s *SolarSystem) ToImageFilename() string {
	return fmt.Sprintf("solar_system_%s.png", s.Name)
}

type SystemTags struct {
	GasMoonTags      []SystemTag `json:"gas-moon"`
	GasMoonTagsCount int
	HotTags          []SystemTag `json:"hot"`
	HotTagsCount     int
	StarNearTags     []SystemTag `json:"star-near"`
	StarNearTagCount int
	ColdTags         []SystemTag `json:"cold"`
	ColdTagsCount    int
	GeneralTags      []SystemTag `json:"general"`
	GeneralCount     int
}

type SystemTag struct {
	TagName     string `json:"tag"`
	Description string `json:"description"`
}

func SystemTagsFromFile(b []byte) (*SystemTags, error) {
	var data SystemTags

	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	data.GasMoonTagsCount = len(data.GasMoonTags)
	data.ColdTagsCount = len(data.ColdTags)
	data.GeneralCount = len(data.GeneralTags)
	data.HotTagsCount = len(data.HotTags)
	data.StarNearTagCount = len(data.StarNearTags)

	return &data, nil
}
