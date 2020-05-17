package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type TmpPlaceHolderModel struct {
	orm.ModelBase         `bson:"-",json:"-"`
	Id                    bson.ObjectId `bson:"_id" json:"_id" `
	GW                    string        `bson:"gw" json:"gw" `
	ProductID             string        `bson:"productid" json:"productid" `
	ContractExpiry        time.Time     `bson:"contractexpiry" json:"contractexpiry" `
	TransactionType       string        `bson:"transactiontype" json:"transactiontype" `
	Qty                   int           `bson:"qty" json:"qty" `
	Price                 float64       `bson:"price" json:"price" `
	Otherinfo             string        `bson:"otherinfo" json:"otherinfo" `
	TransactionID         int           `bson:"transactionid" json:"transactionid" `
	AccountNumber         string        `bson:"accountnumber" json:"accountnumber" `
	TransactionDate       time.Time     `bson:"transactiondate" json:"transactiondate" `
	TransactionTime       string        `bson:"transactiontime" json:"transactiontime" `
	OrderTime             string        `bson:"ordertime" json:"ordertime" `
	Type                  string        `bson:"type" json:"type" `
	StellarorderID        string        `bson:"stellarorderid" json:"stellarorderid" `
	ExchangeorderID       string        `bson:"exchangeorderid" json:"exchangeorderid" `
	Currency              string
	TransactionDateString string `bson:"transactiondatestring" json:"transactiondatestring" `
	FileName              string
	FileType              string
	FilePath              string
	Pointvalue            float64 `bson:"pointvalue" json:"pointvalue" `
	PriceReport           float64
	PriceFractional       float64
	Fullname              string
	Clientnumber          string
	Description           string
	AccountID             string
	Unrealize             bool
	Balance               int
	DateInt               int
	Multiplier            float64
	ContractValue         float64
	Divisor               float64
	Realizedatestr        string
	ClearerId             int
	ClearerCode           string
	ClearerName           string
	ClearerClientNumber   string
	GenerateSystem        bool
	PurchaseSalesDate     int
	ContractId            int
	ContractCode          string
	BalanceClearer        int
	UnrealizeClearer      bool
	IsCutOff              bool
}

func NewTmpPlaceHolderModel() *TmpPlaceHolderModel {
	m := new(TmpPlaceHolderModel)
	m.Id = bson.NewObjectId()
	return m
}

func (e *TmpPlaceHolderModel) RecordID() interface{} {
	return e.Id
}

func (m *TmpPlaceHolderModel) TableName() string {
	return "TmpPlaceHolder"
}
