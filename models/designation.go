package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type DesignationModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId `bson:"_id" , json:"_id" `
	Designation   string
	Codename      string `bson:"codename" , json:"codename" `
	Code          int    `bson:"code" , json:"code" `
}

func NewDesignationModel() *DesignationModel {
	m := new(DesignationModel)
	return m
}
func (e *DesignationModel) RecordID() interface{} {
	return e.Id
}

func (m *DesignationModel) TableName() string {
	return "DesignationList"
}

type DesignationList struct {
	Code        string
	Designation string
}
