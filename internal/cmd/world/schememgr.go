package world

import (
	"tas/internal/model"
	"tas/internal/util"
)

type SchemeType string

const (
	StandardGeneratorScheme SchemeType = "standard"
	CustomGenoratorScheme   SchemeType = "custom"

	sizeFunc          = "size"
	atmosphereFunc    = "atmo"
	temperatureFunc   = "temp"
	hydrographicsFunc = "hydro"
	populationFunc    = "pop"
	governmentFunc    = "gov"
	factionsFunc      = "fact"
	cultureFunc       = "cult"
	lawFunc           = "law"
	starportFunc      = "star"
	techLevelFunc     = "tech"
	highportFunc      = "high"
	basesFunc         = "bases"
	travelFunc        = "trav"
	tradeFunc         = "trade"
)

type generatorFunction func(ctx *util.TASContext, def *model.WorldDefinition)

type generatorScheme map[string]generatorFunction

// the generator scheme decides which functions get called at world generation. By default
// the as-written rules are used, but these can be customized to have other generator
// functions overwrite one or more of the standard functions with a (hopefully) better
// function that generates better results
func generatorSchemeForName(scheme SchemeType) generatorScheme {

	genSchema := make(generatorScheme)

	//establish baseline generators - use the standard functions to do it by-the-book
	genSchema[sizeFunc] = generateSize
	genSchema[atmosphereFunc] = generateAtmosphere
	genSchema[temperatureFunc] = generateTemperature
	genSchema[hydrographicsFunc] = generateHydrographics
	genSchema[populationFunc] = generatePopulation
	genSchema[governmentFunc] = generateGovernment
	genSchema[factionsFunc] = generateFactions
	genSchema[cultureFunc] = generateCulture
	genSchema[lawFunc] = generateLawLevel
	genSchema[starportFunc] = generateStarport
	genSchema[techLevelFunc] = generateTechLevel
	genSchema[highportFunc] = generateHighport
	genSchema[basesFunc] = generateBases
	genSchema[travelFunc] = generateTravelCode
	genSchema[tradeFunc] = generateTradeCodes

	//allow override baseline if desired
	switch scheme {
	case CustomGenoratorScheme:
		genSchema[hydrographicsFunc] = customHydrographics_FixAirlessWaterWorlds
		genSchema[techLevelFunc] = customTechLevel_FixLowTechValues

	}

	return genSchema

}
