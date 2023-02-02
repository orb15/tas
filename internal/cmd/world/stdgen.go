package world

import (
	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"
)

// ---------------------------------------
// Size is 2D-2 pg 249
// ---------------------------------------
func generateSize(ctx *util.TASContext, def *model.WorldDefinition) {
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
func generateAtmosphere(ctx *util.TASContext, def *model.WorldDefinition) {
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
// Temperature - needed to do hydrographics. There are several rules at play, see pg 251
// Also, I am using a special value here to represent "roasting in day, freezing at night"
// ---------------------------------------
func generateTemperature(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	if def.Atmosphere <= 1 {
		def.Temperature = specialTempCodeForNoAtmo
		def.HabitabilityZone = "standard"
	} else {
		atmoMod := 0
		atmoMod = h.AdjustDM(ctx, atmoMod, -2, def.Atmosphere, h.INR, 2, 3)
		atmoMod = h.AdjustDM(ctx, atmoMod, -1, def.Atmosphere, h.IS, 4, 5, 14)
		atmoMod = h.AdjustDM(ctx, atmoMod, 1, def.Atmosphere, h.INR, 8, 9)
		atmoMod = h.AdjustDM(ctx, atmoMod, 6, def.Atmosphere, h.INR, 11, 12)
		atmoMod = h.AdjustDM(ctx, atmoMod, 2, def.Atmosphere, h.IS, 10, 13, 15)

		//this is optional, but I am adding a random location within the "habital zone"
		//in a star system. This is optional per pg 251
		zoneMod := 0
		habZone := "standard"
		zone := dice.Sum(2)
		switch zone {
		case 2:
			zoneMod = -4
			habZone = "extreme cold"
		case 3:
			zoneMod = -2
			habZone = "cold"
		case 11:
			zoneMod = 2
			habZone = "hot"
		case 12:
			zoneMod = 4
			habZone = "extreme hot"
		}

		temp := dice.Sum(2, atmoMod, zoneMod)
		temp = util.BoundTo(temp, tempMin, tempMax)
		def.Temperature = temp
		def.HabitabilityZone = habZone

	}
	log.Debug().Str("habitability-zone", def.HabitabilityZone).Int("temp", def.Temperature).Send()
}

// ---------------------------------------
// Hydrographics 2D-7 + Atmosphere. see pg 251 for various conditions
// ---------------------------------------
func generateHydrographics(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	if def.Size <= 1 {
		def.Hydrographics = 0
	} else {

		atmoMod := 0
		atmoMod = h.AdjustDM(ctx, atmoMod, -4, def.Atmosphere, h.IS, 0, 1, 10, 11, 12, 13, 14, 15)

		tempMod := 0
		if def.Atmosphere != 13 && def.Atmosphere != 15 {
			tempMod = h.AdjustDM(ctx, tempMod, -2, def.Temperature, h.INR, 10, 11)
			tempMod = h.AdjustDM(ctx, tempMod, -6, def.Temperature, h.EQ, 12)
		}

		hydro := dice.Sum(2, -7, def.Atmosphere, atmoMod, tempMod)
		hydro = util.BoundTo(hydro, hydroMin, hydroMax)
		def.Hydrographics = hydro
	}
	log.Debug().Int("hydro", def.Hydrographics).Send()
}

// ---------------------------------------
// Population is 2D-2 pg 252
// ---------------------------------------
func generatePopulation(ctx *util.TASContext, def *model.WorldDefinition) {

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
func generateGovernment(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

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
// Factions D3, special see pg 254
// ---------------------------------------
func generateFactions(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	var factionsList []*model.WorldFaction
	if def.Population > 0 {
		fmod := 0
		fmod = h.AdjustDM(ctx, fmod, 1, def.Government, h.IS, 0, 7)
		fmod = h.AdjustDM(ctx, fmod, -1, def.Government, h.GE, 10)

		numberOfFactions := dice.D3(fmod)
		factionsList = make([]*model.WorldFaction, 0, numberOfFactions)
		for i := 0; i < numberOfFactions; i++ {
			gov := dice.Sum(2, -7, def.Population)
			gov = util.BoundTo(gov, govMin, govMax)
			strength := dice.Sum(2)
			f := &model.WorldFaction{
				GovernmentStyle:  gov,
				RelativeStrength: strength,
			}
			factionsList = append(factionsList, f)
		}
	}
	def.Factions = factionsList
	log.Debug().Int("number of factions", len(factionsList)).Send()
}

// ---------------------------------------
// Culture D66 see pg 254.
// Using a special code for "no culture because of no government"
// ---------------------------------------
func generateCulture(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	//TODO:  handle culture roll 26 - fusion / reroll twice

	def.Culture = specialCultureCodeForNoPop
	if def.Population > 0 {
		def.Culture = dice.D66()
	}
	log.Debug().Int("culture", def.Culture).Send()
}

// ---------------------------------------
// Law Level 2D-7 + Gov see pg 256
// ---------------------------------------
func generateLawLevel(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

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
func generateStarport(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	popMod := 0
	popMod = h.AdjustDM(ctx, popMod, 1, def.Population, h.INR, 8, 9)
	popMod = h.AdjustDM(ctx, popMod, 2, def.Population, h.GE, 10)
	popMod = h.AdjustDM(ctx, popMod, -1, def.Population, h.INR, 3, 4)
	popMod = h.AdjustDM(ctx, popMod, -2, def.Population, h.LE, 2)

	star := dice.Sum(2, popMod)
	star = util.BoundTo(star, starMin, starMax)
	starport := &model.WorldStarportInfo{}
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
	def.Starport = starport
	log.Debug().Int("starport", def.Starport.Value).Send()
}

// ---------------------------------------
// Tech Level 1D + special see pg 258
// ---------------------------------------
func generateTechLevel(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	//this is not a game rule per se, but I saw asteroids with no people generating a tech level of 5
	//and I dont think tech level matters if there is no one to establish one or buy/sell to
	if def.Population == 0 {
		def.TechLevel = 0
	} else {

		//starport modifier
		starMod := 0
		starMod = h.AdjustDM(ctx, starMod, -4, def.Starport.Value, h.EQ, 2)
		starMod = h.AdjustDM(ctx, starMod, 2, def.Starport.Value, h.INR, 7, 8)
		starMod = h.AdjustDM(ctx, starMod, 4, def.Starport.Value, h.INR, 9, 10)
		starMod = h.AdjustDM(ctx, starMod, 6, def.Starport.Value, h.EQ, 11)

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
		govMod = h.AdjustDM(ctx, govMod, 1, def.Government, h.IS, 0, 4)
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
func generateHighport(ctx *util.TASContext, def *model.WorldDefinition) {

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
// Bases see pg 257
// ---------------------------------------
func generateBases(ctx *util.TASContext, def *model.WorldDefinition) {
	baseList := make([]string, 0)

	log := ctx.Logger()
	dice := ctx.Dice()

	corsairLawMod := 0
	corsairLawMod = h.AdjustDM(ctx, corsairLawMod, 2, def.LawLevel, h.EQ, 0)
	corsairLawMod = h.AdjustDM(ctx, corsairLawMod, -2, def.LawLevel, h.GE, 2)

	switch def.Starport.Value {
	case 2, 3, 4:
		if dice.Sum(2, corsairLawMod) >= 10 {
			baseList = append(baseList, "corsair")
		}
	case 5, 6:
		if dice.Sum(2, corsairLawMod) >= 12 {
			baseList = append(baseList, "corsair")
		}
		if dice.Sum(2) >= 8 {
			baseList = append(baseList, "scout")
		}
	case 7, 8:
		if dice.Sum(2) >= 10 {
			baseList = append(baseList, "military")
		}
		if dice.Sum(2) >= 9 {
			baseList = append(baseList, "scout")
		}
	case 9, 10:
		if dice.Sum(2) >= 8 {
			baseList = append(baseList, "military")
		}
		if dice.Sum(2) >= 8 {
			baseList = append(baseList, "naval")
		}
		if dice.Sum(2) >= 9 {
			baseList = append(baseList, "scout")
		}
	case 11:
		if dice.Sum(2) >= 8 {
			baseList = append(baseList, "military")
		}
		if dice.Sum(2) >= 8 {
			baseList = append(baseList, "naval")
		}
		if dice.Sum(2) >= 10 {
			baseList = append(baseList, "scout")
		}
	}

	def.Bases = baseList
	log.Debug().Int("bases present", len(def.Bases)).Send()
}

// ---------------------------------------
// Travel Zone see pg 260
// ---------------------------------------
func generateTravelCode(ctx *util.TASContext, def *model.WorldDefinition) {

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
func generateTradeCodes(ctx *util.TASContext, def *model.WorldDefinition) {

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

	//fluid oceans
	if def.Atmosphere >= 10 && def.Hydrographics >= 1 {
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
