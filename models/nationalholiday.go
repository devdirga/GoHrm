package models

import (
	"time"

	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type NationalHolidaysModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId `bson:"_id" , json:"_id" `
	Location      string
	Date          time.Time
	ListDate      []time.Time
	Year          int
	Month         int
	Description   string
}

func NewNationalHolidaysModel() *NationalHolidaysModel {
	m := new(NationalHolidaysModel)
	return m
}
func (e *NationalHolidaysModel) RecordID() interface{} {
	return e.Id
}

func (m *NationalHolidaysModel) TableName() string {
	return "NationalHolidays"
}
