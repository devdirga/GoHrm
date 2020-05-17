package models

import (
	"github.com/creativelab/orm"
	// "gopkg.in/mgo.v2/bson"
	// "time"
)

type TopMenuModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string `bson:"_id" , json:"_id" `
	PageId        string `bson:"PageId", json:"PageId"`
	Parent        string `bson:"Parent", json:"Parent"`
	Title         string `bson:"Title", json:"Title"`
	Url           string `bson:"Url", json:"Url"`
	IndexMenu     int    `bson:"IndexMenu", json:"IndexMenu"`
	Icon          string
	Enable        bool `bson:"Enable", json:"Enable"`
	Haschild      bool
}

func NewTopMenuModel() *TopMenuModel {
	m := new(TopMenuModel)
	return m
}

func (e *TopMenuModel) RecordID() interface{} {
	return e.Id
}

func (m *TopMenuModel) TableName() string {
	return "TopMenu"
}
