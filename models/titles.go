package models

import (
	"github.com/creativelab/orm"
)

type TitlesModel struct {
	orm.ModelBase       `bson:"-",json:"-"`
	Id             int `bson:"_id" , json:"_id" `
	Title        	string
	Datecreated     string
}

func NewTitlesModel() *TitlesModel {
	m := new(TitlesModel)
	return m
}
func (e *TitlesModel) RecordID() interface{} {
	return e.Id
}

func (m *TitlesModel) TableName() string {
	return "Titles"
}
