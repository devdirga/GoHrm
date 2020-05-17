package repositories

import (
	. "creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type NationalHolidayOrmRepo struct{}

func (r *NationalHolidayOrmRepo) Save(model *NationalHolidaysModel) error {
	return Ctx.Save(model)
}

func (r *NationalHolidayOrmRepo) GetByParam(param tk.M) ([]NationalHolidaysModel, error) {
	datas := []NationalHolidaysModel{}
	crs, err := Ctx.Find(NewNationalHolidaysModel(), param)
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	return datas, err
}

func (r *NationalHolidayOrmRepo) GetAll() ([]NationalHolidaysModel, error) {
	datas := []NationalHolidaysModel{}
	crs, err := Ctx.Find(NewNationalHolidaysModel(), tk.M{})
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	return datas, err
}
