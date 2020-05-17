package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type ChangeOptionModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            string `bson:"_id" , json:"_id"`
	UserId        string
	Name          string
	Email         string
	Remote        RemoteOption
}
type RemoteOption struct {
	RemoteActive      bool
	ConditionalRemote int
	FullMonth         bool
	Monthly           bool
}

func NewChangeOptionModel() *ChangeOptionModel {
	m := new(ChangeOptionModel)
	m.Id = bson.NewObjectId().Hex()
	return m
}

func (e *ChangeOptionModel) RecordID() interface{} {
	return e.Id
}

func (m *ChangeOptionModel) TableName() string {
	return "ChangeOptionUser"
}
