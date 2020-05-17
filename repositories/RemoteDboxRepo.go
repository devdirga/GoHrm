package repositories

import (
	. "creativelab/ecleave-dev/models"

	"github.com/creativelab/dbox"
	tk "github.com/creativelab/toolkit"
)

type RemoteDboxRepo struct{}

func (r *RemoteDboxRepo) GetByParam(filter []*dbox.Filter) ([]RemoteModel, error) {
	datas := []RemoteModel{}
	query := Ctx.Connection.NewQuery().From(NewRemoteModel().TableName())
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

func (r *RemoteDboxRepo) GetByPipe(pipe []tk.M) ([]RemoteModel, error) {
	datas := []RemoteModel{}
	crs, err := Ctx.Connection.NewQuery().From(NewRemoteModel().TableName()).Command("pipe", pipe).Cursor(nil)
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
