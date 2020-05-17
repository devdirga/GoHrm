package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type UserProfileModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string ` bson:"_id" , json:"_id" `
	FirstName     string
	LastName      string
	Email         string
	PhoneNo       string
	Age           string
	Gender        string
	Designation   string
	Location      string
	Departement   string
	Password      string
	EmployeeID    string
	Address       string
}

func NewUserProfileModel() *MailSenderLogModel {
	m := new(MailSenderLogModel)
	m.Id = bson.NewObjectId()
	return m
}

func (e *UserProfileModel) RecordID() interface{} {
	return e.Id
}

func (m *UserProfileModel) TableName() string {
	return "UserProfile"
}
