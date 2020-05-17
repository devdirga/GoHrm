package repositories

import (
	. "creativelab/ecleave-dev/models"

	"github.com/creativelab/dbox"
	tk "github.com/creativelab/toolkit"
)

type OvertimeDboxRepo struct{}

func (r *OvertimeDboxRepo) GetByParam(filter []*dbox.Filter) ([]OvertimeFormModel, error) {
	datas := []OvertimeFormModel{}
	query := Ctx.Connection.NewQuery().From(NewOvertimeFormModel().TableName())
	if len(filter) > 0 {
		query.Where(filter...)
	}

	crs, err := query.Cursor(nil)
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	if err != nil {
		return datas, err
	}

	return datas, nil
}

func (r *OvertimeDboxRepo) GetByPipe(pipe []tk.M) ([]OvertimeFormModel, error) {
	datas := []OvertimeFormModel{}
	crs, err := Ctx.Connection.NewQuery().From(NewOvertimeFormModel().TableName()).Command("pipe", pipe).Cursor(nil)
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	if err != nil {
		return datas, err
	}

	return datas, err
}

func (r *OvertimeDboxRepo) GetNewOvertimeByPipe(pipe []tk.M) ([]tk.M, error) {
	datas := []tk.M{}
	crs, err := Ctx.Connection.NewQuery().From(NewOvertimeModel().TableName()).Command("pipe", pipe).Cursor(nil)
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	if err != nil {
		return datas, err
	}

	return datas, err
}
