package models

import (
	"github.com/creativelab/orm"
	"gopkg.in/mgo.v2/bson"
)

type ProjectRuleModel struct {
	orm.ModelBase `bson:"-",json:"-"`
	Id            bson.ObjectId `bson:"_id" , json:"_id"`
	Name          string        `bson:"Name" , json:"Name"`
	AliasName     string        `bson:"AliasName" , json:"AliasName"`
	Level         int           `bson:"Level" , json:"Level"`
}

func NewProjectRuleModel() *ProjectRuleModel {
	m := new(ProjectRuleModel)

	return m
}
func (e *ProjectRuleModel) RecordID() interface{} {
	return e.Id
}
func (m *ProjectRuleModel) TableName() string {
	return "ProjectRule"
}
