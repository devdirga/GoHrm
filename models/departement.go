package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type DepartementModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId `bson:"_id" , json:"_id" `
	Departement   string
	codename      string
}

func NewDepartementModel() *DepartementModel {
	m := new(DepartementModel)
	return m
}
func (e *DepartementModel) RecordID() interface{} {
	return e.Id
}

func (m *DepartementModel) TableName() string {
	return "DepartementList"
}
