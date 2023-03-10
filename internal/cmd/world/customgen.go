package world

import (
	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"
)

const (
	//we shouldnt see many wrolds below this baseline level unless they have really regressed
	//the planet would have to be low-pop and offer no real challenges technology would be
	//needed to overcome.  Interestingly, this is the 'high-average' roll on a 1d6, which
	//when using the standard tech level generator, sets the tech baseline /  base tech level
	// in that algorithm.
	baseTechLevelAyInhabitedWorld = 4
)

func customHydrographics_FixAirlessWaterWorlds(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	if def.Size <= 1 || def.Atmosphere == 0 {
		def.Hydrographics = 0
	} else {

		atmoMod := 0
		atmoMod = h.AdjustDM(ctx, atmoMod, -4, def.Atmosphere, h.IS, 1, 10, 11, 12, 13, 14, 15)

		tempMod := 0
		if def.Atmosphere != 13 && def.Atmosphere != 15 {
			tempMod = h.AdjustDM(ctx, tempMod, -2, def.Temperature, h.INR, 10, 11)
			tempMod = h.AdjustDM(ctx, tempMod, -6, def.Temperature, h.EQ, 12)
		}

		hydro := dice.Sum(2, -7, def.Atmosphere, atmoMod, tempMod)
		hydro = util.BoundTo(hydro, hydroMin, hydroMax)
		def.Hydrographics = hydro
	}
	log.Debug().Str("custom", "customHydrographics_FixAirlessWaterWorlds").Int("hydro", def.Hydrographics).Send()
}

func customTechLevel_FixLowTechValues(ctx *util.TASContext, def *model.WorldDefinition) {

	log := ctx.Logger()
	dice := ctx.Dice()

	//base tech - we shouldnt see many wrolds below this baseline level unless they have really regressed
	techLevel := baseTechLevelAyInhabitedWorld

	//adjust tech for extremes - if people live in a place, tech needs to be there to support them
	//unless there are special circumstances.
	//the approach below moves the floor tech value upward to account for the solution of problems that would arise under extrme conditions.
	//You can't live on an airless, melting asteroid without tech. It is true that a bunch of backward/unsophisticated people may live in some
	//domed paradise well beyond their comprehension and capability to fix but this should not be the norm. A "harsh world" will require tech
	//that is sustained if the population is to survive.  I think these situations should fall to GM fiat, where they can choose a higher-tech
	//world generated out of this process and intentionally make it a backward low-tech yet domed utopia rather than have all planets in that
	//position be backward, low-tech domed utopias, which is what the standard algorithm would have us believe.

	//Also, the tech level generated here is not indicative of the highest (or lowest) tech on the planet.  It represents the tech that is
	//available wherever people are the most-dense.  The Australian outback is decidedly low tech, but no one would call Australia a low tech or
	//3rd world country.  The tech is where the people are - no people, no tech. Likewise, a few independent, insular communities living on the far side of
	//a planet might well have a lower available tech than the class A starport located around the globe from their position.

	//Anyway - the final tech will be the lowest level of tech needed to plausibly solve the most difficult issue faced by those living on the world

	//size
	switch def.Size {
	case 0, 1:
		techLevel = util.BoundTo(techLevel, 8, techMax) //need to actually be able to have reliable colonization in-system
	case 10:
		techLevel = util.BoundTo(techLevel, 8, techMax) //gravity manipulation helps build life on these planets
	}

	//atmosphere
	switch def.Atmosphere {
	case 2, 4, 7, 9:
		techLevel = util.BoundTo(techLevel, 8, techMax) //need some tech to filter
	case 13:
		techLevel = util.BoundTo(techLevel, 9, techMax) //need baseline colonization
	case 10, 11, 12, 15:
		techLevel = util.BoundTo(techLevel, 10, techMax) //need advanced otherworld colonization
	}

	//Hydrographics
	switch def.Hydrographics {
	case 0:
		techLevel = util.BoundTo(techLevel, 9, techMax) //need basic colonization to make own water
	case 10:
		techLevel = util.BoundTo(techLevel, 10, techMax) //need advanced colonization to live on nothing but water
	}

	//Temperature
	switch def.Temperature {
	case 2, 12:
		techLevel = util.BoundTo(techLevel, 8, techMax) //need basic colonization to live in temp extremes
	}

	//Population
	switch def.Population {
	case 1, 2:
		techLevel = util.BoundTo(techLevel, 8, techMax) //need basic colonization to live with just a few people if they arent going to be worked to death
	case 4, 5:
		techLevel = util.BoundTo(techLevel, 9, techMax) //need basic colonization / terraforming to support this population
	case 6, 7:
		techLevel = util.BoundTo(techLevel, 10, techMax) //highly populated worlds need or will want basic jump drive capabilities and orbital factories
	case 8:
		techLevel = util.BoundTo(techLevel, 11, techMax) //this is a major population center and will enjoy advanced tech over their smaller counterparts
	case 9, 10, 11, 12:
		techLevel = util.BoundTo(techLevel, 12, techMax) //massive societies will have access to this tech
	}

	//tech is silent on government and law. I am not sure how it would matter, arguements could be made either way

	//Starport
	switch def.Starport.Value {
	case 5, 6: //class D
		techLevel = util.BoundTo(techLevel, 9, techMax) //you wont have a usable starport without space ships
	case 7, 8: //class C
		techLevel = util.BoundTo(techLevel, 10, techMax) //possible high port and regular small craft ship yards require more advanced space tech
	case 9, 10: //class B
		techLevel = util.BoundTo(techLevel, 11, techMax) //need this level of AI and other tech to build starships
	case 11: //class A
		techLevel = util.BoundTo(techLevel, 12, techMax) //likelihood of bases and highport pushes tech up higher than Class B
	}

	//at this point we have a baseline established, add a bit of randomness
	//allow tech to drift downward (indicating infrastrucural decay or remoteness)
	//or increase a bit for whatever reason
	adj := dice.Roll()
	switch adj {
	case 1:
		techLevel += -2
	case 2:
		techLevel += -1
	case 6:
		techLevel += 1
	}

	//edge case no population means no tech under almost all circumstances
	if def.Population == 0 {
		techLevel = 0
	}

	def.TechLevel = techLevel
	log.Debug().Str("custom", "customTechLevel_FixLowTechValues").Int("techLevel", def.TechLevel).Send()

}
