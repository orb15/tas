package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"tas/internal/util"
)

const (
	basicUWPRegExString = "^[ABCDEX]{1}[0-9A]{1}[0-9A-F]{1}[0-9A]{1}[0-9A-C]{1}[0-9A-F]{1}[0-9]{1}-[0-9A-F]{1}"
)

type CharacterDataType struct {
	HighestStewardLevel   int  `json:"highest-steward-skill"`
	HighestScoutNavalRank int  `json:"highest-scout-naval-rank"`
	HighestSocSkillDM     int  `json:"highest-soc-skill-dm"`
	ShipIsArmed           bool `json:"ship-is-armed"`
}

type WorldTradeInfoType struct {
	Name string `json:"name"`
	UWP  string `json:"uwp"`
}

type WorldTradeInfo struct {
	Population int
	Starport   string
	ZoneAmber  bool
	ZoneRed    bool
	TechLevel  int
	TradeCodes map[string]struct{}
}

type TradeFacts struct {
	CharacterData     *CharacterDataType    `json:"character-data"`
	RawWorldTradeInfo []*WorldTradeInfoType `json:"world-data"`
	WorldInfoMap      map[string]*WorldTradeInfo
	isValidated       bool
}

func TradeFactsFromFile(b []byte) (*TradeFacts, error) {

	var data TradeFacts
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (t *TradeFacts) Validate() []error {

	errs := make([]error, 0)

	//ensure all character data is reasonable
	sl := t.CharacterData.HighestStewardLevel
	if sl != -2 && !(sl >= 0 && sl <= 6) {
		errs = append(errs, fmt.Errorf("invalid steward skill level: %d", sl))
	}

	snr := t.CharacterData.HighestScoutNavalRank
	if !(snr >= 0 && snr <= 6) {
		errs = append(errs, fmt.Errorf("invalid scout or naval rank level: %d", snr))
	}

	sdm := t.CharacterData.HighestSocSkillDM
	if !(sdm >= -3 && sdm <= 3) {
		errs = append(errs, fmt.Errorf("invalid SOC skill DM: %d", sdm))
	}

	//ensure that all world names are unique and UWP are valid
	nameSet := make(map[string]struct{})

	basicPattern := regexp.MustCompile(basicUWPRegExString)

	for _, w := range t.RawWorldTradeInfo {
		nameSet[w.Name] = struct{}{}
		if !basicPattern.MatchString(w.UWP) {
			errs = append(errs, fmt.Errorf("world: %s has invalid basic UWP: %s", w.Name, w.UWP))
		}
	}

	if len(nameSet) != len(t.RawWorldTradeInfo) {
		errs = append(errs, fmt.Errorf("duplicate world name detected"))
	}

	if len(errs) == 0 {
		t.isValidated = true
	}

	return errs
}

// Parse returns true if []errors contains an actual error or false if []errors is nil or contains warnings we want to wrap like errors
func (t *TradeFacts) Parse() (bool, []error) {

	errs := make([]error, 0)

	//ensure validity before proceeding
	if !t.isValidated {
		errs = append(errs, fmt.Errorf("the UWP data cannot be parsed because it has not been Validated. This is a coding error"))
		return true, errs
	}

	t.WorldInfoMap = make(map[string]*WorldTradeInfo)

	for _, raw := range t.RawWorldTradeInfo {
		wi := &WorldTradeInfo{
			TradeCodes: map[string]struct{}{},
		}

		//starport is always the first value
		wi.Starport = string(raw.UWP[0])

		//population is always the char at index 4 in the basic UWP
		popString := raw.UWP[4]
		pop, _ := util.HexAsInt(string(popString)) //no error check here as regex validated this
		wi.Population = pop

		//tech level is always the char at index 8 in the basic UWP
		techString := raw.UWP[8]
		tech, _ := util.HexAsInt(string(techString)) //no error check here as regex validated this
		wi.TechLevel = tech

		//clean up the raw UWP as much as possible then break it up by spaces
		rawSlices := strings.Split(strings.TrimSpace(raw.UWP), " ")

		//easy option -  no bases, trade codes or zones
		if len(rawSlices) == 1 {
			wi.ZoneAmber = false
			wi.ZoneRed = false
			wi.TradeCodes = make(map[string]struct{})
			t.WorldInfoMap[raw.Name] = wi
			continue
		}

		//begin parsing the remaining elements in the slice and try to determine what we have
		//element 0 is the base UWP and does not matter at this point
		for i := 1; i < len(rawSlices); i++ {

			thisElement := strings.TrimSpace(rawSlices[i])

			switch len(thisElement) {

			case 0: //the user placed an extranious space in the file (probably should not happen with TrimSpace above, just being careful)
				errs = append(errs, fmt.Errorf("world: %s has extranious space in UWP: %s", raw.Name, raw.UWP))

			case 1: //found char of length 1 - could be a base code or trade code or an error

				//on last element
				if i == len(rawSlices)-1 {
					switch thisElement {
					case "A": //amber travel zone
						wi.ZoneAmber = true
					case "R": //red travel zone
						wi.ZoneRed = true
					case "G": //green designation - non-srtandard but will warn and move on
						errs = append(errs, fmt.Errorf("world: %s has unneeded Green Zone designator in UWP travel zone: %s", raw.Name, thisElement))
					case "S", "C", "N", "M", "D", "W": //this is a base designation, ok to ignore
						break
					default: //unknown last char in last position
						errs = append(errs, fmt.Errorf("world: %s has invalid value in UWP travel zone: %s", raw.Name, thisElement))
						return true, errs
					}
				} else { //we found a len 1 string not at last location - this should not happen
					errs = append(errs, fmt.Errorf("world: %s has invalid value in UWP: %s", raw.Name, thisElement))
					return true, errs
				}

			case 3, 4, 5, 6: //this must be a set of base codes, at least, we are going to treat them as that
				errs = append(errs, fmt.Errorf("world: %s has UWP with entry '%s', which we are treating as a Base designation", raw.Name, thisElement))

			case 2: //this is the most interesting as it could be a pair of base codes or an expected 2-letter trade code

				if _, ok := util.ValidTradeCodeSet[thisElement]; ok { //we recognize this pair as a trade code
					if _, exists := wi.TradeCodes[thisElement]; exists {
						errs = append(errs, fmt.Errorf("world: %s has redundant trade code: %s", raw.Name, thisElement))
					}
					wi.TradeCodes[thisElement] = struct{}{}
				} else { //the pair is not a trade code, we will assume it is a base code
					errs = append(errs, fmt.Errorf("world: %s has UWP with entry '%s', which we are treating as a Base designation", raw.Name, thisElement))
				}

			default: //we have more than 6 characters, that is an error
				errs = append(errs, fmt.Errorf("world: %s has UWP with entry '%s', which looks garbled. Something is wrong", raw.Name, thisElement))
				return true, errs
			}
		}

		t.WorldInfoMap[raw.Name] = wi
	}

	return false, errs
}

func (t *TradeFacts) DataForWorldName(name string) (*WorldTradeInfo, bool) {
	data := t.WorldInfoMap[name]
	if data == nil {
		return nil, false
	}
	return data, true
}
