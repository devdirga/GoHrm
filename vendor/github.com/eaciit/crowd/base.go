package crowd

import (
	"errors"

	"github.com/creativelab/toolkit"
)

type ICrowd interface {
}

type SliceBase struct {
	data interface{}
}

func (s *SliceBase) SetData(data interface{}) error {
	if !toolkit.IsPointer(data) || !toolkit.IsSlice(data) {
		return errors.New("crowd.SliceBase.SetData: data is not pointer of slice")
	}

	s.data = data
	return nil
}

func (s *SliceBase) GetData() interface{} {
	if s.data != nil {
		return s.data
	} else {
		return nil
	}
}
func (s *SliceBase) Item(i int) interface{} {
	return toolkit.SliceItem(s.data, i)
}

func (s *SliceBase) Len() int {
	return toolkit.SliceLen(s.data)
}

func (s *SliceBase) Set(i int, d interface{}) error {
	e := toolkit.SliceSetItem(s.data, i, d)
	if e != nil {
		return errors.New(toolkit.Sprintf("SliceBase.Set: [%d] %s", i, e.Error()))
	}
	return nil
}
