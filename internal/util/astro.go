package util

import "math"

const (
	habitableMinAUx100 = 95
	habitableMaxAUx100 = 167
)

// For traveller, we need to take into account what we already know from the UWP - if there is a
// dense atmosphere, then there needs to be heat to allow the creation of a planetary dynamo to generate
// a magnetic field to keep solar wind from stripping away the atmosphere (not always true, this magnetic protection
// might be acheived through other means like the sun - see Venus). There also needs to be replenishment
// of atmosphere lost to low grav or stripped away by radiation - either the solar wind or radiation from a gas giant like Jupiter.
// This replenishment is via outgassing of geothermal gasses - volcanism or otherwise.  Substantial gravity might
// cause geothermal activity (see Jupiter's moon Io) even if the planet is too small to otherwise retain internal heat,
// which is radiated away quickly on smaller planets, or lack the gravity to maintain an atmosphere. Given all this
// complexity, the rule below forces us to align what we do not know - planetary geothermics, with what we do know - atmospheric
// density.  Do note that the converse is true as well. It is possible for a planet to be geothermally active but have no atmosphere
// because it is too small to retain it. Some of Saturn's moons are like this, venting internal gasses directly into space rather
// holding them.  So - it is really hard to determine geothermic activity using hard(er) science. This approach says "given that the planet
// has a dense atmosphere, it must get that from somehwere, regardless of size or promimity to a gravity source, so let's say it is active".
// Note that we do not say _why_ it is active (internal heat left over from formation, radioactives, tidal grav forces or all 3) nor
// do we try to develop things based on size and nearby gravity. We assume that if there is a thick atmosphere, then there is a reason
// and the GM can sort out how or why this is. Note that the UWP is very loose here, in that a Size 1 planet can have an atmosphere of
// 2D-7 + 1 = 6 Standard). This is really stretching things, but I didnt make the tables and am leaning into them here.
func IsGeothermallyActive(atmosphere int) bool {
	return IsDenseAtmosphere(atmosphere)
}

// This is a another gross simplification, but in general, if there is surface fluid available (usually water)
// and heat, then the planet's crust will float on the mantle instead of baking into an immovable blanket. All
// kinds of interesting developments can occur on planets with active tectonics - varied life, formation of certain
// minerals, outgassing that replenishes the atmosphere (though this can happen through eruptions as well)
func HasPlateTectonics(isGeothermallyActive bool, hydrographics int) bool {
	return isGeothermallyActive && hydrographics >= 1
}

// Rotational period is based on conservation of angular momentum from the time the planet was formed
// as well as pure chance (e.g. Venus takes almost a year to rotate because of a huge collision eons ago. It
// also rotates backwards for the same reason). This is just a random roll, and assumes a rocky planet.
// Gas giants tend to rotate very much faster. This will give a range of .5 of an earth day to 2 earth days
func DetermineRotationalPeriodInEarthDays(dice Dice) float32 {
	var baseline float32
	baseline = 1.0
	roll := dice.Dx(20)
	if roll <= 10 {
		roll = -roll
	}
	delta := float32(roll) / float32(20)
	return baseline + delta
}

// this is Keplar's Third Law, simplified for average distance of the planet from the sun during
// its eliptical orbit (expressed in AU) and output in earth years. T^2 = a^3. This assumes the
// planet has much, much less mass than the sun it is orbiting. No sense getting all Newtonian at
// this point.
func DetermineOrbitalPeriodInEarthYears(distAU float32) float32 {
	period := math.Sqrt(math.Pow(float64(distAU), 3))
	return float32(period)
}

// Determine where in the Goldilocks zone this planet sits. Returns distance in AU
// and assumes a Sol-like star
func HabitableZoneDistanceFromSun(dice Dice) float32 {
	span := habitableMaxAUx100 - habitableMinAUx100
	halfspan := span / 2
	midval := halfspan + habitableMinAUx100

	roll := dice.Dx(span)
	if roll < halfspan {
		roll = -roll
	} else {
		roll = roll - halfspan
	}
	newval := midval + roll
	newval = BoundTo(newval, habitableMinAUx100, habitableMaxAUx100)
	return float32(newval) / 100
}

// determine which part of the Habitable Zone a planet sits in
// based on AU from a SOl-like sun
func HabitableZoneTemperatureProfile(distAU float32) string {
	if distAU >= .95 && distAU <= 1.19 {
		return "hot"
	}
	if distAU > 1.19 && distAU <= 1.43 {
		return "temperate"
	}
	return "cold"
}

// returns true if the provided atmosphere is considered dense
func IsDenseAtmosphere(atmosphere int) bool {
	switch atmosphere {
	case 6, 7, 8, 9, 10, 11, 12, 13, 15:
		return true
	}
	return false
}
