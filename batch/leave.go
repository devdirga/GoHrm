package batch

import (
	"creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
	"time"
)

type LeaveBatch struct{}

func (r *LeaveBatch) FixingDataLeave() error {
	repoOrm := new(repositories.LeaveOrmRepo)
	leaves, err := repoOrm.GetAll()
	if err != nil {
		return err
	}

	repoMaster := new(repositories.LeaveMasterOrmRepo)
	masterleaves, err := repoMaster.GetAll()
	if err != nil {
		return err
	}

	leaveGroup := map[string]models.RequestLeaveModel{}
	for _, mleave := range masterleaves {
		idrequest := mleave.Id
		if _, ok := leaveGroup[idrequest]; !ok {
			leaveGroup[idrequest] = mleave
		}
	}

	for _, leave := range leaves {
		idrequest := leave.IdRequest
		if leavemaster, ok := leaveGroup[idrequest]; ok {
			if len(leavemaster.StatusProjectLeader) > 0 {
				leave.StsByLeader = leavemaster.StatusProjectLeader[0].StatusRequest
			}

			leave.StsByManager = leavemaster.StatusManagerProject.StatusRequest
			leave.DayVal, leave.MonthVal, leave.YearVal, err = r.convertDateLeave(leave.DateLeave)
			if err != nil {
				return err
			}

			err := repoOrm.Save(&leave)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *LeaveBatch) convertDateLeave(dateleave string) (int, int, int, error) {
	dateLeaveParse, err := time.Parse("2006-1-2", dateleave)
	if err != nil {
		return 0, 0, 0, err
	}

	return dateLeaveParse.Day(), int(dateLeaveParse.Month()), dateLeaveParse.Year(), nil
}
