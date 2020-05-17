package services

import (
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
	"time"

	tk "github.com/creativelab/toolkit"
)

type LogService struct {
	DataLeave   *RequestLeaveModel
	DataRemote  *RemoteModel
	TypeRequest string
}

func (s *LogService) RequestLog() error {
	data := NewLogLeaveRemoteModel()
	userrepo := repositories.UserOrmRepo{}
	if s.TypeRequest == "leave" {
		p := s.DataLeave
		data.TypeRequest = "leave"
		if p.IsEmergency {
			data.TypeRequest = "emergency leave"
		}
		data.IdRequest = p.Id
		data.Name = p.Name
		data.Userid = p.UserId
		data.Email = p.Email
		data.Project = p.Project
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("02-01-2006", p.LeaveFrom)
		data.DateTo, _ = time.Parse("02-01-2006", p.LeaveTo)
	} else {
		p := s.DataRemote
		data.TypeRequest = "remote"
		data.IdRequest = p.IdOp
		data.Name = p.Name
		data.Userid = p.UserId
		for _, pro := range p.Projects {
			data.Project = append(data.Project, pro.ProjectName)
		}
		user, err := userrepo.GetByID(p.UserId)
		if err != nil {
			return err
		}
		data.Email = user.Email
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.From)
		data.DateTo, _ = time.Parse("2006-01-02", p.To)
	}
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
	repo := repositories.LogDboxRepo{}
	err := repo.SaveLog(data)
	if err != nil {
		return err
	}
	return nil
}
func (s *LogService) ApproveDeclineLog(param tk.M) error {
	data := NewLogLeaveRemoteModel()
	userrepo := repositories.UserOrmRepo{}
	if s.TypeRequest == "leave" {
		p := s.DataLeave
		data.TypeRequest = "leave"
		if p.IsEmergency {
			data.TypeRequest = "emergency leave"
		}
		data.IdRequest = p.Id
		data.Name = p.Name
		data.Userid = p.UserId
		data.Email = p.Email
		data.Project = p.Project
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.LeaveFrom)
		data.DateTo, _ = time.Parse("2006-01-02", p.LeaveTo)
	} else {
		p := s.DataRemote
		data.TypeRequest = "remote"
		data.IdRequest = p.IdOp
		data.Name = p.Name
		data.Userid = p.UserId
		for _, pro := range p.Projects {
			data.Project = append(data.Project, pro.ProjectName)
		}
		user, err := userrepo.GetByID(p.UserId)
		if err != nil {
			return err
		}
		data.Email = user.Email
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.From)
		data.DateTo, _ = time.Parse("2006-01-02", p.To)
	}
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
	return nil
}
func (s *LogService) CancelRequest(param tk.M) error {
	data := NewLogLeaveRemoteModel()
	userrepo := repositories.UserOrmRepo{}
	if s.TypeRequest == "leave" {
		p := s.DataLeave
		data.TypeRequest = "leave"
		if p.IsEmergency {
			data.TypeRequest = "emergency leave"
		}
		data.IdRequest = p.Id
		data.Name = p.Name
		data.Userid = p.UserId
		data.Email = p.Email
		data.Project = p.Project
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.LeaveFrom)
		data.DateTo, _ = time.Parse("2006-01-02", p.LeaveTo)
	} else {
		p := s.DataRemote
		data.TypeRequest = "remote"
		data.IdRequest = p.IdOp
		data.Name = p.Name
		data.Userid = p.UserId
		for _, pro := range p.Projects {
			data.Project = append(data.Project, pro.ProjectName)
		}
		user, err := userrepo.GetByID(p.UserId)
		if err != nil {
			return err
		}
		data.Email = user.Email
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.From)
		data.DateTo, _ = time.Parse("2006-01-02", p.To)
	}
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
	return nil
}
func (s *LogService) RequestExpired(param tk.M) error {
	data := NewLogLeaveRemoteModel()
	userrepo := repositories.UserOrmRepo{}
	if s.TypeRequest == "leave" {
		p := s.DataLeave
		data.TypeRequest = "leave"
		if p.IsEmergency {
			data.TypeRequest = "emergency leave"
		}
		data.IdRequest = p.Id
		data.Name = p.Name
		data.Userid = p.UserId
		data.Email = p.Email
		data.Project = p.Project
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.LeaveFrom)
		data.DateTo, _ = time.Parse("2006-01-02", p.LeaveTo)
	} else {
		p := s.DataRemote
		data.TypeRequest = "remote"
		data.IdRequest = p.IdOp
		data.Name = p.Name
		data.Userid = p.UserId
		for _, pro := range p.Projects {
			data.Project = append(data.Project, pro.ProjectName)
		}
		user, err := userrepo.GetByID(p.UserId)
		if err != nil {
			return err
		}
		data.Email = user.Email
		data.Location = p.Location
		data.DateFrom, _ = time.Parse("2006-01-02", p.From)
		data.DateTo, _ = time.Parse("2006-01-02", p.To)
	}
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
	return nil
}
