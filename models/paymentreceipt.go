package models

import (
	"github.com/creativelab/orm"
)

type PaymentReceiptTypeModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            int `bson:"_id" , json:"_id" `
	Description   string
	InternalDesc  string
	DateCreated   string
	DateUpdated   string
	UpdateUser    string
}

func NewPaymentReceiptTypeModel() *PaymentReceiptTypeModel {
	m := new(PaymentReceiptTypeModel)
	return m
}
func (e *PaymentReceiptTypeModel) RecordID() interface{} {
	return e.Id
}

func (m *PaymentReceiptTypeModel) TableName() string {
	return "PaymentReceiptTypes"
}
