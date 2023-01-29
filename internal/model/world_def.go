package model

type WorldFaction struct {
	GovernmentStyle  int `json:"government-style"`
	RelativeStrength int `json:"relative-strength"`
}

type WorldStarportInfo struct {
	Value        int  `json:"value"`
	HasHighport  bool `json:"has-highport"`
	BerthingCost int  `json:"berthing-cost"`
}

type WorldDefinition struct {
	SubsectorLoc string             `json:"subsector-loc"`
	Starport     *WorldStarportInfo `json:"starport"`

	Size          int `json:"size"`
	Atmosphere    int `json:"atmosphere"`
	Hydrographics int `json:"hydrographics"`
	Population    int `json:"population"`
	Government    int `json:"government"`
	LawLevel      int `json:"law-level"`
	TechLevel     int `json:"tech-level"`

	Bases      []string `json:"bases"`
	TradeCodes []string `json:"trade-codes"`
	TravelZone string   `json:"zone"`

	//these are not part of the official planetary code
	Temperature      int             `json:"temperature"`
	HabitabilityZone string          `json:"habitability-zone"`
	Factions         []*WorldFaction `json:"factions"`
	Culture          int             `json:"culture"`
}
