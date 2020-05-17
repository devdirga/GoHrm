package repositories

import (
	. "creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type RemoteOrmRepo struct{}

func (r *RemoteOrmRepo) Save(model *RemoteModel) error {
	return Ctx.Save(model)
}

func (r *RemoteOrmRepo) GetByParam(param tk.M) ([]RemoteModel, error) {
	datas := []RemoteModel{}
	crs, err := Ctx.Find(NewRemoteModel(), param)
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	return datas, err
}

func (r *RemoteOrmRepo) GetAll() ([]RemoteModel, error) {
	datas := []RemoteModel{}
	crs, err := Ctx.Find(NewRemoteModel(), tk.M{})
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	return datas, err
}

func (r *RemoteOrmRepo) ChekDatelist(param []tk.M) ([]RemoteModel, error) {
	datas := []RemoteModel{}
	csr, err := Ctx.Connection.
		NewQuery().
		Command("pipe", param).
		From("remote").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err := csr.Fetch(&datas, 0, false); err != nil {
		return nil, nil
	}
	return datas, err
}
