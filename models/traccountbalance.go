package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type TrAccountBalanceModel struct {
	orm.ModelBase      `bson:"-",json:"-"`
	Id                 bson.ObjectId `bson:"_id" , json:"_id" `
	ClientNumber       string
	AccountId          string
	BaseCurrencyId     int
	BaseCurrencyCode   string
	TradeDateInt       int
	TradeDateStr       string
	TradeDate          time.Time
	CurrencyId         int
	CurrencyCode       string
	SpotRate           float64
	AccountNumber      string
	AccountCashBalance float64
	PaymentReceipt     float64
	RealizeProfitLoss  float64
	MarketFee          float64
	ClrCommission      float64
	NfaFee             float64
	MiscFee            float64
	TotalFee           float64
	NewCashBalance     float64
	OpenTradeEquity    float64
	TotalEquity        float64
	NewLiquidValue     float64
	MtdPaymentReceipt  float64
	MtdRealizedPl      float64
	MtdMarketFee       float64
	MtdClrCommission   float64
	MtdNfaFee          float64
	MtdMiscFee         float64
	MtdTotalFee        float64
	NlvTd              float64
	NlvTd1             float64
	NlvTd2             float64
	NlvTd3             float64
	NlvTd4             float64
	NlvDiff1           float64
	NlvDiff2           float64
	NlvDiff3           float64
	NlvDiff4           float64
	CurOrder           int
}

func NewTrAccountBalanceModel() *TrAccountBalanceModel {
	m := new(TrAccountBalanceModel)
	m.Id = bson.NewObjectId()
	return m
}

func (e *TrAccountBalanceModel) RecordID() interface{} {
	return e.Id
}

func (m *TrAccountBalanceModel) TableName() string {
	return "TrAccountBalance"
}
