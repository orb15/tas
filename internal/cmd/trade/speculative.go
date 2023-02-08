package trade

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

const (
	assumedOpposingBrokerSkill        = 2
	impossiblyLowModifierForTradeCode = -10
)

var SpecTradeCmdConfig = &cobra.Command{

	Use:   "spec",
	Short: "determines speculative trade modifiers and other trade-related information for a given world",
	Run:   specTradeCmd,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("exactly 2 arguments required - the current world name, and and either 'buy' or 'sell")
		}

		operation := args[1]
		operation = strings.ToLower(strings.TrimSpace(operation))
		if operation != "buy" && operation != "sell" {
			return fmt.Errorf("second argument must be 'buy' or 'sell'")
		}
		return nil
	},
}

func specTradeCmd(cmd *cobra.Command, args []string) {

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

	//load the data we need to build speculative trade data
	tradeFacts, tradeGoodsMap, err := LoadSpeculativeTradeFacts(ctx)
	if err != nil {
		return
	}

	//fetch the arguments (local world name)  and operation and do the generation
	localWorldName := args[0]
	isBuying := true
	if args[1] == "sell" {
		isBuying = false
	}

	localData, ok := tradeFacts.DataForWorldName(localWorldName)
	if !ok {
		err := fmt.Errorf("the local world: %s is not defined in the trade data file", localWorldName)
		log.Error().Err(err).Msg("unable to generate trade data")
		return
	}

	log.Debug().Bool("isBuying", isBuying).Str("world", localWorldName).Send()

	summary := GenerateSpeculativeTrade(ctx, localData, tradeGoodsMap, isBuying)
	summary.WorldName = localWorldName
	writeSpeculativeOutput(ctx, summary, isBuying)
}

func GenerateSpeculativeTrade(ctx *util.TASContext, localData *model.WorldTradeInfo, tradeGoodsMap model.TradeGoodsMap, isBuying bool) model.SpeculativeTradeSummary {
	log := ctx.Logger()

	log.Info().Msg("Beginning speculative trade generation...")

	//calc DM to find a supplier or broker
	findSupplierBrokerDM := 0
	switch localData.Starport {
	case "A":
		findSupplierBrokerDM = 6
	case "B":
		findSupplierBrokerDM = 4
	case "C":
		findSupplierBrokerDM = 2
	}

	//notes
	var notes = []string{"see pg 241.",
		"Choose an option at bottom of pg 241 to find a Supplier or Broker. Use provided DM for this check. Depending on method used, this takes some time.",
		"Dealing with a Broker adds a flat DM+2 to all price rolls plus uses the Broker's Broker or Streetwise skill (2D/3) instead of the players', but costs 10-20% of the total order price",
		"For each Lot of Trade Goods, determine a Purchase Price by: using the DM and Base Price provided for that lot, the players' (or Broker's) Broker skill and a roll of 3D. Consult table pg 243",
		"The same process must be done for selling goods to determine the price a purchaser is willing to pay",
		"In the case of selling, the DM for every possible trade good on the current world is provided as we don't know what is being sold",
		"Illegal Goods may be listed but are only available to be bought or sold via roleplay and/or a local Broker",
	}

	summary := model.SpeculativeTradeSummary{
		FindSupplierOrBrokerDM: findSupplierBrokerDM,
		TradeNotes:             notes,
	}

	//if we are buying, we need to generate goods available on this planet
	if isBuying {
		summary.TransactionType = "buy"
		summary.TradeLots = generateTradeLots(ctx, localData, tradeGoodsMap, isBuying)
	} else { //we are selling, so we don't need to generate goods, just list all DMs for any type of good that _might_ be sold - which is all of them!
		summary.TransactionType = "sell"
		summary.TradeLots = buildSellersDMs(ctx, localData, tradeGoodsMap)
	}

	log.Info().Msg("Speculative trade generation complete")
	return summary
}

func writeSpeculativeOutput(ctx *util.TASContext, summary model.SpeculativeTradeSummary, isBuying bool) {
	var sb strings.Builder

	if isBuying {

		sb.WriteString("Speculative Trade Offerings - Characters are Purchaing Goods")

		sb.WriteString(h.NL)
		sb.WriteString(h.NL + "DM to Find Supplier or Broker to Aid in Purchase:" + h.SP + fmt.Sprintf("%d", summary.FindSupplierOrBrokerDM))
		sb.WriteString(h.NL)
		sb.WriteString(h.NL + "Trade Lots Available For Purchase")
		for _, l := range summary.TradeLots {
			sb.WriteString(h.NL + h.TAB + "Lot ID:" + h.SP + fmt.Sprintf("%d", l.LotId))
			sb.WriteString(h.NL + h.TAB + h.TAB + "Type of Goods:" + h.SP + l.Type)
			sb.WriteString(h.NL + h.TAB + h.TAB + "Example Goods:" + h.SP + l.Example)
			sb.WriteString(h.NL + h.TAB + h.TAB + "Base Price:" + h.SP + fmt.Sprintf("%d", l.BasePrice))
			sb.WriteString(h.NL + h.TAB + h.TAB + "Tons Available:" + h.SP + fmt.Sprintf("%d", l.TonsAvail))
			sb.WriteString(h.NL + h.TAB + h.TAB + "Purchase Price DM:" + h.SP + fmt.Sprintf("%d", l.OfferPriceDM))
		}
	} else {
		sb.WriteString("Speculative Trade Offerings - Characters are Selling Goods")

		sb.WriteString(h.NL)
		sb.WriteString(h.NL + "DM to Find Supplier or Broker to Aid in Sale:" + h.SP + fmt.Sprintf("%d", summary.FindSupplierOrBrokerDM))
		sb.WriteString(h.NL)
		sb.WriteString(h.NL + "Sale of Goods Owned - Trade Table")
		for _, l := range summary.TradeLots {
			sb.WriteString(h.NL + h.TAB + "Trade Identifier:" + h.SP + fmt.Sprintf("%d", l.LotId))
			sb.WriteString(h.NL + h.TAB + h.TAB + "Type of Goods:" + h.SP + l.Type)
			sb.WriteString(h.NL + h.TAB + h.TAB + "Base Price:" + h.SP + fmt.Sprintf("%d", l.BasePrice))
			sb.WriteString(h.NL + h.TAB + h.TAB + "Sale Price DM:" + h.SP + fmt.Sprintf("%d", l.OfferPriceDM))
		}
	}
	sb.WriteString(h.NL)
	for _, sn := range summary.TradeNotes {
		sb.WriteString(h.NL + sn)
	}

	fmt.Println(sb.String())

	//also write to file if requested
	writeToFile, _ := ctx.Config().Flags.GetBool(util.ToFileFlagName)
	if writeToFile {
		h.WrappedJSONFileWriter(ctx, summary, summary.ToFileName())
	}
}

func generateTradeLots(ctx *util.TASContext, localData *model.WorldTradeInfo, tradeGoodsMap model.TradeGoodsMap, isBuying bool) []*model.SpeculativeTradeLot {

	log := ctx.Logger()
	dice := ctx.Dice()
	tradeLots := make([]model.SpeculativeTradeLot, 0)

	//determine goods available DM
	availabilityDM := 0
	if localData.Population <= 3 {
		availabilityDM = -3
	}
	if localData.Population >= 9 {
		availabilityDM = 3
	}

	lotID := 0

	//tradeGoodsMap is the data on pgs 243, 244

	//First Step: Look through this map and generate a (potential) Lot for all Common Goods
	//and each Advanced and Illegal Good that applies per the world's trade codes
	log.Debug().Msg("starting first pass lot creation")
	for _, dataRow := range tradeGoodsMap {
		log.Debug().Str("type", dataRow.Type).Msg("attempting creation of a new trade lot")
		newLot, success := buildTradeLot(ctx, availabilityDM, localData, dataRow, true, isBuying)

		if success { //success will be false when the planets codes are not applicable to the trade good or the quantity is 0 or less
			log.Debug().Msg("creation of a first-pass lot succeeded")
			lotID += 1
			newLot.LotId = lotID
			tradeLots = append(tradeLots, newLot)
		} else {
			log.Debug().Msg("creation of a first-pass lot failed")
		}
	}

	//Second Step: the world will have a number of additional, random goods even if they dont qualify for them
	//we just pick a number of items from the table at random = Population value
	log.Debug().Msg("starting second pass lot creation")
	for i := 0; i < localData.Population; i++ {

		dataRow := tradeGoodsMap[dice.D66()]
		newLot, success := buildTradeLot(ctx, availabilityDM, localData, dataRow, false, isBuying)

		if success { //success will be false when the quantity is 0 or less
			log.Debug().Str("type", dataRow.Type).Msg("creation of a second-pass lot succeeded")
			lotID += 1
			newLot.LotId = lotID
			tradeLots = append(tradeLots, newLot)
		} else {
			log.Debug().Str("type", dataRow.Type).Msg("creation of a second-pass lot failed")
		}
	}

	//consolidate the trade lots - lots of the same goods need to be combined into larger lots
	consolidatedLots := combineTradeLots(ctx, tradeLots)

	return consolidatedLots
}

func combineTradeLots(ctx *util.TASContext, rawLots []model.SpeculativeTradeLot) []*model.SpeculativeTradeLot {

	log := ctx.Logger()
	log.Debug().Msg("combining trade lots")
	combinedLots := make([]*model.SpeculativeTradeLot, 0, len(rawLots))

	log.Debug().Int("raw-lot-count", len(rawLots)).Msg("lots before consolidation")

	//move raw lots into a map, combining lots with same Type value
	lotMap := make(map[string]model.SpeculativeTradeLot)

	for _, lot := range rawLots {

		existingLot, exists := lotMap[lot.Type]
		if !exists {
			lotMap[lot.Type] = lot
		} else {
			existingLot.TonsAvail += lot.TonsAvail
			lotMap[lot.Type] = existingLot
		}
	}

	for _, lot := range lotMap {
		lotref := lot
		combinedLots = append(combinedLots, &lotref)
	}

	//sort these by lot id
	sortLotsByLotId(combinedLots)

	log.Debug().Int("combined-lot-count", len(combinedLots)).Msg("lots after consolidation")

	return combinedLots
}

func buildTradeLot(ctx *util.TASContext, availabilityDM int, localData *model.WorldTradeInfo, dataRow *model.TradeGood, mustQualify bool, isBuying bool) (model.SpeculativeTradeLot, bool) {
	newLot := model.SpeculativeTradeLot{}

	log := ctx.Logger()
	dice := ctx.Dice()

	//quick check - determine availability. No sense doing a bunch of thinking if the availability is zero
	qty := dice.Sum(dataRow.TonsDice) - availabilityDM
	if qty <= 0 {
		log.Debug().Str("good-type", dataRow.Type).Msg("lot has insufficient quantity")
		return newLot, false
	}

	//so something is available...

	//common good, always avail on every world
	if dataRow.Value >= 11 && dataRow.Value <= 16 { //common good, always avail on every world
		newLot.BasePrice = dataRow.BasePrice
		newLot.Example = dataRow.Examples
		newLot.Type = dataRow.Type
		newLot.TonsAvail = qty * dataRow.TonsMultiplier
		newLot.OfferPriceDM = calculatePriceDM(localData, dataRow, isBuying)

		log.Debug().Str("good-type", dataRow.Type).Msg("created a new common lot")

		return newLot, true
	}

	//need to check availability - does this world match a good's availability?
	match := false
	for wtc := range localData.TradeCodes {
		for _, rav := range dataRow.Availability {
			if wtc == rav {
				log.Debug().Str("good-type", dataRow.Type).Str("world-tag", wtc).Msg("found a matching world tag and availability tag")
				match = true
				break
			}
		}
		if match {
			break
		} else {
			log.Debug().Str("good-type", dataRow.Type).Str("world-tag", wtc).Msg("did not match any tags")
		}
	}

	//the world has a trade code also listed in the data row we are looking at or we dont care
	//because we are on Phase 2 of goods generation and all goods are acceptible in this phase
	if match || !mustQualify {
		newLot.BasePrice = dataRow.BasePrice
		newLot.Example = dataRow.Examples
		newLot.Type = dataRow.Type
		newLot.TonsAvail = qty * dataRow.TonsMultiplier
		newLot.OfferPriceDM = calculatePriceDM(localData, dataRow, true)
		log.Debug().Str("good-type", dataRow.Type).Msg("created a new advanced or illegal lot")

		return newLot, true
	}

	return newLot, false
}

func calculatePriceDM(localData *model.WorldTradeInfo, dataRow *model.TradeGood, isBuying bool) int {

	//we start by assuming that the opposing business has a 'skill 2' Broker (or Streetwise in the case of illegal goods)
	priceDM := assumedOpposingBrokerSkill

	//fetch world trade codes, augmented by Travel Zone information
	worldTradeCodes := localData.TradeCodes
	if localData.ZoneAmber {
		worldTradeCodes["Amber Zone"] = struct{}{}
	}
	if localData.ZoneRed {
		worldTradeCodes["Red Zone"] = struct{}{}
	}

	//determine highest purchase price DM for world trade code. -10 here allows for the highest offset to be lower than 0
	highestPurchaseDM := impossiblyLowModifierForTradeCode
	for _, ptc := range dataRow.PurchaseDMs {
		if _, listed := worldTradeCodes[ptc.Code]; listed {
			highestPurchaseDM = h.MaxInt(highestPurchaseDM, ptc.Mod)
		}
	}

	//now that we fetched the highest value, handle the case where the trade code wasn't listed
	//if we dont do this, we would end up applying this huge negative mod when the world didnt match any of the trade codes in the table
	if highestPurchaseDM == impossiblyLowModifierForTradeCode {
		highestPurchaseDM = 0
	}

	//determine highest sale price DM for world trade code. -10 here allows for the highest offset to be lower than 0
	highestSaleDM := impossiblyLowModifierForTradeCode
	for _, ptc := range dataRow.PurchaseDMs {
		if _, listed := worldTradeCodes[ptc.Code]; listed {
			highestSaleDM = h.MaxInt(highestSaleDM, ptc.Mod)
		}
	}

	if highestSaleDM == impossiblyLowModifierForTradeCode {
		highestSaleDM = 0
	}

	//offset one mod against the other to get the net effect of buying on or selling to this world for the given product
	if isBuying {
		priceDM = priceDM + highestPurchaseDM - highestSaleDM
	} else {
		priceDM = priceDM + highestSaleDM - highestPurchaseDM
	}

	return priceDM
}

func buildSellersDMs(ctx *util.TASContext, localData *model.WorldTradeInfo, tradeGoodsMap model.TradeGoodsMap) []*model.SpeculativeTradeLot {

	purchaseDMs := make([]*model.SpeculativeTradeLot, 0)

	//calculate a DM for every row in the trade table on page 244 & 245. We don't know what goods the seller has, so give them a DM for every possible
	//good they can be carrying. We don;t need to concern ourselves with availability or "appropriateness"
	for _, dataRow := range tradeGoodsMap {
		lot := &model.SpeculativeTradeLot{
			LotId:        dataRow.Value, //we don;t really care about lot Ids, so use d66 value to make it easy to match results
			Type:         dataRow.Type,
			BasePrice:    dataRow.BasePrice,
			OfferPriceDM: calculatePriceDM(localData, dataRow, false),
		}
		purchaseDMs = append(purchaseDMs, lot)
	}

	//sort these by D66 identity to make it easy to find the applicable info
	sortLotsByLotId(purchaseDMs)

	return purchaseDMs
}

func sortLotsByLotId(tradeLots []*model.SpeculativeTradeLot) {
	sort.Slice(tradeLots, func(i, j int) bool {
		return tradeLots[i].LotId <= tradeLots[j].LotId
	})
}

func LoadSpeculativeTradeFacts(ctx *util.TASContext) (*model.TradeFacts, model.TradeGoodsMap, error) {
	log := ctx.Logger()

	//determine name of trade facts file from flags
	tradeDataFilename, err := ctx.Config().Flags.GetString(TradeFileFlagName)
	if err != nil {
		tradeDataFilename = defaultTradeDataFilename
	}

	// load source data files
	log.Info().Msg("loading trade data files")

	tradeDataFilenameWithPath := "data-local/" + tradeDataFilename
	tradeGoodFilenameWithPath := "data/" + tradeGoodFilename

	var sourceFiles = []string{tradeDataFilenameWithPath, tradeGoodFilenameWithPath}

	fileData := util.IngestFiles("", sourceFiles)
	if !util.AllFilesReadOk(fileData) {
		log.Error().Msg("one or more files failed to load as expected")
		for _, f := range fileData {
			if !f.Ok() {
				log.Error().Err(f.Err).Str("filename", f.Name).Send()
			}
		}
		return nil, nil, errors.New(h.UnableToContinueBecauseOfErrors)
	}
	log.Info().Msg("trade data files load complete")

	//parse trade data
	log.Info().Msg("parsing trade data files...")
	tradeFacts := &model.TradeFacts{}
	tradeGoods := model.TradeGoodsMap{}

	for filename, fd := range fileData {

		switch filename {

		case tradeDataFilenameWithPath:
			tradeFacts, err = model.TradeFactsFromFile(fd.Data)
			if err != nil {
				return nil, nil, err
			}

		case tradeGoodFilenameWithPath:
			tradeGoods, err = model.TradeGoodsFromFile(fd.Data)
			if err != nil {
				return nil, nil, err
			}

		}
	}

	//ensure the trade data from the file is reasonable
	errs := tradeFacts.Validate()
	if len(errs) > 0 {
		log.Error().Msg("trade data from file is not valid. See the following lines for more information")
		for _, e := range errs {
			log.Error().Err(e).Send()
		}
		return nil, nil, errors.New(h.UnableToContinueBecauseOfErrors)
	}

	//parse the trade data and do further, detailed validation
	hasErrors, errs := tradeFacts.Parse()
	if hasErrors { //Parse truly failed with real errors
		log.Error().Msg("parse of world UWP data failed. The following lines should point you to the problem")
		for _, e := range errs {
			log.Error().Err(e).Send()
		}
		return nil, nil, errors.New(h.UnableToContinueBecauseOfErrors)
	} else { //Parse data did not truiely have errors but did have warnings we need to tell the user about
		log.Warn().Msg("parse of the UWP's succeeded but some assumptions were made about the data. The following lines provide more info, please check them carefully")
		for _, e := range errs {
			log.Warn().Err(e).Send()
		}
	}

	log.Info().Msg("parsing trade data files complete")
	return tradeFacts, tradeGoods, nil
}
