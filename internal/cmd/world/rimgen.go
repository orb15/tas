package world

import (
	"tas/internal/model"
	"tas/internal/util"
)

// This generator generates 'rim' worlds. Low population, low law worlds found on the
// 'wild west frontier' of inhabited space. This is the Firefly-esque sector generator

func rimPopulation_lowPopulation(ctx *util.TASContext, def *model.WorldDefinition) {
	generatePopulation(ctx, def)
	if def.Population > 5 { // hundreds of thousands
		def.Population = 5
	}
}

func rimFactions_noFactions(_ *util.TASContext, def *model.WorldDefinition) {
	def.Factions = nil
}

func rimGovernment_limitGovTypes(ctx *util.TASContext, def *model.WorldDefinition) {
	generateGovernment(ctx, def)

	// do not allow big government bureaucracies
	done := false
	for !done {
		if def.Government == 8 || def.Government == 9 {
			generateGovernment(ctx, def)
		} else {
			done = true
		}
	}
}

func rimLawLevel_noLaws(_ *util.TASContext, def *model.WorldDefinition) {
	def.LawLevel = 0
}

func rimStarport_maxCStarport(ctx *util.TASContext, def *model.WorldDefinition) {
	generateStarport(ctx, def)
	if def.Starport.Value >= 8 {
		def.Starport.Value = 8
		def.Starport.BerthingCost = (util.NewDice().Dx(6) * 100)
		def.Starport.HasHighport = false
	}
}

func rimBases_noBases(_ *util.TASContext, def *model.WorldDefinition) {
	def.Bases = nil
}

func rimTravelZone_atmoAmberOnly(_ *util.TASContext, def *model.WorldDefinition) {
	// only mark a world Amber if the atmosphere warrants it
	if def.Atmosphere >= 10 {
		def.TravelZone = "Amber"
		return
	}
	def.TravelZone = "Green"
}
