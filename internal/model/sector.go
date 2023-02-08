package model

type SectorWorld struct {
	WorldSummaryData *WorldSummary `json:"world"`
	HasGasGiant      bool          `json:"has-gas-giant"`
}

type Sector struct {
	Name   string         `json:"name"`
	Worlds []*SectorWorld `json:"worlds"`
}

func (s *Sector) ToFileName() string {
	return "sector-" + s.Name
}
