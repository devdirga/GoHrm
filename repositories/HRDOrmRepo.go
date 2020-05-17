package repositories

import (
	. "creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type HRDOrmRepo struct{}

func (r *HRDOrmRepo) GetByParam(param tk.M) ([]HRDAdminModel, error) {
	rows := []HRDAdminModel{}
	crs, err := Ctx.Find(NewHRDAdminModel(), param)
	if crs != nil {
		defer crs.Close()
	}
	if err != nil {
		return rows, err
	}

	err = crs.Fetch(&rows, 0, false)

	return rows, err
}

func (r *HRDOrmRepo) Delete(hrd *HRDAdminModel) error {
	return Ctx.Delete(hrd)
}
