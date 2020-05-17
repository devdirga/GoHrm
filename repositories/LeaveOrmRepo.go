package repositories

import (
	. "creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type LeaveOrmRepo struct{}

func (r *LeaveOrmRepo) GetByParamApprovalLeave(filter tk.M) ([]AprovalRequestLeaveModel, error) {
	rows := []AprovalRequestLeaveModel{}

	crs, err := Ctx.Find(NewAprovalRequestLeaveModel(), filter)
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

func (r *LeaveOrmRepo) GetByParamMasterLeave(filter tk.M) ([]RequestLeaveModel, error) {
	rows := []RequestLeaveModel{}

	crs, err := Ctx.Find(NewRequestLeave(), filter)
	if err != nil {
		return rows, err
	}

	err = crs.Fetch(&rows, 0, false)
	if err != nil {
		return rows, err
	}

	return rows, err
}

func (r *LeaveOrmRepo) Save(model *AprovalRequestLeaveModel) error {
	return Ctx.Save(model)
}

func (r *LeaveOrmRepo) GetAll() ([]AprovalRequestLeaveModel, error) {
	datas := []AprovalRequestLeaveModel{}

	csr, err := Ctx.Find(NewAprovalRequestLeaveModel(), nil)
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
