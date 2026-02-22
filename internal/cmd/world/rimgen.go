package world

import (
	"tas/internal/model"
	"tas/internal/util"
)

// This generator generates 'rim world' worlds - worlds that are uninhabited, with no population
// or tech - virgin worlds, basically, ripe for exploration

func rimPopulation_noPopulation(_ *util.TASContext, def *model.WorldDefinition) {
	def.Population = 0
}

func rimGovernment_noGovernment(_ *util.TASContext, def *model.WorldDefinition) {
	def.Government = 0
}
func rimFactions_noFactionst(_ *util.TASContext, def *model.WorldDefinition) {
	def.Factions = nil
}

func rimCulture_noCulture(_ *util.TASContext, def *model.WorldDefinition) {
	def.Culture = specialCultureCodeForNoPop
}

func rimLawLevel_noLaws(_ *util.TASContext, def *model.WorldDefinition) {
	def.LawLevel = 0
}

func rimTechLevel_noTechLevel(_ *util.TASContext, def *model.WorldDefinition) {
	def.TechLevel = 0
}

func rimStarport_noStarport(_ *util.TASContext, def *model.WorldDefinition) {
	def.Starport = &model.WorldStarportInfo{Value: 2}
}

func rimHighport_noHighport(_ *util.TASContext, def *model.WorldDefinition) {
	def.Starport.HasHighport = false
}

func rimBases_noBases(_ *util.TASContext, def *model.WorldDefinition) {
	def.Bases = nil
}

func rimTradeCode_removeObviousTradeCodes(ctx *util.TASContext, def *model.WorldDefinition) {
	// generate the basic trade codes
	generateTradeCodes(ctx, def)

	// filter out low-tech and barren, which always appear
	newCodes := make([]string, 0)
	for _, c := range def.TradeCodes {
		if c == "barren" || c == "low population" {
			continue
		}
		newCodes = append(newCodes, c)
	}
	def.TradeCodes = newCodes
}

func rimTravelZone_atmoAmberOnly(_ *util.TASContext, def *model.WorldDefinition) {
	// only mark a world Amber if the atmosphere warrants it
	if def.Atmosphere >= 10 {
		def.TravelZone = "Amber"
		return
	}
	def.TravelZone = "Green"
}
