package orm

import (
	//"fmt"
	//_ "github.com/creativelab/dbox"
	tk "github.com/creativelab/toolkit"
	//err "github.com/creativelab/errorlib"
)

type IModel interface {
	//Find(map[string]interface{}) base.ICursor
	RecordID() interface{}
	PreSave() error
	PostSave() error
	SetM(IModel) IModel
	TableName() string
	PrepareID() interface{}
}

type ModelBase struct {
	model IModel
}

func NewModel(m IModel) IModel {
	m.SetM(m)
	return m
}

func (m *ModelBase) SetM(model IModel) IModel {
	m.model = model
	return model
}

func (m *ModelBase) RecordID() interface{} {
	if m.model == nil {
		return nil
	}
	return tk.Id(m.model)
}

func (m *ModelBase) PrepareID() interface{} {
	return nil
}

func (m *ModelBase) PreSave() error {
	return nil
}

func (m *ModelBase) PostSave() error {
	return nil
}
