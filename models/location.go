package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type LocationModel struct {
	orm.ModelBase  `bson:"-",json:"-"`
	Id             bson.ObjectId `bson:"_id",json:"_id"`
	Location       string        `bson:"Location",json:"Location"`
	CodeLocation   string        `bson:"CodeLocation",json:"CodeLocation"`
	TimeZone       string        `bson:"TimeZone",json:"TimeZone"`
	MemberLocation []string      `bson:"MemberLocation",json:"MemberLocation"`
	PC             []PClist      `bson:"PC",json:"PC"`
}

type PClist struct {
	IdEmp       string
	Name        string
	Location    string
	Email       string
	PhoneNumber string
	UserId      string
}

func NewLocationModel() *LocationModel {
	m := new(LocationModel)
	return m
}

func (m *LocationModel) TableName() string {
	return "LocationList"
}

func (m *LocationModel) RecordID() interface{} {
	return m.Id
}
