package services

import (
	"bytes"
	"creativelab/ecleave-dev/helper"
	"html/template"
	"path/filepath"
	"time"

	gomail "gopkg.in/gomail.v2"
)

type MailService struct {
	Conf             MailConf
	MailSubject      string
	From             string
	To               []string
	Filename         string
	FileParamRequest MailFileParamRequest
}

type MailFileParamRequest struct {
	DateCreate     string
	LeaderProject  string
	NameEmployee   string
	Reason         string
	TypeOfRemote   string
	ListProject    []string
	RemoteFrom     string
	RemoteTo       string
	ListDate       []string
	DetailsDate    []DetailDate
	UrlDecline     string
	UrlApproval    string
	ManagerProject string
	Status         string
	Note           string
	DayDuration    string
}
type MailConf struct {
	EmailHost             string
	EmailPort             int
	EmailOperator         string
	EmailOperatorPassword string
	FileTemplatePath      string
}

type DetailDate struct {
	DateLeave string
	Status    string
	Note      string
}

func (s *MailService) Init() {
	config := helper.ReadConfig()

	s.Conf.EmailHost = "smtp.office365.com"
	s.Conf.EmailPort = 587
	s.Conf.EmailOperator = "hrd@creativelab.com"
	s.Conf.EmailOperatorPassword = "KnxNP5aO"
	s.Conf.FileTemplatePath = config.GetString("EmailTemplatePath")
}

func DelayProcess(n time.Duration) {
	time.Sleep(n * time.Second)
}

func (s *MailService) SendEmail() error {
	conf := gomail.NewPlainDialer(s.Conf.EmailHost, s.Conf.EmailPort, s.Conf.EmailOperator, s.Conf.EmailOperatorPassword)
	_ = conf
	mailsubj := s.MailSubject
	//setHeader
	m := gomail.NewMessage()
	defer m.Reset()

	m.SetHeader("From", s.From)
	m.SetHeader("To", s.To...)
	m.SetHeader("Subject", mailsubj)

	template, err := s.ReadTemplate()
	if err != nil {
		return err
	}

	m.SetBody("text/html", string(template))

	DelayProcess(5)
	err = conf.DialAndSend(m)
	return err
}

func (s *MailService) ReadTemplate() ([]byte, error) {
	templ, err := template.ParseFiles(filepath.Join(s.Conf.FileTemplatePath, s.Filename))
	param := s.FileParamRequest
	body := []byte{}

	if err != nil {
		return body, err
	}

	buffer := new(bytes.Buffer)
	if err = templ.Execute(buffer, param); err != nil {
		return body, err
	}

	body = buffer.Bytes()

	return body, nil
}
