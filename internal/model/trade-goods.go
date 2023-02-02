package model

import (
	"encoding/json"
)

type TradeGoodsMap map[int]*TradeGood

type TradeGoodJSON struct {
	CommonGoods   []TradeGood `json:"common-goods"`
	AdvancedGoods []TradeGood `json:"advanced-goods"`
	IllegalGoods  []TradeGood `json:"illegal-goods"`
}

type TradeDM struct {
	Code string `json:"code"`
	Mod  int    `json:"mod"`
}

type TradeGood struct {
	Value          int       `json:"value"`
	Type           string    `json:"type"`
	TonsDice       int       `json:"tons-dice"`
	TonsMultiplier int       `json:"tons-multi"`
	BasePrice      int       `json:"base-price"`
	Examples       string    `json:"examples"`
	Availability   []string  `json:"availability"`
	PurchaseDMs    []TradeDM `json:"purchase-dms"`
	SaleDMs        []TradeDM `json:"sale-dms"`
}

func TradeGoodsFromFile(b []byte) (TradeGoodsMap, error) {

	var data TradeGoodJSON
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	dataMap := make(TradeGoodsMap)
	for _, d := range data.CommonGoods {
		tg := d
		dataMap[d.Value] = &tg
	}
	for _, d := range data.AdvancedGoods {
		tg := d
		dataMap[d.Value] = &tg
	}
	for _, d := range data.IllegalGoods {
		tg := d
		dataMap[d.Value] = &tg
	}
	return dataMap, nil
}
