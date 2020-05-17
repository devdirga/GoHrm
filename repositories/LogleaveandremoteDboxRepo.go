package repositories

import (
	. "creativelab/ecleave-dev/models"

	db "github.com/creativelab/dbox"
)

type LogDboxRepo struct{}

func (c *LogDboxRepo) SaveLog(data *LogleaveRemoteModel) error {
	csr, err := Ctx.Connection.NewQuery().Select().From(NewLogLeaveRemoteModel().TableName()).Where(db.Eq("idrequest", data.IdRequest)).Cursor(nil)
	if err != nil {
		return err
	}
	defer csr.Close()
	result := []LogleaveRemoteModel{}
	err = csr.Fetch(&result, 0, false)
	if err != nil {
		return err
	}
	if csr.Count() > 0 {
		existData := result[0]
		list := data.ListLog
		existData.ListLog = append(existData.ListLog, list...)
		err = Ctx.Save(&existData)
		if err != nil {
			return err
		}
	} else {
		err = Ctx.Save(data)
		if err != nil {
			return err
		}
	}
	return nil
}
