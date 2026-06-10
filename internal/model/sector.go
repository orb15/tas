package model

type SectorSystem struct {
	WorldSummaryData *WorldSummary `json:"world"`
	HasGasGiant      bool          `json:"has-gas-giant"`
}

type Sector struct {
	Name   string          `json:"name"`
	Worlds []*SectorSystem `json:"worlds"`
}

func (s *Sector) ToFileName() string {
	return "sector-" + s.Name
}
