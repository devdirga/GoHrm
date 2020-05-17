package models

import (
	"time"

	"github.com/creativelab/orm"
)

type EmailModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            int `bson:"_id" , json:"_id" `
	Email         string
	FirstName     string
	LastName      string
	DateCreated   time.Time
	DateUpdated   time.Time
	UpdateUser    string
}

func NewEmailModel() *EmailModel {
	m := new(EmailModel)
	return m
}
func (e *EmailModel) RecordID() interface{} {
	return e.Id
}

func (m *EmailModel) TableName() string {
	return "Emails"
}
