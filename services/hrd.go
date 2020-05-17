package services

import (
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"

	"github.com/creativelab/dbox"
	tk "github.com/creativelab/toolkit"
)

type HrdServices struct{}

func (s *HrdServices) DeleteByID(id string) error {
	hrdR := repositories.HRDOrmRepo{}

	hrd, err := s.GetByID(id)
	if err != nil {
		return err
	}

	return hrdR.Delete(&hrd)
}

func (s *HrdServices) GetByID(id string) (HRDAdminModel, error) {
	dboxFilter := []*dbox.Filter{}
	dboxFilter = append(dboxFilter, dbox.Eq("_id", id))
	filter := tk.M{}.Set("where", dbox.And(dboxFilter...))

	row := HRDAdminModel{}
	hrds, err := new(repositories.HRDOrmRepo).GetByParam(filter)
	if err != nil {
		return row, err
	}

	if len(hrds) > 0 {
		row = hrds[0]
	}

	return row, err
}
