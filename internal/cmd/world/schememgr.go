package world

import (
	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"
)

const (
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
func generatorSchemeForName(scheme h.SchemeType) generatorScheme {

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
	case h.UpTechmGeneratorScheme:
		genSchema[hydrographicsFunc] = upTechHydrographics_FixAirlessWaterWorlds
		genSchema[techLevelFunc] = upTechTechLevel_FixLowTechValues
	case h.PristineGeneratorScheme:
		genSchema[populationFunc] = pristinePopulation_noPopulation
		genSchema[governmentFunc] = pristineGovernment_noGovernment
		genSchema[factionsFunc] = pristineFactions_noFactionst
		genSchema[cultureFunc] = pristineCulture_noCulture
		genSchema[lawFunc] = pristineLawLevel_noLaws
		genSchema[starportFunc] = pristineStarport_noStarport
		genSchema[techLevelFunc] = pristineTechLevel_noTechLevel
		genSchema[highportFunc] = pristineHighport_noHighport
		genSchema[tradeFunc] = pristineTradeCode_removeObviousTradeCodes
		genSchema[basesFunc] = pristineBases_noBases
		genSchema[travelFunc] = pristineTravelZone_atmoAmberOnly
	}
	return genSchema
}
