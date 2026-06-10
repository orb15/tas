package model

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

	TradeCodes []string `json:"trade-codes"`
	TravelZone string   `json:"zone"`
}
