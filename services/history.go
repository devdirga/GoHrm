package services

import (
	"creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type HistoryService struct {
	IDRequest      string
	Name           string
	UserId         string
	RequestType    string
	Desc           string
	StatusApproval string
	Status         string
	Reason         string
	ManagerApprove string
	DateTo         string
	DateFrom       string
}

func (s *HistoryService) Push(IsEmergency bool) error {
	history, _ := s.GetHistoryByUserId(s.UserId)

	if history.Id == "" {
		history.Id = bson.NewObjectId().Hex()
	}
	history.UserId = s.UserId
	historyDetail := models.HistoryDetails{}
	historyDetail.Name = s.Name
	historyDetail.UserId = s.UserId
	historyDetail.IdRequest = s.IDRequest
	historyDetail.Description = s.Desc
	historyDetail.RequestType = s.RequestType
	historyDetail.StatusApproval = s.StatusApproval
	historyDetail.Status = s.Status
	historyDetail.Reason = s.Reason
	historyDetail.ManagerApprove = s.ManagerApprove
	historyDetail.IsEmergency = IsEmergency
	historyDetail.DateFrom = s.DateFrom
	historyDetail.DateTo = s.DateTo

	found := false
	index := -1
	for k, hisDetail := range history.Leavehistory {
		if hisDetail.IdRequest == s.IDRequest {
			found = true
			index = k
			break
		}
	}

	now := time.Now().UTC()
	if found {
		historyDetail.UpdatedAt = now
		historyDetail.CreatedAt = history.Leavehistory[index].CreatedAt
		history.Leavehistory[index] = historyDetail
	} else {
		historyDetail.CreatedAt = now
		historyDetail.UpdatedAt = now
		history.Leavehistory = append(history.Leavehistory, historyDetail)
	}

	return new(repositories.HistoryRequestOrmRepo).Save(&history)
}

func (s *HistoryService) GetHistoryByUserId(userid string) (models.HistoryLeaveModel, error) {
	return new(repositories.HistoryRequestOrmRepo).GetHistoryByUserId(userid)
}
func (s *HistoryService) GetHistoryForAdmin(userid string, jobrolelevel int) ([]models.HistoryLeaveModel, error) {
	return new(repositories.HistoryRequestOrmRepo).GetHistoryForAdmin(userid, jobrolelevel)
}
