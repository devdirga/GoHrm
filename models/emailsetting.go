package models

import (
	// "time"

	"github.com/creativelab/orm"
)

type EmailSettingModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            int `bson:"_id" , json:"_id" `
	SenderEmail   string
	SenderName    string
	SmtpAddress   string
	Port          int
	Username      string
	Password      string
	QcReportTo    string //recipient for QC email
}

func NewEmailSettingModel() *EmailSettingModel {
	m := new(EmailSettingModel)
	return m
}
func (e *EmailSettingModel) RecordID() interface{} {
	return e.Id
}

func (m *EmailSettingModel) TableName() string {
	return "EmailSetting"
}
