package repositories

import (
	"creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type LeaveMasterDboxRepo struct{}

func (r *LeaveMasterDboxRepo) GetByPipe(pipe []tk.M) ([]models.RequestLeaveModel, error) {
	datas := []models.RequestLeaveModel{}
	crs, err := Ctx.Connection.NewQuery().From(models.NewRequestLeave().TableName()).Command("pipe", pipe).Cursor(nil)
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
