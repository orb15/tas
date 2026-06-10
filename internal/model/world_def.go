package model

type WorldStarportInfo struct {
	Value        int  `json:"value"`
	HasHighport  bool `json:"has-highport"`
	BerthingCost int  `json:"berthing-cost"`
}

type CustomDetailsInfo struct {
	IsGeothermallyActive       bool    `json:"geothermally-active"`
	HasPlateTectonics          bool    `json:"has-plate-tectonics"`
	DistanceFromSunAU          float32 `json:"dist-from-sun-au"`
	HabitabiltyZoneTempProfile string  `json:"habitability-zone-temp-profile"`
	RotationalPeriodEarthDays  float32 `json:"rotational-period-days"`
	OrbitalPeriodEarthYears    float32 `json:"orbital-period-years"`
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

	CustomDetails CustomDetailsInfo `json:"custom-details"`
}
