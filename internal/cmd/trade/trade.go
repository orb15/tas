package trade

import (
	"errors"
	"fmt"
	"strings"

	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"

	"github.com/spf13/cobra"
)

const (
	defaultTradeDataFilename = "trade-data.json"
	tradeGoodFilename        = "trade-goods.json"

	TradeFileFlagName = "file"
)

var TradeCmdConfig = &cobra.Command{

	Use:   "trade",
	Short: "determines trade modifiers and other trade-related information",
	Run:   tradeCmd,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("exactly 2 arguments required - the source and destination world names")
		}
		return nil
	},
}

func tradeCmd(cmd *cobra.Command, args []string) {

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

	//load the data we need to build standard trade data
	tradeFacts, err := LoadStandardTradeFacts(ctx)
	if err != nil {
		return
	}

	//fetch the arguments (from/to world names) and do the generation
	from := args[0]
	to := args[1]

	summary, err := GenerateStandardTrade(ctx, from, to, tradeFacts)
	if err != nil {
		log.Error().Err(err).Msg("trade DM calculations failed")
		return
	}

	writeStandardOutput(ctx, summary)

}

func GenerateStandardTrade(ctx *util.TASContext, from string, to string, tradeFacts *model.TradeFacts) (*model.StandardTradeModifiers, error) {

	log := ctx.Logger()
	log.Info().Msg("Beginning standard trade generation...")

	//be sure the from and to worlds exist
	fromData, ok := tradeFacts.DataForWorldName(from)
	if !ok {
		err := fmt.Errorf("the origin/from world: %s is not defined in the trade data file", from)
		return nil, err
	}

	toData, ok := tradeFacts.DataForWorldName(to)
	if !ok {
		err := fmt.Errorf("the destination/to world: %s is not defined in the trade data file", from)
		return nil, err
	}

	alltrade := &model.StandardTradeModifiers{
		PassengerTrade: generatePassengers(ctx, fromData, toData, tradeFacts),
	}

	//handle freight next, as we need the core DM to do Mail
	coreFreightDM, freightSummary := generateFreight(ctx, fromData, toData, tradeFacts)
	alltrade.FreightTrade = freightSummary

	//mail
	alltrade.MailTrade = generateMail(ctx, fromData, coreFreightDM, tradeFacts)

	log.Info().Msg("All standard trade generation complete")
	return alltrade, nil
}

func writeStandardOutput(ctx *util.TASContext, summary *model.StandardTradeModifiers) {
	var sb strings.Builder

	sb.WriteString("Standard Trade Offerings")

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Passenger Trade")
	for _, p := range summary.PassengerTrade.PassengerDMs {
		sb.WriteString(h.NL + h.TAB + "Type of Passage:" + h.SP + p.PassageType)
		sb.WriteString(h.NL + h.TAB + h.TAB + "DM:" + h.SP + fmt.Sprintf("%d", p.DM))
		sb.WriteString(h.NL + h.TAB + h.TAB + "Requirements:" + h.SP + p.Requirements)
	}
	sb.WriteString(h.NL)
	for _, pn := range summary.PassengerTrade.PassengerNotes {
		sb.WriteString(h.NL + h.TAB + pn)
	}

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Freight Trade")
	for _, f := range summary.FreightTrade.FreightDMs {
		sb.WriteString(h.NL + h.TAB + "Cargo Type:" + h.SP + f.LotType)
		sb.WriteString(h.NL + h.TAB + h.TAB + "DM:" + h.SP + fmt.Sprintf("%d", f.DM))
	}
	sb.WriteString(h.NL)
	for _, fn := range summary.FreightTrade.FreightNotes {
		sb.WriteString(h.NL + h.TAB + fn)
	}

	sb.WriteString(h.NL)
	sb.WriteString(h.NL + "Mail Service")
	sb.WriteString(h.NL + h.TAB + h.TAB + "DM:" + h.SP + fmt.Sprintf("%d", summary.MailTrade.MailDM))
	sb.WriteString(h.NL + h.TAB + h.TAB + "Lots Available:" + h.SP + fmt.Sprintf("%d", summary.MailTrade.LotsAvail))
	sb.WriteString(h.NL)
	for _, mn := range summary.MailTrade.MailNotes {
		sb.WriteString(h.NL + h.TAB + mn)
	}

	fmt.Println(sb.String())
}

func generateMail(ctx *util.TASContext, fromData *model.WorldTradeInfo, coreFreightDM int, tradeFacts *model.TradeFacts) model.MailTradeSummary {
	log := ctx.Logger()
	dice := ctx.Dice()

	mailDM := tradeFacts.CharacterData.HighestScoutNavalRank + tradeFacts.CharacterData.HighestSocSkillDM

	log.Info().Msg("Beginning mail generation...")

	//freight core DM
	mailDM = h.AdjustDM(ctx, mailDM, -2, coreFreightDM, h.LE, -10)
	mailDM = h.AdjustDM(ctx, mailDM, -1, coreFreightDM, h.INR, -9, -5)
	mailDM = h.AdjustDM(ctx, mailDM, 0, coreFreightDM, h.INR, -4, 4)
	mailDM = h.AdjustDM(ctx, mailDM, 1, coreFreightDM, h.INR, 5, 9)
	mailDM = h.AdjustDM(ctx, mailDM, 2, coreFreightDM, h.GT, 10)

	//tech level
	mailDM = h.AdjustDM(ctx, mailDM, -4, fromData.TechLevel, h.LE, 5)

	//armed ship
	mailDM = h.AdjustZoneDM(mailDM, 2, tradeFacts.CharacterData.ShipIsArmed)

	var notes = []string{"see pg 241.",
		"Roll 2D + mail DM. On 12+, mail is entrusted to the ship and crew.",
		"Each lot is 5T and all lots must be taken or none.",
		"Each lot is paid ar Cr25000 on delivery.",
		"There is no price adjustment for ditance",
	}

	summary := model.MailTradeSummary{
		MailDM:    mailDM,
		LotsAvail: dice.Roll(),
		MailNotes: notes,
	}

	log.Info().Msg("Mail generation complete")
	return summary
}

func generateFreight(ctx *util.TASContext, fromData *model.WorldTradeInfo, toData *model.WorldTradeInfo, tradeFacts *model.TradeFacts) (int, model.FreightTradeSummary) {
	log := ctx.Logger()

	freights := make([]model.FreightDM, 0, 3) //3 -> major, minor, incidental

	log.Info().Msg("Beginning freight generation...")

	//DMs that apply to all freight lots
	coreDM := 0

	//starport - cannot use adjustDM because starport is a string, even generics wont help here
	coreDM = h.AdjustStarportDM(coreDM, 2, fromData.Starport, "A")
	coreDM = h.AdjustStarportDM(coreDM, 2, toData.Starport, "A")
	coreDM = h.AdjustStarportDM(coreDM, 1, fromData.Starport, "B")
	coreDM = h.AdjustStarportDM(coreDM, 1, toData.Starport, "B")
	coreDM = h.AdjustStarportDM(coreDM, -1, fromData.Starport, "E")
	coreDM = h.AdjustStarportDM(coreDM, -1, toData.Starport, "E")
	coreDM = h.AdjustStarportDM(coreDM, -3, fromData.Starport, "X")
	coreDM = h.AdjustStarportDM(coreDM, -3, toData.Starport, "X")

	//travel zone - same issue with bools
	coreDM = h.AdjustZoneDM(coreDM, -2, fromData.ZoneAmber)
	coreDM = h.AdjustZoneDM(coreDM, -2, toData.ZoneAmber)
	coreDM = h.AdjustZoneDM(coreDM, -6, fromData.ZoneRed)
	coreDM = h.AdjustZoneDM(coreDM, -6, toData.ZoneRed)

	//population
	coreDM = h.AdjustDM(ctx, coreDM, -4, fromData.Population, h.LE, 1)
	coreDM = h.AdjustDM(ctx, coreDM, -4, toData.Population, h.LE, 1)
	coreDM = h.AdjustDM(ctx, coreDM, 2, fromData.Population, h.INR, 6, 7)
	coreDM = h.AdjustDM(ctx, coreDM, 2, toData.Population, h.INR, 6, 7)
	coreDM = h.AdjustDM(ctx, coreDM, 4, fromData.Population, h.GE, 8)
	coreDM = h.AdjustDM(ctx, coreDM, 4, toData.Population, h.GE, 8)

	//tech level
	coreDM = h.AdjustDM(ctx, coreDM, -1, fromData.TechLevel, h.LE, 6)
	coreDM = h.AdjustDM(ctx, coreDM, -1, toData.TechLevel, h.LE, 6)
	coreDM = h.AdjustDM(ctx, coreDM, 2, fromData.TechLevel, h.GE, 9)
	coreDM = h.AdjustDM(ctx, coreDM, 2, toData.TechLevel, h.GE, 9)

	//major cargo
	f := model.FreightDM{
		LotType: "major",
		DM:      coreDM - 4,
	}
	freights = append(freights, f)

	//minor cargo
	f = model.FreightDM{
		LotType: "minor",
		DM:      coreDM,
	}
	freights = append(freights, f)

	//incidental cargo
	f = model.FreightDM{
		LotType: "incidental",
		DM:      coreDM + 2,
	}
	freights = append(freights, f)

	var notes = []string{"see pg 240.",
		"To the given DM, add the Effect of an (8+) Broker or Streetwise check.",
		"Use DM -1 for each parsec beyond 1 between source and destination worlds",
		"Then roll on Freight Traffic table using the calculated DM to determine number of lots available in each catagory. Lots are all-or-nothing, pay on delivery.",
		"There is a penalty for late arrival.",
		"Freight is almost worthless in terms of its value, so absconding with Freight and not delivering it is worth less than delivering it!"}

	summary := model.FreightTradeSummary{
		FreightDMs:   freights,
		FreightNotes: notes,
	}

	log.Info().Msg("Freight generation complete")
	return coreDM, summary
}

func generatePassengers(ctx *util.TASContext, fromData *model.WorldTradeInfo, toData *model.WorldTradeInfo, tradeFacts *model.TradeFacts) model.PassengerTradeSummary {

	log := ctx.Logger()

	passengers := make([]model.PassengerDM, 0, 4) //4 -> high, middle, basic, low

	log.Info().Msg("Beginning passenger generation...")

	//DMs that apply to all berths
	coreDM := tradeFacts.CharacterData.HighestStewardLevel

	//starport - cannot use adjustDM because starport is a string, even generics wont help here
	coreDM = h.AdjustStarportDM(coreDM, 2, fromData.Starport, "A")
	coreDM = h.AdjustStarportDM(coreDM, 2, toData.Starport, "A")
	coreDM = h.AdjustStarportDM(coreDM, 1, fromData.Starport, "B")
	coreDM = h.AdjustStarportDM(coreDM, 1, toData.Starport, "B")
	coreDM = h.AdjustStarportDM(coreDM, -1, fromData.Starport, "E")
	coreDM = h.AdjustStarportDM(coreDM, -1, toData.Starport, "E")
	coreDM = h.AdjustStarportDM(coreDM, -3, fromData.Starport, "X")
	coreDM = h.AdjustStarportDM(coreDM, -3, toData.Starport, "X")

	//travel zone - same issue with bools
	coreDM = h.AdjustZoneDM(coreDM, 1, fromData.ZoneAmber)
	coreDM = h.AdjustZoneDM(coreDM, 1, toData.ZoneAmber)
	coreDM = h.AdjustZoneDM(coreDM, -4, fromData.ZoneRed)
	coreDM = h.AdjustZoneDM(coreDM, -4, toData.ZoneRed)

	//population
	coreDM = h.AdjustDM(ctx, coreDM, -4, fromData.Population, h.LE, 1)
	coreDM = h.AdjustDM(ctx, coreDM, -4, toData.Population, h.LE, 1)
	coreDM = h.AdjustDM(ctx, coreDM, 1, fromData.Population, h.INR, 6, 7)
	coreDM = h.AdjustDM(ctx, coreDM, 3, fromData.Population, h.GE, 8)
	coreDM = h.AdjustDM(ctx, coreDM, 3, toData.Population, h.GE, 8)

	//High Passengers
	p := model.PassengerDM{
		PassageType:  "high",
		DM:           coreDM - 4,
		Requirements: "Per passenger: 1 stateroom and 1T cargo space. 1 dedicated steward per 10 passengers (round up)",
	}
	passengers = append(passengers, p)

	//Middle Passengers
	p = model.PassengerDM{
		PassageType:  "middle",
		DM:           coreDM,
		Requirements: "Per passenger: 1 stateroom and 100kg cargo space. 1 dedicated steward per 100 passengers (round up)",
	}
	passengers = append(passengers, p)

	//Basic Passengers
	p = model.PassengerDM{
		PassageType:  "basic",
		DM:           coreDM,
		Requirements: "Per passenger: 0.5 stateroom and 10kg cargo space. Also requires 2T free space for misc care, feeding and recreation of all basic passengers",
	}
	passengers = append(passengers, p)

	//Low Passengers
	p = model.PassengerDM{
		PassageType:  "low",
		DM:           coreDM + 1,
		Requirements: "Per passenger: 1 low berth and 10kg storage space",
	}
	passengers = append(passengers, p)

	//create a notes section to summarize next steps
	var notes = []string{"see pg 239.",
		"To the given DM, add the Effect of a (8+) Broker, Carouse or Streetwise check.",
		"Use DM -1 for each parsec beyond 1.",
		"Then roll on Passenger Traffic table using the calculated DM to determine number of available passengers at each berth level",
	}

	summary := model.PassengerTradeSummary{
		PassengerDMs:   passengers,
		PassengerNotes: notes,
	}

	log.Info().Msg("Passenger generation complete")
	return summary
}

func LoadStandardTradeFacts(ctx *util.TASContext) (*model.TradeFacts, error) {
	log := ctx.Logger()

	//determine name of trade facts file from flags
	tradeDataFilename, err := ctx.Config().Flags.GetString(TradeFileFlagName)
	if err != nil {
		tradeDataFilename = defaultTradeDataFilename
	}

	// load source data files
	log.Info().Msg("loading trade data files")

	var sourceFiles = []string{tradeDataFilename}

	fileData := util.IngestFiles("data-local/", sourceFiles)
	if !util.AllFilesReadOk(fileData) {
		log.Error().Msg("one or more files failed to load as expected")
		for _, f := range fileData {
			if !f.Ok() {
				log.Error().Err(f.Err).Str("filename", f.Name).Send()
			}
		}
		return nil, errors.New(h.UnableToContinueBecauseOfErrors)
	}
	log.Info().Msg("trade data files load complete")

	//parse trade data
	log.Info().Msg("parsing trade data files...")
	tradeFacts := &model.TradeFacts{}

	for filename, fd := range fileData {

		switch filename {

		case tradeDataFilename:
			tradeFacts, err = model.TradeFactsFromFile(fd.Data)
			if err != nil {
				return nil, err
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
		return nil, errors.New(h.UnableToContinueBecauseOfErrors)
	}

	//parse the trade data and do further, detailed validation
	hasErrors, errs := tradeFacts.Parse()
	if hasErrors { //Parse truly failed with real errors
		log.Error().Msg("parse of world UWP data failed. The following lines should point you to the problem")
		for _, e := range errs {
			log.Error().Err(e).Send()
		}
		return nil, errors.New(h.UnableToContinueBecauseOfErrors)
	} else { //Parse data did not truiely have errors but did have warnings we need to tell the user about
		log.Warn().Msg("parse of the UWP's succeeded but some assumptions were made about the data. The following lines provide more info, please check them carefully")
		for _, e := range errs {
			log.Warn().Err(e).Send()
		}
	}
	log.Info().Msg("parsing trade data files complete")
	return tradeFacts, nil
}
