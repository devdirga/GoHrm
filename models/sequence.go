package models

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/creativelab/orm"
)

type SequenceModel struct {
	orm.ModelBase       `bson:"-",json:"-"`
	Id                  bson.ObjectId `bson:"_id" , json:"_id" `
	Collname        string
	Lastnumber 		int
	Locked         	bool
}

func NewSequenceModel() *SequenceModel {
	m := new(SequenceModel)
	m.Id = bson.NewObjectId()
	return m
}
func (e *SequenceModel) RecordID() interface{} {
	return e.Id
}

func (m *SequenceModel) TableName() string {
	return "Sequence"
}
