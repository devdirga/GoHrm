package services

import (
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
	"time"

	tk "github.com/creativelab/toolkit"
)

// LogServiceOvertime ...
type LogServiceOvertime struct {
	TypeRequest  string
	DataOvertime *OvertimeModel
}

// RequestLog ...
func (s *LogServiceOvertime) RequestLog() error {

	if s.TypeRequest == "overtime" {

		tk.Println("2. SEND LOG OVERTIME")

		p := s.DataOvertime
		tk.Println("3. SEND LOG OVERTIME")
		data := NewLogLeaveRemoteModel()
		data.TypeRequest = s.TypeRequest
		data.IdRequest = p.Id
		data.Name = p.Name
		data.Userid = p.UserId
		data.Email = p.Email
		data.Location = p.Location
		data.Project = append(data.Project, p.Project)
		data.DateFrom, _ = time.Parse("2006-01-02", p.DayList[0].Date)
		data.DateTo, _ = time.Parse("2006-01-02", p.DayList[len(p.DayList)-1].Date)
		data.DateLogCreated = time.Now()
		data.DateLogCreatedStr = data.DateLogCreated.Format("2006-01-02 15:04:05")
		list := ListLogModel{}
		list.IdRequest = data.IdRequest
		list.NameLogBy = data.Name
		list.EmailNameLogBy = data.Email
		list.RequestBy = data.Name
		list.DateLog = data.DateLogCreated
		list.DateLogStr = data.DateLogCreatedStr
		list.Description = "create a request"
		list.Status = "request"
		data.ListLog = append(data.ListLog, list)

		list = ListLogModel{}
		list.IdRequest = data.IdRequest
		list.NameLogBy = p.Name
		list.EmailNameLogBy = p.Email
		list.RequestBy = data.Name
		list.DateLog = data.DateLogCreated
		list.DateLogStr = data.DateLogCreatedStr
		list.Description = "Request sent by leader"
		list.Status = "Sent"
		data.ListLog = append(data.ListLog, list)

		repo := repositories.LogDboxRepo{}
		err := repo.SaveLog(data)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApproveDeclineLog ...
func (s *LogServiceOvertime) ApproveDeclineLog(param tk.M) error {
	if s.TypeRequest == "overtime" {
		p := s.DataOvertime

		data := NewLogLeaveRemoteModel()
		data.TypeRequest = s.TypeRequest
		data.IdRequest = p.Id
		data.Name = p.Name
		data.Userid = p.UserId
		data.Project = append(data.Project, p.Project)
		data.Email = p.Email
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.DayList[0].Date)
		data.DateTo, _ = time.Parse("2006-01-02", p.DayList[0].Date)
		data.DateLogCreated = time.Now()
		data.DateLogCreatedStr = data.DateLogCreated.Format("2006-01-02 15:04:05")
		list := ListLogModel{}
		list.IdRequest = data.IdRequest
		list.NameLogBy = param.GetString("NameLogBy")
		list.EmailNameLogBy = param.GetString("EmailNameLogBy")
		list.RequestBy = data.Name
		list.DateLog = data.DateLogCreated
		list.DateLogStr = data.DateLogCreatedStr
		list.Description = param.GetString("Desc")
		list.Status = param.GetString("Status")
		data.ListLog = append(data.ListLog, list)
		repo := repositories.LogDboxRepo{}
		err := repo.SaveLog(data)
		if err != nil {
			return err
		}
	}
	return nil
}

// CancelRequest ...
func (s *LogServiceOvertime) CancelRequest(param tk.M) error {
	if s.TypeRequest == "overtime" {
		p := s.DataOvertime
		for _, each := range p.MembersOvertime {
			data := NewLogLeaveRemoteModel()
			data.TypeRequest = s.TypeRequest
			data.IdRequest = p.Id
			data.Name = each.Name
			data.Userid = each.UserId
			data.Project = append(data.Project, p.Project)
			data.Email = each.Email
			data.Location = each.Location
			data.DateFrom, _ = time.Parse("2006-01-02", p.DayList[0].Date)
			data.DateTo, _ = time.Parse("2006-01-02", p.DayList[0].Date)
			data.DateLogCreated = time.Now()
			data.DateLogCreatedStr = data.DateLogCreated.Format("2006-01-02 15:04:05")
			list := ListLogModel{}
			list.IdRequest = data.IdRequest
			list.NameLogBy = param.GetString("NameLogBy")
			list.EmailNameLogBy = param.GetString("EmailNameLogBy")
			list.RequestBy = data.Name
			list.DateLog = data.DateLogCreated
			list.DateLogStr = data.DateLogCreatedStr
			list.Description = param.GetString("Desc")
			list.Status = param.GetString("Status")
			data.ListLog = append(data.ListLog, list)
			repo := repositories.LogDboxRepo{}
			err := repo.SaveLog(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RequestExpired ...
func (s *LogServiceOvertime) RequestExpired(param tk.M) error {
	if s.TypeRequest == "overtime" {
		p := s.DataOvertime
		data := NewLogLeaveRemoteModel()
		data.TypeRequest = s.TypeRequest
		data.IdRequest = p.Id
		data.Name = p.Name
		data.Userid = p.UserId
		data.Project = append(data.Project, p.Project)
		data.Email = p.Email
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.DayList[0].Date)
		data.DateTo, _ = time.Parse("2006-01-02", p.DayList[0].Date)
		data.DateLogCreated = time.Now()
		data.DateLogCreatedStr = data.DateLogCreated.Format("2006-01-02 15:04:05")
		list := ListLogModel{}
		list.IdRequest = data.IdRequest
		list.NameLogBy = param.GetString("NameLogBy")
		list.EmailNameLogBy = param.GetString("EmailNameLogBy")
		list.RequestBy = data.Name
		list.DateLog = data.DateLogCreated
		list.DateLogStr = data.DateLogCreatedStr
		list.Description = param.GetString("Desc")
		list.Status = param.GetString("Status")
		data.ListLog = append(data.ListLog, list)
		repo := repositories.LogDboxRepo{}
		err := repo.SaveLog(data)
		if err != nil {
			return err
		}
	}
	return nil
}
