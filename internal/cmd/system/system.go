package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

const (
	uwpRegExString = `([a-zA-Z0-9\- \.']{1,}) [0-9]{4} [A-EX]{1}([0-9A]{1})[0-9A-F]{5}\-[0-9A-F]`
	uwpIsAsteroid  = `\-[0-9A-F].* AS.*`
	uwpIsFluid     = `\-[0-9A-F].* FL.*`
	uwpIsIcecap    = `\-[0-9A-F].* IC.*`

	minAUBandRockyWorldsInner = .37 // 37 an ugly prime - looks jittery
	minAUBandRockyWorldsOuter = .63 // 63 an ugly prime - looks jittery
	minAUBandGasGiants        = 4.3 // 43 is also ugly - looks jittery

	minHabitableAU  = 0.95
	maxHabitalbleAU = 1.67

	ordinalUnset = -1

	maxAsteroidBeltCount = 3

	sunCloseCutoffAU                      = .2
	distanceToSunToBeConsideredNearStarAU = .33

	coldAUOffset = 2.63

	systemTagsFile = "system-tags.json"
)

var (
	uwpRegExp     = regexp.MustCompile(uwpRegExString)
	uwpIsASRegExp = regexp.MustCompile(uwpIsAsteroid)
	uwpIsFLRegExp = regexp.MustCompile(uwpIsFluid)
	uwpIsICRegExp = regexp.MustCompile(uwpIsIcecap)
)

var SystemCmdConfig = &cobra.Command{

	Use:   "system",
	Short: "builds a solar system",
	Run:   systemCmd,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("exactly 3 arguments required - a quoted string for the UWP, the location of the primary in AU from the sun and a true/false indicator if the system contains gas giants")
		}

		_, _, err := parseUWP(args[0])
		if err != nil {
			return fmt.Errorf("first argument must be a valid UWP. %w", err)
		}

		_, err = parseAU(args[1])
		if err != nil {
			return fmt.Errorf("Second argument must be a valid AU distance. %w", err)
		}

		_, err = parseGasGiants(args[2])
		if err != nil {
			return fmt.Errorf("third argument must be a true/false value %w", err)
		}

		return nil
	},
}

func systemCmd(cmd *cobra.Command, args []string) {

	//create a config to hold all data passed into this call
	cfg, err := util.NewTASConfig().
		WithArgs(args).
		WithCmd(cmd)
	if err != nil {
		fmt.Println()
		fmt.Printf("Unable to create config. This is a critical error: %s\n", err)
		fmt.Println()
		return
	}

	//build a context to make all data easily available between calls
	loglevel, _ := cfg.Flags.GetString(util.LogLevelFlagName)
	log := util.NewLogger(loglevel)
	ctx := util.NewContext().
		WithLogger(log).
		WithDice().
		WithConfig(cfg)

	//core info. Note args already validated above
	uwp := args[0]
	primaryAU, _ := parseAU(args[1])
	hasGasGiants, _ := parseGasGiants(args[2])

	//load the system tags from JSON
	systemTags, err := loadSystemTags(ctx)
	if err != nil {
		log.Err(err).Msg("failed to load system tags file")
		return
	}

	// create the solar system
	solarsys := generateSolarSystem(ctx, uwp, primaryAU, hasGasGiants, systemTags)

	//dump the solar system for now
	bytes, err := json.MarshalIndent(*solarsys, "", " ")
	if err != nil {

	}
	fmt.Println(string(bytes))

}

// top-level generator function
func generateSolarSystem(ctx *util.TASContext, uwp string, primaryAuDist float32, giantsPresent bool, systemTags *model.SystemTags) *model.SolarSystem {

	log := ctx.Logger()
	log.Info().Msg("generating solar system...")

	// parse the UWP - it is already validated
	systemName, primarySize, _ := parseUWP(uwp)

	//very basics
	sys := &model.SolarSystem{}
	sys.Name = systemName
	sys.Orbitals = make([]*model.SystemOrbital, 0, 16)
	sys.HasGasGiants = giantsPresent

	//primary basics
	primary := &model.SystemOrbital{}
	primary.IsPrimary = true
	primary.IsGasGiant = false
	primary.UWP = uwp
	primary.OrbitalDistanceAU = primaryAuDist
	primary.Size = primarySize
	primary.OrbitalNumber = ordinalUnset
	primary.Moons = make([]model.Moon, 0)
	sys.Orbitals = append(sys.Orbitals, primary)

	//the trade codes provide implied information about where in solar system the primary is located or provide other
	//information about the primary's location. The rest of the solar system needs to be built around this information
	//in a believable way.
	switch {
	case worldHasTradeCode(uwp, uwpIsASRegExp):
		//no one lives on an asteroid unless they must. This means no other possible world is
		//remotely habitable or exists
		generateJustAsteroidBelt(ctx, sys)
	case worldHasTradeCode(uwp, uwpIsFLRegExp) || worldHasTradeCode(uwp, uwpIsICRegExp):
		// planet is very cold (3+ AU) and non-water liquid on surface (Fl) or it has boiled away (Ic)
		sys.Orbitals[0].OrbitalDistanceAU = calcColdAU(ctx, sys.Orbitals[0].OrbitalDistanceAU)
		generateSolarSystemInsideOut(ctx, sys, systemTags)
		return sys
	default:
		//the planet is in the habitable zone (by default) - it may be a garden or barren, but it is better than anything else
		//simply because it is in the habitable zone
		generateSolarSystemInsideOut(ctx, sys, systemTags)
	}

	log.Info().Msg("solar system generation complete")
	return sys
}

// people live on an asteroid belt - so the solar system must be very strange
// because people will live almost anywhere other than a barren misshaped rock at 0g
func generateJustAsteroidBelt(ctx *util.TASContext, sys *model.SolarSystem) {

	dice := ctx.Dice()

	// note we have a primary asteroid
	sys.Orbitals[0].IsAsteroid = true
	sys.TotalOrbitalCount = 1

	//no one lives on an asteroid if there is a remotely reasonable planet to live on instead
	sys.RockyCount = 0

	//if there are gas giants present, there will be only 1. It will have no moons b/c otherwise
	//people would likely move there instead. If there were more than 1 giant and none had moons,
	//that would be suspect.
	if sys.HasGasGiants {

		sys.TotalOrbitalCount += 1
		sys.GasGiantCount = 1

		var nextAU float32
		currentAU := sys.Orbitals[0].OrbitalDistanceAU

		nextAU = currentAU + minAUBandGasGiants

		// gas giants can be farther apart. Add some variablility
		bump := dice.Roll()
		switch bump {
		case 1, 2:
			// do nothing
		case 3, 4:
			nextAU += float32(dice.Dx(5))
		case 5, 6:
			nextAU += float32(dice.Dx(9))
		}

		gas := &model.SystemOrbital{}
		gas.OrbitalDistanceAU = nextAU
		gas.IsGasGiant = true
		gas.OrbitalNumber = ordinalUnset
		gas.Size = calcGasGiantSizeSize(ctx)
		sys.Orbitals = append(sys.Orbitals, gas)
	}

	//sort and add orbital numbers
	addSystemOrdinals(sys)
}

// this function is for building out a full solar system when the primary is a rocky planet
func generateSolarSystemInsideOut(ctx *util.TASContext, sys *model.SolarSystem, systemTags *model.SystemTags) {

	sys.TotalOrbitalCount = determineTotalOrbitalCount(ctx)

	//determine number of asteroid belts, gas giants and rocky planets. These totals EXCLUDE the Primary!
	rocky, gas, belts := determineOtherOrbitalCounts(ctx, sys.TotalOrbitalCount, sys.HasGasGiants)
	sys.AsteroidCount = belts
	sys.GasGiantCount = gas
	sys.RockyCount = rocky

	//general idea and constraints
	// - we know where the primary is located (typically habitable zone but not required)
	// - fill in the inner systems first (rocky or asteroid) working from the primary inward
	// - fill in the outer systems (rocky  or asteroid) working from the primary outward
	// - add gias giants and their moons outside the last rocky planet or belt
	// - respect the 'dead bands' to prevent planets from being too close to one another

	// do we have anything to place inside the primary habitable zone?
	beltsRemain := belts
	rockyRemain := rocky
	currentAU := sys.Orbitals[0].OrbitalDistanceAU

	innerBeltPlaced := false
	innerIsFull := beltsRemain+rockyRemain == 0 || currentAU < sunCloseCutoffAU
	dice := ctx.Dice()
	var nextAU float32

	loopCount := 0

	//fill the inner area, working from primary inward
	for !innerIsFull {

		// crash prevent
		loopCount++
		if loopCount == 100 {
			ctx.Logger().Fatal().
				Float32("nextAU", nextAU).
				Int("beltsRemain", beltsRemain).
				Int("rockyRemain", rockyRemain).
				Msg("Inner crash!")
		}

		nextAU = currentAU - minAUBandRockyWorldsInner

		//if we are still in the habitable zone, try again
		if nextAU >= minHabitableAU {
			currentAU = nextAU
			continue
		}

		// no more room, we are done
		if nextAU < sunCloseCutoffAU {
			break
		}

		switch {

		// we only want one inner belt to be placed
		case beltsRemain > 0 && !innerBeltPlaced:
			if dice.D3() == 1 {
				belt := &model.SystemOrbital{}
				belt.IsAsteroid = true
				belt.OrbitalDistanceAU = nextAU
				belt.OrbitalNumber = ordinalUnset
				belt.Size = 0
				sys.Orbitals = append(sys.Orbitals, belt)

				//upkeep
				innerBeltPlaced = true
				beltsRemain -= 1
				currentAU = nextAU
				innerIsFull = beltsRemain+rockyRemain == 0 || currentAU < sunCloseCutoffAU
				continue
			}

			//if we have a rocky to place, place it
		case rockyRemain > 0:
			planet := &model.SystemOrbital{}
			planet.OrbitalDistanceAU = nextAU
			planet.OrbitalNumber = ordinalUnset
			planet.Size = calcRockyPlanetSize(ctx)
			sys.Orbitals = append(sys.Orbitals, planet)

			//upkeep
			rockyRemain -= 1
			currentAU = nextAU
			innerIsFull = beltsRemain+rockyRemain == 0 || currentAU < sunCloseCutoffAU
			continue

		// if we have no more rocky planets and cannot place another belt, we are done
		// with the inner placements
		case rockyRemain == 0 && innerBeltPlaced:
			innerIsFull = true
			continue
		}
	}

	//the inner system is full at this point. Anything that remains is for the outer system
	//so, work from the primary outward
	gasRemain := sys.GasGiantCount
	placeablesRemain := beltsRemain + rockyRemain + gasRemain
	currentAU = sys.Orbitals[0].OrbitalDistanceAU
	beltPlaceLastIteration := false
	loopCount = 0
	nextAU = 0
	for placeablesRemain > 0 {

		// crash prevent
		loopCount++
		if loopCount == 100 {
			ctx.Logger().Fatal().
				Float32("nextAU", nextAU).
				Int("beltsRemain", beltsRemain).
				Int("rockyRemain", rockyRemain).
				Int("gasRemain", gasRemain).
				Bool("beltPlaceLastIteration", beltPlaceLastIteration).
				Msg("Outer crash!")
		}

		//we place all remaining asteroid belts and rocky worlds first. This is harder than it seems because
		//we want belts to be intermixed with rocky planets if at all possible - but we may not have any rocky
		//planets left
		if rockyRemain > 0 || beltsRemain > 0 {

			nextAU = currentAU + minAUBandRockyWorldsOuter

			// make sure we are outside the habitable zone
			if nextAU < maxHabitalbleAU {
				currentAU = nextAU
				continue
			}

			switch {

			//both remain and we didn't just place a belt - 50/50 on what to place
			case rockyRemain > 0 && beltsRemain > 0 && !beltPlaceLastIteration:
				if dice.Dx(2) == 1 {
					belt := &model.SystemOrbital{}
					belt.IsAsteroid = true
					belt.OrbitalDistanceAU = nextAU
					belt.OrbitalNumber = ordinalUnset
					belt.Size = 0
					sys.Orbitals = append(sys.Orbitals, belt)

					//upkeep
					beltPlaceLastIteration = true
					beltsRemain -= 1
					currentAU = nextAU
					placeablesRemain = beltsRemain + rockyRemain + gasRemain
					continue
				}
				planet := &model.SystemOrbital{}
				planet.OrbitalDistanceAU = nextAU
				planet.OrbitalNumber = ordinalUnset
				planet.Size = calcRockyPlanetSize(ctx)
				sys.Orbitals = append(sys.Orbitals, planet)

				//upkeep
				beltPlaceLastIteration = false
				rockyRemain -= 1
				currentAU = nextAU
				placeablesRemain = beltsRemain + rockyRemain + gasRemain
				continue

			// both remain but we placed a belt last time - place a rocky
			case rockyRemain > 0 && beltsRemain > 0 && beltPlaceLastIteration:
				planet := &model.SystemOrbital{}
				planet.OrbitalDistanceAU = nextAU
				planet.OrbitalNumber = ordinalUnset
				planet.Size = calcRockyPlanetSize(ctx)
				sys.Orbitals = append(sys.Orbitals, planet)

				//upkeep
				beltPlaceLastIteration = false
				rockyRemain -= 1
				currentAU = nextAU
				placeablesRemain = beltsRemain + rockyRemain + gasRemain
				continue

			//only rocky left, place planet
			case rockyRemain > 0 && beltsRemain == 0:
				planet := &model.SystemOrbital{}
				planet.OrbitalDistanceAU = nextAU
				planet.OrbitalNumber = ordinalUnset
				planet.Size = calcRockyPlanetSize(ctx)
				sys.Orbitals = append(sys.Orbitals, planet)

				//upkeep
				beltPlaceLastIteration = false
				rockyRemain -= 1
				currentAU = nextAU
				placeablesRemain = beltsRemain + rockyRemain + gasRemain
				continue

			// we are out of rocky, must place belts - which is good b/c we didnt place one last time
			case rockyRemain == 0 && beltsRemain > 0 && !beltPlaceLastIteration:
				belt := &model.SystemOrbital{}
				belt.IsAsteroid = true
				belt.OrbitalDistanceAU = nextAU
				belt.OrbitalNumber = ordinalUnset
				belt.Size = 0
				sys.Orbitals = append(sys.Orbitals, belt)

				//upkeep
				beltPlaceLastIteration = true
				beltsRemain -= 1
				currentAU = nextAU
				placeablesRemain = beltsRemain + rockyRemain + gasRemain
				continue

			//we are out of rocky, must place belts  and we just placed a belt
			//need to invent a rocky planet out of cloth, add it, and let things run
			case rockyRemain == 0 && beltsRemain > 0 && beltPlaceLastIteration:
				planet := &model.SystemOrbital{}
				planet.OrbitalDistanceAU = nextAU
				planet.OrbitalNumber = ordinalUnset
				planet.Size = calcRockyPlanetSize(ctx)
				sys.Orbitals = append(sys.Orbitals, planet)

				//upkeep
				beltPlaceLastIteration = false
				currentAU = nextAU
				sys.TotalOrbitalCount += 1
				sys.RockyCount += 1
				placeablesRemain = beltsRemain + rockyRemain + gasRemain
				continue
			}
		}

		// we are on to gas giants - we can just place these
		if gasRemain > 0 {

			nextAU = currentAU + minAUBandGasGiants

			// gas giants can be farther apart. Add some variablility
			bump := dice.Roll()
			switch bump {
			case 1, 2, 3, 4:
				// do nothing
			case 5, 6:
				currentAU = nextAU + float32(dice.Dx(5))
				continue
			}

			gas := &model.SystemOrbital{}
			gas.OrbitalDistanceAU = nextAU
			gas.IsGasGiant = true
			gas.OrbitalNumber = ordinalUnset
			gas.Size = calcGasGiantSizeSize(ctx)
			sys.Orbitals = append(sys.Orbitals, gas)

			//upkeep
			beltPlaceLastIteration = false
			gasRemain -= 1
			currentAU = nextAU
			placeablesRemain = beltsRemain + rockyRemain + gasRemain
			continue
		}
	}

	//sort and add orbital numbers
	addSystemOrdinals(sys)

	//tag rocky planets
	addTagsToRockyPlanets(ctx, sys, systemTags)

	//determine which planets have moons
	addMoonsToPlanets(ctx, sys, systemTags)
}

func addTagsToRockyPlanets(ctx *util.TASContext, sys *model.SolarSystem, systemTags *model.SystemTags) {

	for _, orbital := range sys.Orbitals {
		// no need to add tags to these
		if orbital.IsAsteroid || orbital.IsGasGiant || orbital.IsPrimary {
			continue
		}

		var isCold, isHot, isStarNear bool
		switch {

		//if outside max hab zone, it is cold
		case orbital.OrbitalDistanceAU > maxHabitalbleAU:
			isCold = true

		//if within mercury-distance of the star, then near-star
		case orbital.OrbitalDistanceAU <= distanceToSunToBeConsideredNearStarAU:
			isStarNear = true
		//if inside the hab zone, then it is hot
		case orbital.OrbitalDistanceAU < minHabitableAU:
			isHot = true
		//in the hab zone
		default:
			// do nothing
		}

		orbital.Tags = generateTags(ctx, true, false, false, isHot, isCold, isStarNear, systemTags)
	}
}

func addMoonsToPlanets(ctx *util.TASContext, sys *model.SolarSystem, systemTags *model.SystemTags) {
	dice := ctx.Dice()

	for _, orbital := range sys.Orbitals {

		switch {

		//asteroids do not have moons
		case orbital.IsAsteroid:
			continue

		// gas giants have 1-10 moons each
		case orbital.IsGasGiant:
			moonCount := dice.Dx(10)
			allMoons := make([]model.Moon, 0, 1)
			for i := 0; i < moonCount; i++ {
				m := model.Moon{
					Size: dice.D3(), // the larger moons in our system are of this size at most
					Tags: generateTags(ctx, false, true, false, false, true, false, systemTags),
				}
				allMoons = append(allMoons, m)
			}
			orbital.Moons = allMoons

		//not a giant or asteroid, must be rocky
		default:

			// inner plannets cannot sustain moons as the star would strip them away
			if orbital.OrbitalDistanceAU < minHabitableAU {
				continue
			}

			// moon is cold if planet is outside habitable zone
			isCold := orbital.OrbitalDistanceAU >= maxHabitalbleAU
			allMoons := make([]model.Moon, 0, 1)
			if dice.Roll() >= 4 {
				m := model.Moon{
					Size: dice.Dx(2), // rocky planets have smaller moons than gas giants
					Tags: generateTags(ctx, false, false, true, false, isCold, false, systemTags),
				}
				allMoons = append(allMoons, m)
			}
			orbital.Moons = allMoons
		}
	}
}

func generateTags(ctx *util.TASContext, isRockyPlanet bool, isGasMoon bool, isRockyMoon bool, isHot bool, isCold bool, isStarNear bool, systemTags *model.SystemTags) []model.SystemTag {
	switch {
	case isRockyPlanet:
		return generateRockyPlanetTags(ctx, systemTags, isHot, isCold, isStarNear)
	case isRockyMoon: // inner planets cannot suppoort moons, the sun's gravity is too strong
		return generateRockyMoonTags(ctx, systemTags, isCold)
	case isGasMoon: //all gas moons are by defintion away from stars and cold
		return generateGasGiantMoonTags(ctx, systemTags)
	}
	return nil
}

func generateGasGiantMoonTags(ctx *util.TASContext, systemTags *model.SystemTags) []model.SystemTag {
	dice := ctx.Dice()
	tags := make([]model.SystemTag, 0, 1)
	tagtype := dice.Dx(8)

	switch {
	case tagtype <= 6: // general tag
		r := dice.Dx(systemTags.GeneralCount) - 1
		t := systemTags.GeneralTags[r]
		tags = append(tags, t)
	case tagtype == 7: // cold tag
		r := dice.Dx(systemTags.ColdTagsCount) - 1
		t := systemTags.ColdTags[r]
		tags = append(tags, t)
	case tagtype == 8: // gas-moon tag
		r := dice.Dx(systemTags.GasMoonTagsCount) - 1
		t := systemTags.GasMoonTags[r]
		tags = append(tags, t)
	}
	return tags
}

func generateRockyPlanetTags(ctx *util.TASContext, systemTags *model.SystemTags, isHot bool, isCold bool, isStarNear bool) []model.SystemTag {
	dice := ctx.Dice()
	tags := make([]model.SystemTag, 0, 1)

	switch {
	case isStarNear: // general tag
		tagtype := dice.Roll()
		switch {
		case tagtype <= 3: // hot tag
			r := dice.Dx(systemTags.HotTagsCount) - 1
			t := systemTags.HotTags[r]
			tags = append(tags, t)
		case tagtype >= 4: // star tag
			r := dice.Dx(systemTags.StarNearTagCount) - 1
			t := systemTags.StarNearTags[r]
			tags = append(tags, t)
		}
	case isHot: // add a hot and general tag
		r := dice.Dx(systemTags.HotTagsCount) - 1
		t := systemTags.HotTags[r]
		tags = append(tags, t)
		r = dice.Dx(systemTags.GeneralCount) - 1
		t = systemTags.GeneralTags[r]
		tags = append(tags, t)
	case isCold: // add a cold and general tag
		r := dice.Dx(systemTags.ColdTagsCount) - 1
		t := systemTags.ColdTags[r]
		tags = append(tags, t)
		r = dice.Dx(systemTags.GeneralCount) - 1
		t = systemTags.GeneralTags[r]
		tags = append(tags, t)
	}
	return tags
}

func generateRockyMoonTags(ctx *util.TASContext, systemTags *model.SystemTags, isCold bool) []model.SystemTag {
	dice := ctx.Dice()
	tags := make([]model.SystemTag, 0, 1)

	if isCold {
		if dice.Roll() >= 5 {
			r := dice.Dx(systemTags.ColdTagsCount) - 1
			t := systemTags.ColdTags[r]
			tags = append(tags, t)
		}
	}
	r := dice.Dx(systemTags.GeneralCount) - 1
	t := systemTags.GeneralTags[r]
	tags = append(tags, t)

	return tags
}

// determines how many rocky planets, gas giants and asteroid belts exist in addition to the primary system
func determineOtherOrbitalCounts(ctx *util.TASContext, totalOrbitals int, hasGiants bool) (int, int, int) {

	rockyCount := 0
	gasCount := 0
	asteroidCount := 0
	dice := ctx.Dice()

	//-1 here to account for the primary
	remainingOrbitals := totalOrbitals - 1

	//if giants are present, they can (by fiat) be no more than 40% of the planets
	if hasGiants {
		maxPossibleGiants := int(.4 * float32(remainingOrbitals))

		//possible b/c we are truncating. We need at least 1 giant since we know they are present
		if maxPossibleGiants == 0 {
			maxPossibleGiants = 1
		}

		//strictly random thereafter
		gasCount = dice.Dx(maxPossibleGiants)
		remainingOrbitals -= gasCount
	}
	if remainingOrbitals == 0 { //if we are out of orbitals, then we are done
		return rockyCount, gasCount, asteroidCount
	}

	//let's just treat asteroids as strictly linear. The +1 here gives us the ability to subtract 1
	//so a zero result is possible
	asteroidCount = dice.Dx(maxAsteroidBeltCount+1) - 1
	asteroidCount = util.BoundTo(asteroidCount, 0, remainingOrbitals)
	remainingOrbitals -= asteroidCount
	if remainingOrbitals == 0 { //if we are out of orbitals, then we are done
		return rockyCount, gasCount, asteroidCount
	}

	//anything left is a rocky planet
	rockyCount = remainingOrbitals

	return rockyCount, gasCount, asteroidCount
}

// ------------------------------------------------------
// generic helpers
// ------------------------------------------------------
func parseUWP(uwp string) (string, int, error) {
	matches := uwpRegExp.FindStringSubmatch(uwp)
	if len(matches) != 3 {
		return "", 0, fmt.Errorf("Invalid UWP: %s  Use '\"UNK 0000 B784886-6 RI\"' format", uwp)
	}

	// convert size to int
	var size int64
	if matches[2] == "A" {
		size = 10
	} else {
		size, _ = strconv.ParseInt(matches[2], 10, 16) // can ignore err b/c of regex
	}
	return matches[1], int(size), nil
}

func parseAU(au string) (float32, error) {
	v, err := strconv.ParseFloat(au, 32)
	if err != nil {
		return 0.0, err
	}

	if v < minHabitableAU || v > maxHabitalbleAU {
		return 0.0, fmt.Errorf("Primary AU is out of range: %s. Must between %f and %f", au, minHabitableAU, maxHabitalbleAU)
	}
	return float32(v), nil
}

func parseGasGiants(arg string) (bool, error) {
	la := strings.ToLower(arg)
	switch la {
	case "t", "true", "y", "yes":
		return true, nil

	case "f", "false", "n", "no":
		return false, nil
	}
	return false, fmt.Errorf("invalid truth value: %s", arg)
}

func determineTotalOrbitalCount(ctx *util.TASContext) int {
	// a bell curve of 1-16 planets w/ average of 8.5
	dice := ctx.Dice()
	r := dice.Dx(4) + dice.Dx(4) + dice.Dx(4) + dice.Dx(4) + dice.Dx(4)
	r = r - 4
	return r
}

func worldHasTradeCode(uwp string, re *regexp.Regexp) bool {
	return re.MatchString(uwp)
}

func calcRockyPlanetSize(ctx *util.TASContext) int {
	dice := ctx.Dice()
	roll := dice.Roll()
	return util.BoundTo(roll, 1, 5)
}

func calcGasGiantSizeSize(ctx *util.TASContext) int {
	dice := ctx.Dice()
	return dice.Sum(2) + 10
}

func calcColdAU(ctx *util.TASContext, baselineAU float32) float32 {
	dice := ctx.Dice()

	if dice.Dx(2) == 1 {
		return baselineAU + coldAUOffset
	}
	return baselineAU + coldAUOffset + 1.2
}

func addSystemOrdinals(sys *model.SolarSystem) {
	sort.SliceStable(sys.Orbitals, func(i, j int) bool {
		return sys.Orbitals[i].OrbitalDistanceAU < sys.Orbitals[j].OrbitalDistanceAU
	})

	for i := 1; i <= sys.TotalOrbitalCount; i++ {
		sys.Orbitals[i-1].OrbitalNumber = i
	}
}

func loadSystemTags(ctx *util.TASContext) (*model.SystemTags, error) {
	log := ctx.Logger()

	// load source data files
	log.Info().Msg("loading system tags file...")

	var sourceFiles = []string{systemTagsFile}
	fileData := util.IngestFiles("data/", sourceFiles)
	if !util.AllFilesReadOk(fileData) {
		log.Error().Msg("one or more files failed to load as expected")
		for _, f := range fileData {
			if !f.Ok() {
				log.Error().Err(f.Err).Str("filename", f.Name).Send()
			}
		}
		return nil, errors.New(h.UnableToContinueBecauseOfErrors)
	}
	log.Info().Msg("system tag data file load complete")

	fd := fileData[systemTagsFile]
	systemTags, err := model.SystemTagsFromFile(fd.Data)
	if err != nil {
		return nil, err
	}
	return systemTags, nil
}
