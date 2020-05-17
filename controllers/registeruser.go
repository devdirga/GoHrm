package controllers

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"

	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	"github.com/creativelab/orm"
	tk "github.com/creativelab/toolkit"
	xl "github.com/tealeg/xlsx"
	gomail "gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2/bson"
)

type RegisterUserController struct {
	*BaseController
}

func (c *RegisterUserController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	dataRegisterUser := make([]RegisterUserModel, 0)
	crsRegisterUser, errRegisterUser := c.Ctx.Find(NewRegisterUserModel(), nil)

	if crsRegisterUser != nil {
		defer crsRegisterUser.Close()
	} else if crsRegisterUser == nil {
		return c.SetResultInfo(true, "Error when build query", nil)
	}
	defer crsRegisterUser.Close()
	if errRegisterUser != nil {
		return c.SetResultInfo(true, errRegisterUser.Error(), nil)
	}

	errRegisterUser = crsRegisterUser.Fetch(&dataRegisterUser, 0, false)
	if errRegisterUser != nil {
		return c.SetResultInfo(true, errRegisterUser.Error(), nil)
	}
	//get data Project Rule -- start
	dataProjRule := make([]ProjectRuleModel, 0)
	crsProjRule, errProjRule := c.Ctx.Find(NewProjectRuleModel(), nil)

	if crsProjRule != nil {
		defer crsProjRule.Close()
	}
	defer crsProjRule.Close()
	if errProjRule != nil {
		tk.Println(errProjRule)
	}

	errProjRule = crsProjRule.Fetch(&dataProjRule, 0, false)
	if errProjRule != nil {
		tk.Println(errProjRule)
	}
	//get data Project Rule -- end

	return dataRegisterUser
}
func (c *RegisterUserController) SaveRegisterUser(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := new(RegisterUserModel)
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if payload.Id == "" {
		payload.Id = bson.NewObjectId().Hex()
	}

	err = c.Ctx.Save(payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	return c.SetResultInfo(false, "Data has been save", nil)
}
func (c *RegisterUserController) SaveRegisterUserByLogin(UserID string, Ctx *orm.DataContext) interface{} {
	payload := RegisterUserModel{}
	payload.UserID = UserID
	payload.IsRegister = false

	if payload.Id == "" {
		payload.Id = bson.NewObjectId().Hex()
	}

	_ = Ctx.Save(&payload)
	return nil
}

func (c *RegisterUserController) RegisterEmail(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := struct {
		Email []string
		ID    []string
	}{}

	err := k.GetPayload(&payload)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	rl := ProjectRuleController(*c)
	rule, err := rl.GetProjectRule(k)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// tk.Println("-------------- ", rule[2].Id.Hex())

	randStr := String(15)

	for key, val := range payload.Email {
		usr := new(SysUserModel)
		usr.Id = bson.NewObjectId().Hex()
		usr.EmpId = ""
		usr.Designation = ""
		usr.Username = val
		usr.Fullname = val
		usr.Enable = true
		usr.PhoneNumber = ""
		usr.Email = ""
		usr.YearLeave = 12
		usr.PublicLeave = 12
		usr.Roles = "dasboard-user"
		Password := helper.GetMD5Hash(string(randStr))
		usr.Password = Password
		usr.Location = "Indonesia"
		usr.IsChangePassword = false
		usr.AccesRight = ""
		usr.ProjectRuleID = rule[2].Id.Hex()
		usr.ProjectRuleName = rule[2].Name
		err = c.Ctx.Save(usr)
		if err != nil {
			continue
		}
		c.SendEmailSignUp(k, usr, val, string(randStr))
		var inputRegisterModel RegisterUserModel
		inputRegisterModel.Id = payload.ID[key]
		inputRegisterModel.UserID = val
		inputRegisterModel.IsRegister = true
		c.SaveRegisterUserByAdmin(inputRegisterModel)
	}

	return c.SetResultInfo(true, "Data already Save", nil)
}

func (c *RegisterUserController) SendEmailSignUp(k *knot.WebContext, dataUser *SysUserModel, mail string, randomString string) interface{} {
	k.Config.OutputType = knot.OutputJson
	sConf := helper.ReadConfig()

	urlResetPsw := sConf.GetString("BaseUrlEmail") + "/login/default"

	// tk.Println("-------------------- ", dataUser.Fullname)

	// config := ReadConfig()

	// mailServer := k.Config
	conf := gomail.NewPlainDialer("smtp.office365.com", 587, "admin.support@creativelab.com", "DFOP4vfsiOw1roNZ")

	m := gomail.NewMessage()
	mailsubj := tk.Sprintf("%v", "Sign up")
	m.SetHeader("Subject", mailsubj)
	m.SetHeader("From", "admin.support@creativelab.com")
	m.SetHeader("To", mail)

	mailContain := map[string]string{
		"to": dataUser.Username, "password": randomString, "urlLink": urlResetPsw,
	}
	mailController := MailController(*c)
	bd, er := mailController.File("signup.html", mailContain)
	if er != nil {
		return c.SetResultInfo(true, er.Error(), nil)
	}
	m.SetBody("text/html", bd)
	if err := conf.DialAndSend(m); err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	m.Reset()

	return c.SetResultInfo(false, "successfully send email", nil)
}

func (c *RegisterUserController) SaveRegisterUserByAdmin(param RegisterUserModel) interface{} {
	payload := RegisterUserModel{}
	payload.Id = param.Id
	payload.UserID = param.UserID
	payload.IsRegister = param.IsRegister

	_ = c.Ctx.Save(&payload)
	return nil
}

func (c *RegisterUserController) UploadFile(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	reader, err := k.Request.MultipartReader()
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	var fileLocation, fileName string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		fileName = tk.RandomString(32) + filepath.Ext(part.FileName())
		fileLocation = filepath.Join("assets/doc/", fileName)
		dst, err := os.Create(fileLocation)
		if dst != nil {
			defer dst.Close()
		}
		if err != nil {
			return c.SetResultError(err.Error(), nil)
		}

		if _, err := io.Copy(dst, part); err != nil {
			return c.SetResultError(err.Error(), nil)
		}
	}

	xlFile, err := xl.OpenFile(fileLocation)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	if len(xlFile.Sheets) == 0 {
		return c.SetResultError("Excel contains no sheet", nil)
	}
	//get data Project Rule -- start
	dataProjRule := make([]ProjectRuleModel, 0)
	crsProjRule, errProjRule := c.Ctx.Find(NewProjectRuleModel(), nil)

	if crsProjRule != nil {
		defer crsProjRule.Close()
	}
	defer crsProjRule.Close()
	if errProjRule != nil {
		tk.Println(errProjRule)
	}

	errProjRule = crsProjRule.Fetch(&dataProjRule, 0, false)
	if errProjRule != nil {
		tk.Println(errProjRule)
	}
	//get data Project Rule -- end
	isNoRows := true
	for iRow, row := range xlFile.Sheets[0].Rows {
		if iRow == 0 {
			continue
		}

		if len(row.Cells) == 0 {
			continue
		}

		isNoRows = false

		var ID, UserID, EmpID, Name, RoleName string

		if len(row.Cells) > 0 {
			ID = row.Cells[0].String()

		}
		if len(row.Cells) > 1 {
			UserID = row.Cells[1].String()
		}
		if len(row.Cells) > 2 {
			EmpID = row.Cells[2].String()
		}
		if len(row.Cells) > 3 {
			Name = row.Cells[3].String()
		}
		if len(row.Cells) > 4 {
			RoleName = row.Cells[4].String()
		}

		if strings.TrimSpace(ID) == "" && strings.TrimSpace(UserID) == "" {
			continue
		}

		if strings.TrimSpace(ID) == "" {
			ID = bson.NewObjectId().Hex()
		}
		var findProjRuleID string
		var valProjRule = strings.ToLower(strings.TrimSpace(RoleName))
		if RoleName != "" {
			for _, valData := range dataProjRule {
				if strings.ToLower(valData.Name) == valProjRule || strings.ToLower(valData.AliasName) == valProjRule {

					findProjRuleID = valData.Id.Hex()
					continue

				}
			}
		}

		payload := new(RegisterUserModel)
		payload.Id = ID
		payload.UserID = UserID
		payload.IsRegister = true
		payload.EmpID = EmpID
		payload.Name = Name
		payload.RoleID = findProjRuleID
		err = c.Ctx.Save(payload)
		if err != nil {

		}

	}

	if isNoRows {
		return c.SetResultError("Excel contains no data", nil)
	}

	return c.SetResultOK("Ok")
}

func (c *RegisterUserController) Delete(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := new(RegisterUserModel)
	err := k.GetPayload(&payload)
	queryRegUser := c.Ctx.Connection.NewQuery()

	err = queryRegUser.
		From(new(RegisterUserModel).TableName()).
		Delete().
		Where(dbox.Eq("_id", payload.Id)).
		Exec(nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *RegisterUserController) GetDataRegUser(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	p := k.Session("username")
	data := make([]*RegisterUserModel, 0)
	var dbFilter []*dbox.Filter
	if p != nil {

		dbFilter = append(dbFilter, dbox.Eq("UserID", p))

		query := tk.M{}

		if len(dbFilter) > 0 {
			query.Set("where", dbox.And(dbFilter...))
		}

		crs, errdata := c.Ctx.Find(NewRegisterUserModel(), query)
		defer crs.Close()
		if errdata != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}
		errdata = crs.Fetch(&data, 0, false)
		if errdata != nil {
			return c.SetResultInfo(true, errdata.Error(), nil)
		}
	}

	return data
}
