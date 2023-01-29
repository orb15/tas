package model

import (
	"strings"

	"github.com/rs/zerolog"
)

/*

	The Universal World Profile is (see pg 248):

	Cogri 0101 CA6A643-9 N RI WA A  where:

	Corgi			-> Name of system
	0101			-> 4-digit number is hex location on subsector map
	C					-> Starport quality
	A					-> Size
	6					-> Atmosphere type
	A					-> Hydrographic percentage
	6					-> Population
	4					-> Government type
	3					-> Law Level
	9					-> Tech Level

	N					-> List of bases present: (N)aval, (M)ilitary, (S)cout and/or (C)orsair
	RI WA			-> Any number of Trade Code abbreviations (e.g. rich and waterworld in this example)
	A					-> A Travel Zone indicator (G)reen, (A)mber or (R)ed
*/

const (
	sp = " "
	ds = "-"
)

type ExtendedStarportSummary struct {
	Starport     string `json:"staport"`
	Quality      string `json:"quality"`
	Fuel         string `json:"fuel"`
	Facilities   string `json:"facilities"`
	HasHighport  string `json:"has-highport"`
	BerthingCost string `json:"berthing-cost"`
}

type ExtendedTemperatureSummary struct {
	Classification     string `json:"classification"`
	AverageTemperature string `json:"avg-temperature"`
	Description        string `json:"description"`
	HabitabilityZone   string `json:"habitability-zone"`
}

type ExtendedFactionsSummary struct {
	Government       string `json:"government"`
	RelativeStrength string `json:"relative-strength"`
}

type ExtendedCultureSummary struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ExtendedWorldSummary struct {
	StarportSummary    ExtendedStarportSummary    `json:"starport"`
	TemperatureSummary ExtendedTemperatureSummary `json:"temperature"`
	FactionsSummary    []ExtendedFactionsSummary  `json:"factions"`
	CulturDetail       ExtendedCultureSummary     `json:"cultural-detail"`
	LongDescription    string                     `json:"long-description"`
}

type WorldSummary struct {
	Name          string   `json:"name"`
	HexLocation   string   `json:"hex-location"`
	Starport      string   `json:"staport"`
	Size          string   `json:"size"`
	Atmosphere    string   `json:"atmosphere"`
	Hydrographics string   `json:"hydrographics"`
	Population    string   `json:"population"`
	Government    string   `json:"government"`
	LawLevel      string   `json:"law-level"`
	TechLevel     string   `json:"tech-level"`
	Bases         []string `json:"bases"`
	TradeCodes    []string `json:"trade-codes"`
	TravelZone    string   `json:"travel-zone"`

	ExtendedData ExtendedWorldSummary `json:"extended-data"`
}

func (w WorldSummary) ToUWP() string {
	var uwp strings.Builder

	uwp.WriteString(w.Name)
	uwp.WriteString(sp + w.HexLocation)
	uwp.WriteString(sp + w.Starport)
	uwp.WriteString(w.Size)
	uwp.WriteString(w.Atmosphere)
	uwp.WriteString(w.Hydrographics)
	uwp.WriteString(w.Population)
	uwp.WriteString(w.Government)
	uwp.WriteString(w.LawLevel)
	uwp.WriteString(ds + w.TechLevel)

	if len(w.Bases) > 0 {
		uwp.WriteString(sp)
		for _, b := range w.Bases {
			uwp.WriteString(b)
		}
	}

	for _, c := range w.TradeCodes {
		uwp.WriteString(sp + c)
	}

	if w.TravelZone != "g" && w.TravelZone != "G" {
		uwp.WriteString(sp + w.TravelZone)
	}

	return uwp.String()
}

func (w WorldSummary) MarshalZerologObject(e *zerolog.Event) {
	val := w.ToUWP()
	e.Str("UWP", val)
}
