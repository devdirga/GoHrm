package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type MailSenderLogModel struct {
	orm.ModelBase   `bson:"-",json:"-"`
	Id              bson.ObjectId ` bson:"_id" , json:"_id" `
	TradeDate       string
	ClientId        string
	ClientName      string
	Destination     []MailSenderDetailModel
	AttachmentCount int
	Checked         int
	Status          string
	LastSent        time.Time
	LastError       string
}

type MailSenderDetailModel struct {
	SentType  string
	EmailId   int
	Email     string
	EmailName string
}

func NewMailSenderLogModel() *MailSenderLogModel {
	m := new(MailSenderLogModel)
	m.Id = bson.NewObjectId()
	return m
}

func (e *MailSenderLogModel) RecordID() interface{} {
	return e.Id
}

func (m *MailSenderLogModel) TableName() string {
	return "MailSenderLog"
}
