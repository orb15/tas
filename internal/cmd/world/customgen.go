package world

import (
	"tas/internal/model"
	"tas/internal/util"
)

// adds certain custom information that might be important, or at least
// add virsimlitude
func calcCustomDetails(ctx *util.TASContext, def *model.WorldDefinition) {
	dice := ctx.Dice()

	details := model.CustomDetailsInfo{}
	details.IsGeothermallyActive = util.IsGeothermallyActive(def.Atmosphere)
	details.HasPlateTectonics = util.HasPlateTectonics(details.IsGeothermallyActive, def.Hydrographics)
	details.DistanceFromSunAU = util.HabitableZoneDistanceFromSun(dice)
	details.HabitabiltyZoneTempProfile = util.HabitableZoneTemperatureProfile(details.DistanceFromSunAU)
	details.OrbitalPeriodEarthYears = util.DetermineOrbitalPeriodInEarthYears(details.DistanceFromSunAU)
	details.RotationalPeriodEarthDays = util.DetermineRotationalPeriodInEarthDays(dice)
	def.CustomDetails = details
}
