package repositories

import (
	. "creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type LeaveMasterOrmRepo struct{}

func (r *LeaveMasterOrmRepo) GetByParam(filter tk.M) ([]RequestLeaveModel, error) {
	rows := []RequestLeaveModel{}

	crs, err := Ctx.Find(NewRequestLeave(), filter)
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return rows, err
	}

	err = crs.Fetch(&rows, 0, false)
	if err != nil {
		return rows, err
	}

	return rows, nil
}

func (r *LeaveMasterOrmRepo) GetAll() ([]RequestLeaveModel, error) {
	datas := []RequestLeaveModel{}

	csr, err := Ctx.Find(NewRequestLeave(), nil)
	if csr != nil {
		defer csr.Close()
	}

	if err != nil {
		return datas, err
	}

	err = csr.Fetch(&datas, 0, false)
	if err != nil {
		return datas, err
	}

	return datas, nil
}
