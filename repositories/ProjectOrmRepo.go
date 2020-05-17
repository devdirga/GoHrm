package repositories

import (
	. "creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type ProjectOrmRepo struct{}

func (r *ProjectOrmRepo) GetByParam(param tk.M) ([]ProjectModel, error) {
	rows := []ProjectModel{}

	crs, err := Ctx.Find(NewListProject(), param)
	if crs != nil {
		defer crs.Close()
	}
	if err != nil {
		return rows, err
	}

	err = crs.Fetch(&rows, 0, false)
	return rows, err
}
func (r *ProjectOrmRepo) GetByPipe(pipe []tk.M) ([]ProjectModel, error) {
	datas := []ProjectModel{}
	crs, err := Ctx.Connection.NewQuery().From(NewListProject().TableName()).Command("pipe", pipe).Cursor(nil)
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
func (r *ProjectOrmRepo) GetByPipeProjection(pipe []tk.M) ([]tk.M, error) {
	datas := []tk.M{}
	crs, err := Ctx.Connection.NewQuery().From(NewListProject().TableName()).Command("pipe", pipe).Cursor(nil)
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
func (r *ProjectOrmRepo) Delete(project *ProjectModel) error {
	return Ctx.Delete(project)
}

func (r *ProjectOrmRepo) GetByOvertime(pipe []tk.M) ([]tk.M, error) {
	data := []tk.M{}
	csr, err := Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("EmployeeOvertime").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil, nil
	}

	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil, nil
	}
	return data, err
}

func (r *ProjectOrmRepo) GetByDateleave(pipe []tk.M) ([]tk.M, error) {
	data := []tk.M{}
	csr, err := Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("requestLeaveByDate").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil, nil
	}

	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil, nil
	}
	return data, err
}

func (r *ProjectOrmRepo) GetByDateRemote(pipe []tk.M) ([]tk.M, error) {
	data := []tk.M{}
	csr, err := Ctx.Connection.
		NewQuery().
		Command("pipe", pipe).
		From("remote").
		Cursor(nil)

	if csr != nil {
		defer csr.Close()
	}
	if err != nil {
		return nil, nil
	}

	if err := csr.Fetch(&data, 0, false); err != nil {
		return nil, nil
	}
	tk.Println("---------- prj ", data)
	return data, err
}
