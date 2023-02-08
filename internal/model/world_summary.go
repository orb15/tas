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
	us = "_"
)

type WorldSummary struct {
	UWP           string   `json:"uwp"`
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

func (w WorldSummary) ToFileName() string {
	var sb strings.Builder

	sb.WriteString(w.Starport)
	sb.WriteString(w.Size)
	sb.WriteString(w.Atmosphere)
	sb.WriteString(w.Hydrographics)
	sb.WriteString(w.Population)
	sb.WriteString(w.Government)
	sb.WriteString(w.LawLevel)
	sb.WriteString(ds + w.TechLevel)
	sb.WriteString(".json")

	return sb.String()
}

func (w WorldSummary) ToLongFileName() string {
	var sb strings.Builder

	sb.WriteString(w.HexLocation)
	sb.WriteString(sp + w.Name)
	sb.WriteString(sp + w.Starport)
	sb.WriteString(w.Size)
	sb.WriteString(w.Atmosphere)
	sb.WriteString(w.Hydrographics)
	sb.WriteString(w.Population)
	sb.WriteString(w.Government)
	sb.WriteString(w.LawLevel)
	sb.WriteString(ds + w.TechLevel)
	sb.WriteString(".json")

	return sb.String()
}

func (w WorldSummary) MarshalZerologObject(e *zerolog.Event) {
	val := w.ToUWP()
	e.Str("UWP", val)
}

type ExtendedStarportSummary struct {
	Quality      string `json:"quality"`
	Fuel         string `json:"fuel"`
	Facilities   string `json:"facilities"`
	HasHighport  string `json:"has-highport"`
	BerthingCost string `json:"berthing-cost"`
}

type ExtendedSizeSummary struct {
	Diameter string `json:"diameter"`
	Gravity  string `json:"gravity"`
}

type ExetendedAtmosphereSummary struct {
	Composition               string `json:"composition"`
	Pressure                  string `json:"pressure"`
	GearRequired              string `json:"gear-required"`
	TemperatureClassification string `json:"temp-classification"`
	AverageTemperature        string `json:"avg-temperature"`
	TemperatureDescription    string `json:"temp-description"`
	HabitabilityZone          string `json:"habitability-zone"`
}

type ExtendedHydrographicsSummary struct {
	Percentage  string `json:"percent-water"`
	Description string `json:"description"`
}

type ExtendedPopulationSummary struct {
	Inhabitants string `json:"inhabitants"`
}

type ExtendedGovernmentSummary struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example"`
	Contraband  string `json:"contraband"`
}

type ExtendedFactionsSummary struct {
	Government        string                    `json:"government"`
	RelativeStrength  string                    `json:"relative-strength"`
	GovernmentDetails ExtendedGovernmentSummary `json:"gov-details"`
}

type ExtendedCultureSummary struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ExtendedLawSummary struct {
	BannedWeapons string `json:"banned-weapons"`
	BannedArmor   string `json:"banned-armor"`
}

type ExtendedTechLevelSummary struct {
	Catagory    string `json:"catagpory"`
	Description string `json:"description"`
}

type ExtendedBaseSummary struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ExtendedWorldSummary struct {
	StarportDetails      ExtendedStarportSummary      `json:"starport"`
	SizeDetails          ExtendedSizeSummary          `json:"size"`
	AtmosphereDetails    ExetendedAtmosphereSummary   `json:"atmosphere"`
	HydrographicsDetails ExtendedHydrographicsSummary `json:"hydrographics"`
	PopulationDetails    ExtendedPopulationSummary    `json:"population"`
	GovernmentDetails    ExtendedGovernmentSummary    `json:"government"`
	FactionDetails       []ExtendedFactionsSummary    `json:"factions"`
	CulturDetails        ExtendedCultureSummary       `json:"culture"`
	LawDetails           ExtendedLawSummary           `json:"law-level"`
	TechDetails          ExtendedTechLevelSummary     `json:"tech-level"`
	BaseDetails          []ExtendedBaseSummary        `json:"bases"`
	LongDescription      string                       `json:"long-description"`
}
