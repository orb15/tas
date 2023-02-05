package model

import (
	"strings"
	"time"
)

type SpeculativeTradeLot struct {
	LotId        int    `json:"lot-id"`
	Type         string `json:"type"`
	Example      string `json:"example"`
	TonsAvail    int    `json:"tons-avail"`
	BasePrice    int    `json:"base-price"`
	OfferPriceDM int    `json:"offer-price-dm"`
}

type SpeculativeTradeSummary struct {
	WorldName              string                `json:"world"`
	TransactionType        string                `json:"transaction-type"`
	FindSupplierOrBrokerDM int                   `json:"find-supplier-broker"`
	TradeLots              []SpeculativeTradeLot `json:"trade-lots"`
	TradeNotes             []string              `json:"notes"`
}

func (s SpeculativeTradeSummary) ToFileName() string {
	var sb strings.Builder

	now := time.Now()

	sb.WriteString("spectrade")
	sb.WriteString(us + s.TransactionType)
	sb.WriteString(us + s.WorldName)
	sb.WriteString(ds + now.Format("20060102150405"))
	sb.WriteString(".json")

	return sb.String()
}

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
	From           string                `json:"from-world"`
	To             string                `json:"to-world"`
	PassengerTrade PassengerTradeSummary `json:"passenger-trade"`
	FreightTrade   FreightTradeSummary   `json:"freight-trade"`
	MailTrade      MailTradeSummary      `json:"mail-trade"`
}

func (s StandardTradeModifiers) ToFileName() string {
	var sb strings.Builder

	now := time.Now()

	sb.WriteString("stdtrade")
	sb.WriteString(us + s.From)
	sb.WriteString(us + s.To)
	sb.WriteString(ds + now.Format("20060102150405"))
	sb.WriteString(".json")

	return sb.String()
}
