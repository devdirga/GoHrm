package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type HRDAdminModel struct {
	orm.ModelBase    `bson:"-",json:"-"`
	Id               string `bson:"_id",json:"_id"`
	ManagingDirector DetailsOcupation
	AccountManager   DetailsOcupation
	Staf             []DetailsOcupation
}

type DetailsOcupation struct {
	IdEmp    string
	Name     string
	Location string
	Email    string
	Contact  string
}

func NewHRDAdminModel() *HRDAdminModel {
	m := new(HRDAdminModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *HRDAdminModel) RecordID() interface{} {
	return e.Id
}

func (e *HRDAdminModel) TableName() string {
	return "HR"
}
