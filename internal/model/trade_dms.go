package model

type PassengerDM struct {
	PassageType  string `json:"type"`
	DM           int    `json:"dm"`
	Requirements string `json:"requirements"`
}

type PassengerTradeSummary struct {
	PassengerDMs   []PassengerDM `json:"dms"`
	PassengerNotes []string      `json:"notes"`
}

type FreightDM struct {
	LotType string `json:"lot-type"`
	DM      int    `json:"dm"`
}

type FreightTradeSummary struct {
	FreightDMs   []FreightDM `json:"dms"`
	FreightNotes []string    `json:"notes"`
}

type MailTradeSummary struct {
	MailDM    int      `json:"dm"`
	LotsAvail int      `json:"lots-avail"`
	MailNotes []string `json:"notes"`
}

type StandardTradeModifiers struct {
	PassengerTrade PassengerTradeSummary `json:"passenger-trade"`
	FreightTrade   FreightTradeSummary   `json:"freight-trade"`
	MailTrade      MailTradeSummary      `json:"mail-trade"`
}

type SpeculativeTradeLot struct {
	LotId        int    `json:"lot-id"`
	Type         string `json:"type"`
	Example      string `json:"example"`
	TonsAvail    int    `json:"tons-avail"`
	BasePrice    int    `json:"base-price"`
	OfferPriceDM int    `json:"offer-price-dm"`
}

type SpeculativeTradeSummary struct {
	FindSupplierOrBrokerDM int                   `json:"find-supplier-broker"`
	TradeLots              []SpeculativeTradeLot `json:"trade-lots"`
	TradeNotes             []string              `json:"notes"`
}
