package repositories

import (
	. "creativelab/ecleave-dev/models"

	"github.com/creativelab/dbox"

	tk "github.com/creativelab/toolkit"
)

type UserOrmRepo struct{}

func (r *UserOrmRepo) GetByID(id string) (SysUserModel, error) {
	row := SysUserModel{}
	filter := tk.M{}.Set("where", dbox.And(dbox.Eq("_id", id)))

	csr, err := Ctx.Find(NewSysUserModel(), filter)
	if csr != nil {
		defer csr.Close()
	}

	if err != nil {
		return row, err
	}

	err = csr.Fetch(&row, 1, false)
	if err != nil {
		return row, err
	}

	return row, nil
}
func (r *UserOrmRepo) GetByParam(filter tk.M) ([]SysUserModel, error) {
	row := []SysUserModel{}
	crs, err := Ctx.Find(NewSysUserModel(), filter)
	if crs != nil {
		defer crs.Close()
	}
	if err != nil {
		return row, err
	}

	err = crs.Fetch(&row, 0, false)
	if err != nil {
		return row, err
	}

	return row, err
}

func (r *UserOrmRepo) GetByPipe(pipe []tk.M) ([]SysUserModel, error) {
	datas := []SysUserModel{}
	crs, err := Ctx.Connection.NewQuery().From(NewSysUserModel().TableName()).Command("pipe", pipe).Cursor(nil)
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

func (r *UserOrmRepo) Save(user SysUserModel) (SysUserModel, error) {
	err := Ctx.Save(&user)

	return user, err
}
