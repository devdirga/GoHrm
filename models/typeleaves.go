package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type TypeLeavesModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId `bson:"_id",json:"_id"`
	Description   string
	Days          int
	Code          string
}

func NewTypeLeavesModel() *TypeLeavesModel {
	m := new(TypeLeavesModel)
	return m
}

func (m *TypeLeavesModel) TableName() string {
	return "TypeLeave"
}
