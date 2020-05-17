package repositories

import (
	. "creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type OvertimeOrmRepo struct{}

func (r *OvertimeOrmRepo) Save(model *OvertimeFormModel) error {
	return Ctx.Save(model)
}

func (r *OvertimeOrmRepo) GetByParam(param tk.M) ([]OvertimeFormModel, error) {
	datas := []OvertimeFormModel{}
	crs, err := Ctx.Find(NewOvertimeFormModel(), param)
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	return datas, err
}

func (r *OvertimeOrmRepo) GetAll() ([]OvertimeFormModel, error) {
	datas := []OvertimeFormModel{}
	crs, err := Ctx.Find(NewOvertimeFormModel(), tk.M{})
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	return datas, err
}
