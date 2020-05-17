package repositories

import (
	"creativelab/ecleave-dev/models"

	"github.com/creativelab/dbox"

	tk "github.com/creativelab/toolkit"
)

type HistoryRequestOrmRepo struct{}

func (r *HistoryRequestOrmRepo) GetHistoryByUserId(userid string) (models.HistoryLeaveModel, error) {
	row := models.HistoryLeaveModel{}
	filter := tk.M{}.Set("where", dbox.And(dbox.Eq("userid", userid))).Set("order", []string{"-leavehistory.createdat"})

	csr, err := Ctx.Find(models.NewHistoryLeaveModel(), filter)
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
func (r *HistoryRequestOrmRepo) GetHistoryForAdmin(userid string, jobrolelevel int) ([]models.HistoryLeaveModel, error) {
	row := []models.HistoryLeaveModel{}
	filter := tk.M{}.Set("order", []string{"-leavehistory.createdat"})

	csr, err := Ctx.Find(models.NewHistoryLeaveModel(), filter)
	if csr != nil {
		defer csr.Close()
	}
	tk.Println(filter)
	if err != nil {
		return row, err
	}

	err = csr.Fetch(&row, 0, false)
	if err != nil {
		return row, err
	}

	return row, nil
}
func (r *HistoryRequestOrmRepo) GetByParam(filter tk.M) ([]models.HistoryLeaveModel, error) {
	rows := []models.HistoryLeaveModel{}

	csr, err := Ctx.Find(models.NewHistoryLeaveModel(), filter)
	if err != nil {
		return rows, err
	}

	err = csr.Fetch(&rows, 0, false)
	if err != nil {
		return rows, err
	}

	return rows, nil
}

func (r *HistoryRequestOrmRepo) Save(history *models.HistoryLeaveModel) error {
	return Ctx.Save(history)
}
