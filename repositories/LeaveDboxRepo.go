package repositories

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"

	tk "github.com/creativelab/toolkit"
)

type LeaveDboxRepo struct{}

func (r *LeaveDboxRepo) GetLastLeaveDate() AprovalRequestLeaveModel {
	data := AprovalRequestLeaveModel{}

	crs, err := Ctx.Connection.NewQuery().From(NewAprovalRequestLeaveModel().TableName()).Order("-dateleave").Take(1).Cursor(nil)
	if crs != nil {
		defer crs.Close()
	}
	if err != nil {
		return data
	}

	err = crs.Fetch(&data, 1, false)
	if err != nil {
		return data
	}

	return data
}

func (r *LeaveDboxRepo) GetByPipe(pipe []tk.M) ([]AprovalRequestLeaveModel, error) {
	datas := []AprovalRequestLeaveModel{}
	crs, err := Ctx.Connection.NewQuery().From(NewAprovalRequestLeaveModel().TableName()).Command("pipe", pipe).Cursor(nil)
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
func (r *LeaveDboxRepo) GetByPipeSpecial(pipe []tk.M, special bool) ([]AprovalRequestLeaveModel, error) {
	datas := []AprovalRequestLeaveModel{}
	crs, err := Ctx.Connection.NewQuery().From(NewAprovalRequestLeaveModel().TableName()).Command("pipe", pipe).Cursor(nil)
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
	IdsLeave := []string{}
	for _, each := range datas {
		IdsLeave = append(IdsLeave, each.IdRequest)
	}
	// ===
	newdatas := []AprovalRequestLeaveModel{}
	reqdatas := []RequestLeaveModel{}
	fixIdsLeave := helper.DistincValue(IdsLeave)
	tk.Println(fixIdsLeave)
	pipe2 := []tk.M{}
	// tk.Println("id ----------- ", fixIdsLeave)
	match := tk.M{}.Set("_id", tk.M{}.Set("$in", fixIdsLeave))
	match = tk.M{}.Set("isspecials", special)
	pipe2 = append(pipe2, tk.M{}.Set("$match", match))
	crs, err = Ctx.Connection.NewQuery().From(NewRequestLeave().TableName()).Command("pipe", pipe2).Cursor(nil)
	defer crs.Close()
	if err != nil {
		return newdatas, err
	}
	err = crs.Fetch(&reqdatas, 0, false)
	if err != nil {
		return newdatas, err
	}
	if special == true {
		tk.Println("sepsial ------- ", special)
		//tk.Println("sepsial -------1  ", tk.JsonString(reqdatas))
	}
	// tk.Println("sepsial -------1  ", tk.JsonString(reqdatas))
	if len(reqdatas) > 0 {
		for _, data := range datas {
			for _, each := range reqdatas {
				if each.Id == data.IdRequest {
					if each.IsSpecials == special {
						if special == true {
							tk.Println("sepsial ------- ", special)
							newdatas = append(newdatas, data)
						} else {
							newdatas = append(newdatas, data)
						}

					}

				}
			}
		}
	}
	// tk.Println("sepsial ------- ", tk.JsonString(newdatas))
	return newdatas, err
}

func (r *LeaveDboxRepo) GetBySL(pipe []tk.M) ([]tk.M, error) {
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
