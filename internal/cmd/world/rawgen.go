package world

import (
	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"
)

/*
Deltas from RAW

The following generators try to maintain a true-to-RAW experience so UWPs are as "RAW as possible". In some cases
I have made the deliberate choice to NOT do this.  Those choices are noted here:

1. For starport generation, if Population = 0, then Starport is of type X (non pop -> no starport)
2. For the Fluid Ocens (Fl) Trade Code, which indicates that the oceans are some kind of non-water fluid, this is, by RAW
supposed to be assigned when Atmosphere is 10+ and Hydrographics are 1+, but Fl should only apply if Atmosphere is 10-12 inclusive
(exotic, corrosive or insidious) as Atmosphere of 13 is 'very dense', 14 is 'low' and 15 is 'unusual' - which I read as GM fiat.
*/

// ---------------------------------------
// Size is 2D-2 pg 249
// ---------------------------------------
func rawGenerateSize(ctx *util.TASContext, def *model.WorldDefinition) {
	log := ctx.Logger()
	dice := ctx.Dice()

	size := dice.Sum(2, -2)
	size = util.BoundTo(size, sizeMin, sizeMax)
	def.Size = size
	log.Debug().Int("size", def.Size).Send()
}

// ---------------------------------------
// Atmosphere is 2D-7 + size pg 250 but is nil for small worlds
// ---------------------------------------
func rawGenerateAtmosphere(ctx *util.TASContext, def *model.WorldDefinition) {
	log := ctx.Logger()
	dice := ctx.Dice()

	def.Atmosphere = 0
	if def.Size > 1 {
		atmos := dice.Sum(2, -7, def.Size)
		atmos = util.BoundTo(atmos, atmoMin, atmoMax)
		def.Atmosphere = atmos
		log.Debug().Int("atmos", def.Atmosphere).Send()
	}
}

// ---------------------------------------
// Hydrographics 2D-7 + Atmosphere. see pg 251 for various conditions
// ---------------------------------------
func rawGenerateHydrographics(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	if def.Size <= 1 {
		def.Hydrographics = 0
	} else {
		atmoMod := 0
		atmoMod = h.AdjustDM(ctx, atmoMod, -4, def.Atmosphere, h.IS, 0, 1, 10, 11, 12, 13, 14, 15)

		hydro := dice.Sum(2, -7, def.Atmosphere, atmoMod)
		hydro = util.BoundTo(hydro, hydroMin, hydroMax)
		def.Hydrographics = hydro
	}
	log.Debug().Int("hydro", def.Hydrographics).Send()
}

// ---------------------------------------
// Population is 2D-2 pg 252
// ---------------------------------------
func rawGeneratePopulation(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	pop := dice.Sum(2, -2)
	pop = util.BoundTo(pop, popMin, popMax)
	def.Population = pop
	log.Debug().Int("pop", def.Population).Send()
}

// ---------------------------------------
// Government 2D-7 + Pop + special see 252
// ---------------------------------------
func rawGenerateGovernment(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	//per page 252
	if def.Population == 0 {
		def.Government = 0
	} else {
		gov := dice.Sum(2, -7, def.Population)
		gov = util.BoundTo(gov, govMin, govMax)
		def.Government = gov
	}
	log.Debug().Int("gov", def.Government).Send()
}

// ---------------------------------------
// Law Level 2D-7 + Gov see pg 256
// ---------------------------------------
func rawGenerateLawLevel(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	//per page 252
	if def.Population == 0 {
		def.LawLevel = 0
	} else {
		law := dice.Sum(2, -7, def.Government)
		law = util.BoundTo(law, lawMin, lawMax)
		def.LawLevel = law
	}
	log.Debug().Int("law level", def.LawLevel).Send()
}

// ---------------------------------------
// Starport 2D + special, see pg 257
// ---------------------------------------
func rawGenerateStarport(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()
	starport := &model.WorldStarportInfo{}

	//per page 252 - well, if there is no population, there cannot be a starport but this is not
	//explicitly in the rules. I think it should have been though
	if def.Population == 0 {
		starport.Value = starMin
	} else {
		popMod := 0
		popMod = h.AdjustDM(ctx, popMod, 1, def.Population, h.INR, 8, 9)
		popMod = h.AdjustDM(ctx, popMod, 2, def.Population, h.GE, 10)
		popMod = h.AdjustDM(ctx, popMod, -1, def.Population, h.INR, 3, 4)
		popMod = h.AdjustDM(ctx, popMod, -2, def.Population, h.LE, 2)

		star := dice.Sum(2, popMod)
		star = util.BoundTo(star, starMin, starMax)
		starport.Value = star

		switch starport.Value {
		case 2, 3, 4:
			starport.HasHighport = false
			starport.BerthingCost = 0
		case 5, 6:
			starport.BerthingCost = dice.Roll(1) * 10
		case 7, 8:
			starport.BerthingCost = dice.Roll(1) * 100
		case 9, 10:
			starport.BerthingCost = dice.Roll(1) * 500
		case 11:
			starport.BerthingCost = dice.Roll(1) * 1000
		}
	}
	def.Starport = starport
	log.Debug().Int("starport", def.Starport.Value).Send()
}

// ---------------------------------------
// Tech Level 1D + special see pg 258
// ---------------------------------------
func rawGenerateTechLevel(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	//per page 252
	if def.Population == 0 {
		def.TechLevel = 0
	} else {
		//starport modifier
		starMod := 0
		starMod = h.AdjustDM(ctx, starMod, 6, def.Starport.Value, h.EQ, 11)     //Class A
		starMod = h.AdjustDM(ctx, starMod, 4, def.Starport.Value, h.INR, 9, 10) //Class B
		starMod = h.AdjustDM(ctx, starMod, 2, def.Starport.Value, h.INR, 7, 8)  //Class C
		starMod = h.AdjustDM(ctx, starMod, -4, def.Starport.Value, h.EQ, 2)     //Class X

		//size modifier
		sizeMod := 0
		sizeMod = h.AdjustDM(ctx, sizeMod, 2, def.Size, h.LE, 1)
		sizeMod = h.AdjustDM(ctx, sizeMod, 1, def.Size, h.INR, 2, 4)

		//atmosphere mod
		atmoMod := 0
		atmoMod = h.AdjustDM(ctx, atmoMod, 1, def.Atmosphere, h.LE, 3)
		atmoMod = h.AdjustDM(ctx, atmoMod, 1, def.Atmosphere, h.GE, 10)

		//hydrographics mod
		hydroMod := 0
		hydroMod = h.AdjustDM(ctx, hydroMod, 1, def.Hydrographics, h.IS, 0, 9)
		hydroMod = h.AdjustDM(ctx, hydroMod, 2, def.Hydrographics, h.EQ, 10)

		//population modifier
		popMod := 0
		popMod = h.AdjustDM(ctx, popMod, 1, def.Population, h.IS, 1, 2, 3, 4, 5, 8)
		popMod = h.AdjustDM(ctx, popMod, 2, def.Population, h.EQ, 9)
		popMod = h.AdjustDM(ctx, popMod, 4, def.Population, h.EQ, 10)

		//government
		govMod := 0
		govMod = h.AdjustDM(ctx, govMod, 1, def.Government, h.IS, 0, 5)
		govMod = h.AdjustDM(ctx, govMod, 2, def.Government, h.EQ, 7)
		govMod = h.AdjustDM(ctx, govMod, -2, def.Government, h.INR, 13, 14)

		techMods := starMod + sizeMod + atmoMod + hydroMod + popMod + govMod

		tech := dice.Roll(starMod, sizeMod, atmoMod, hydroMod, popMod, govMod)

		if tech < 1+techMods || tech > 6+techMods {
			log.Warn().Int("starMod", starMod).Int("sizeMod", sizeMod).Int("atmoMod", atmoMod).Int("hydroMod", hydroMod).Int("popMod", popMod).Int("govMod", govMod).
				Int("totalTech", tech).Msg("suspicious tech calculation")
		}

		tech = util.BoundTo(tech, techMin, techMax)

		//adjust tech level for atmospheric limits
		switch def.Atmosphere {
		case 0, 1, 10, 15:
			tech = util.BoundTo(tech, 8, techMax)
		case 2, 3, 13, 14:
			tech = util.BoundTo(tech, 5, techMax)
		case 4, 7, 9:
			tech = util.BoundTo(tech, 3, techMax)
		case 11:
			tech = util.BoundTo(tech, 9, techMax)
		case 12:
			tech = util.BoundTo(tech, 10, techMax)
		}

		def.TechLevel = tech
	}

	log.Debug().Int("tech level", def.TechLevel).Send()
}

// ---------------------------------------
// Highport see pg 257
// ---------------------------------------
func rawGenerateHighport(ctx *util.TASContext, def *model.WorldDefinition) {

	dice := ctx.Dice()

	highportTarget := 0
	switch def.Starport.Value {
	case 2, 3, 4:
		break //no highport possible
	case 5, 6:
		highportTarget = 12
	case 7, 8:
		highportTarget = 10
	case 9, 10:
		highportTarget = 8
	case 11:
		highportTarget = 6
	}

	if def.Starport.Value <= 4 {
		def.Starport.HasHighport = false
		return
	}

	hpTechMod := 0
	hpTechMod = h.AdjustDM(ctx, hpTechMod, 1, def.TechLevel, h.INR, 9, 11)
	hpTechMod = h.AdjustDM(ctx, hpTechMod, 2, def.TechLevel, h.GE, 12)

	hpPopMod := 0
	hpPopMod = h.AdjustDM(ctx, hpPopMod, 1, def.Population, h.GE, 9)
	hpPopMod = h.AdjustDM(ctx, hpPopMod, -1, def.Population, h.LE, 6)

	def.Starport.HasHighport = dice.Sum(2, hpTechMod, hpPopMod) >= highportTarget
}

// ---------------------------------------
// Travel Zone see pg 260
// ---------------------------------------
func rawGenerateTravelCode(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()

	isAmber := false

	if def.Atmosphere >= 10 {
		isAmber = true
	}

	if (def.Government == 0 || def.Government == 7 || def.Government == 10) && (def.LawLevel == 0 || def.LawLevel >= 9) {
		isAmber = true
	}

	code := "green"
	if isAmber {
		code = "amber"
	}
	def.TravelZone = code
	log.Debug().Str("travel code", def.TravelZone).Send()
}

// ---------------------------------------
// Trade Codes pg 260
// ---------------------------------------
func rawGenerateTradeCodes(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()

	codes := make([]string, 0)

	//agricultural
	if def.Atmosphere >= 4 && def.Atmosphere <= 9 && def.Hydrographics >= 4 && def.Hydrographics <= 8 && def.Population >= 5 && def.Population <= 7 {
		codes = append(codes, "agricultural")
	}

	//asteroid
	if def.Size == 0 && def.Atmosphere == 0 && def.Hydrographics == 0 {
		codes = append(codes, "asteroid")
	}

	//barren
	if def.Population == 0 && def.Government == 0 && def.LawLevel == 0 {
		codes = append(codes, "barren")
	}

	//desert
	if def.Atmosphere >= 2 && def.Atmosphere <= 9 && def.Hydrographics == 0 {
		codes = append(codes, "desert")
	}

	//fluid oceans - note this is not RAW, per RAW this should be Atmosphere 10+. This change
	//assigns Fl only when the atmosphere is exotic, corrosive or insidious rather than "dense" or other
	//less hostile things.
	if def.Atmosphere >= 10 && def.Atmosphere <= 12 && def.Hydrographics >= 1 {
		codes = append(codes, "fluid oceans")
	}

	//garden
	if def.Size >= 6 && def.Size <= 8 && (def.Atmosphere == 5 || def.Atmosphere == 6 || def.Atmosphere == 8) && def.Hydrographics >= 5 && def.Hydrographics <= 7 {
		codes = append(codes, "garden")
	}

	//high pop
	if def.Population >= 9 {
		codes = append(codes, "high population")
	}

	//high tech
	if def.TechLevel >= 12 {
		codes = append(codes, "high tech")
	}

	//ice capped
	if def.Atmosphere <= 1 && def.Hydrographics >= 1 {
		codes = append(codes, "ice-capped")
	}

	//industrial
	indAtmo := false
	switch def.Atmosphere {
	case 0, 1, 2, 4, 7, 9, 10, 11, 12:
		indAtmo = true
	}
	if indAtmo && def.Population >= 9 {
		codes = append(codes, "industrial")
	}

	//low pop
	if def.Population <= 3 {
		codes = append(codes, "low population")
	}

	//low tech
	if def.Population >= 1 && def.TechLevel <= 5 {
		codes = append(codes, "low tech")
	}

	//non-agricultural
	if def.Atmosphere <= 3 && def.Hydrographics <= 3 && def.Population >= 6 {
		codes = append(codes, "non-agricultural")
	}

	//non-industrial
	if def.Population >= 4 && def.Population <= 6 {
		codes = append(codes, "non-industrial")
	}

	//poor
	if def.Atmosphere >= 2 && def.Atmosphere <= 5 && def.Hydrographics <= 3 {
		codes = append(codes, "poor")
	}

	//rich
	if (def.Atmosphere == 6 || def.Atmosphere == 8) && def.Population >= 6 && def.Population <= 8 && def.Government >= 4 && def.Government <= 9 {
		codes = append(codes, "rich")
	}

	//vaccum
	if def.Atmosphere == 0 {
		codes = append(codes, "vacuum")
	}

	//waterworld
	watAtmo := false
	switch def.Atmosphere {
	case 3, 4, 5, 6, 7, 8, 9, 13:
		watAtmo = true
	}
	if watAtmo && def.Hydrographics >= 10 {
		codes = append(codes, "waterworld")
	}

	def.TradeCodes = codes
	log.Debug().Int("number of trade codes", len(def.TradeCodes)).Send()
}
