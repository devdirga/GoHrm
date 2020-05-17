package services

import (
	"creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"

	tk "github.com/creativelab/toolkit"
)

type LogUserService struct {
	UserId    string
	DateMonth string
}

func (s *LogUserService) ConstructDashboardLogUser() (interface{}, error) {
	allleaves, err := s.GetAllLeaveByMonth()
	if err != nil {
		return nil, err
	}

	leaves := []models.RequestLeaveModel{}
	emergencies := []models.RequestLeaveModel{}

	for _, leave := range allleaves {
		if leave.IsEmergency == false {
			leaves = append(leaves, leave)
		} else {
			emergencies = append(emergencies, leave)
		}
	}

	remoteGroups := map[string]tk.M{}

	remotes, err := s.GetAllRemoteByMonth()
	if err != nil {
		return nil, err
	}

	for _, remote := range remotes {
		idrequest := remote.IdOp
		if _, ok := remoteGroups[idrequest]; !ok {
			remoteGroups[idrequest] = tk.M{}.Set("Remotes", []models.RemoteModel{}).Set("DateList", []string{})
		}

		remotesTemp := remoteGroups[idrequest].Get("Remotes").([]models.RemoteModel)
		remotesTemp = append(remotesTemp, remote)
		remoteGroups[idrequest].Set("Remotes", remotesTemp)

		datesTemp := remoteGroups[idrequest].Get("DateList").([]string)
		datesTemp = append(datesTemp, remote.DateLeave)
		remoteGroups[idrequest].Set("DateList", datesTemp)
	}

	return tk.M{}.Set("Leave", leaves).Set("ELeave", emergencies).Set("DataRemote", remoteGroups), nil
}

func (s *LogUserService) GetAllLeaveByMonth() ([]models.RequestLeaveModel, error) {
	match := tk.M{}.Set("$match", tk.M{}.Set("datecreateleave", tk.M{}.Set("$regex", ".*"+s.DateMonth+".*")).Set("userid", s.UserId))

	return new(repositories.LeaveMasterDboxRepo).GetByPipe([]tk.M{match})
}

func (s *LogUserService) GetAllRemoteByMonth() ([]models.RemoteModel, error) {
	match := tk.M{}.Set("$match", tk.M{}.Set("dateleave", tk.M{}.Set("$regex", ".*"+s.DateMonth+".*")).Set("userid", s.UserId))

	return new(repositories.RemoteDboxRepo).GetByPipe([]tk.M{match})
}
