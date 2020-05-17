package controllers

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	//	"strings"

	gomail "gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2/bson"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	//	"gopkg.in/mgo.v2/bson"
)

type LoginController struct {
	*BaseController
}

const RandString = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (c *LoginController) Default(k *knot.WebContext) interface{} {
	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.LayoutTemplate = ""
	return ""
}
func (c *DashboardController) DefaultMobile(k *knot.WebContext) interface{} {
	setConf := []string{
		"_modal.html",
		"_loader.html",
	}
	getDefault := c.SetTypeDevice(k, setConf)
	return getDefault
}
func (c *DashboardController) SetTypeDevice(k *knot.WebContext, setConf []string) tk.M {
	c.LoadBase(k)
	viewData := tk.M{}
	if k.Session("jobrolename") != nil {
		viewData.Set("JobRoleName", k.Session("jobrolename").(string))
		viewData.Set("UserId", k.Session("userid").(string))
		viewData.Set("JobRoleLevel", k.Session("jobrolelevel"))
	} else {
		viewData.Set("JobRoleName", "")
		viewData.Set("JobRoleLevel", "")
		viewData.Set("UserId", "")
	}

	DataAccess := c.SetViewData(k, viewData)

	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.IncludeFiles = setConf
	return DataAccess
}
func (c *LoginController) Do(k *knot.WebContext) interface{} {
	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputJson
	formData := struct {
		UserName   string
		Password   string
		RememberMe bool
	}{}
	// message := ""
	// isValid := false
	err := k.GetPayload(&formData)
	if err != nil {
		c.WriteLog(err)
		// message = "Backend Error " + err.Error()
		return c.SetResultInfo(true, err.Error(), nil)
	}
	q := tk.M{}.Set("where", db.Eq("username", formData.UserName))
	cur, err := c.Ctx.Find(new(SysUserModel), q)
	if err != nil {
		return tk.M{}.Set("Valid", false).Set("Message", err.Error())
	}
	res := make([]SysUserModel, 0)
	if cur != nil {
		defer cur.Close()
	} else {
		return nil
	}
	//	defer c.Ctx.Close()
	// defer cur.Close()
	err = cur.Fetch(&res, 0, false)
	if err != nil {
		return tk.M{}.Set("Valid", false).Set("Message", err.Error())
	}

	// fmt.Println("------------", res)
	if len(res) > 0 {
		resUser := res[0]
		var valProjRole ProjectRuleModel
		if resUser.ProjectRuleID != "" {
			where := tk.M{}.Set("where", db.Eq("_id", bson.ObjectIdHex(resUser.ProjectRuleID)))
			cur, err := c.Ctx.Find(new(ProjectRuleModel), where)
			if err != nil {
				return tk.M{}.Set("Valid", false).Set("Message", err.Error())
			}
			resProjRole := make([]ProjectRuleModel, 0)
			if cur != nil {
				defer cur.Close()
			} else {
				return nil
			}
			err = cur.Fetch(&resProjRole, 0, false)
			if err != nil {
				return tk.M{}.Set("Valid", false).Set("Message", err.Error())
			}
			if len(resProjRole) > 0 {
				valProjRole.Name = resProjRole[0].Name
				valProjRole.Level = resProjRole[0].Level
			}
		}

		if helper.GetMD5Hash(formData.Password) == resUser.Password {
			if resUser.Enable == true {

				resroles := make([]SysRolesModel, 0)
				crsR, errR := c.Ctx.Find(new(SysRolesModel), tk.M{}.Set("where", db.Eq("name", resUser.Roles)))
				if crsR != nil {
					defer crsR.Close()
				} else {
					return c.SetResultInfo(true, "error query", nil)
				}
				if errR != nil {
					return c.SetResultInfo(true, errR.Error(), nil)
				}
				errR = crsR.Fetch(&resroles, 0, false)
				if errR != nil {
					return c.SetResultInfo(true, errR.Error(), nil)
				}
				// defer crsR.Close()

				k.SetSession("userid", string(resUser.Id))
				k.SetSession("empid", resUser.EmpId)
				k.SetSession("username", resUser.Username)
				k.SetSession("usermodel", resUser)
				k.SetSession("jobrolename", valProjRole.Name)
				k.SetSession("jobrolelevel", valProjRole.Level)
				k.SetSession("roles", resroles)
				k.SetSession("location", resUser.Location)
				// isValid = true
				// k.SetSession("onremote", strconv.FormatBool(resUser.Remote.RemoteActive))
				// k.SetSession("conditionalremote", resUser.Remote.ConditionalRemote)
				// k.SetSession("fullmonth", strconv.FormatBool(resUser.Remote.FullMonth))
				// k.SetSession("monthly", strconv.FormatBool(resUser.Remote.Monthly))
			} else {
				// message = "Your account is disabled, please contact administrator to enable it."
				return c.SetResultInfo(true, "Your account is disabled, please contact administrator to enable it.", nil)
			}
		} else {
			// message = "Invalid Username or password!"
			return c.SetResultInfo(true, "Invalid Username or password!", nil)
		}
	} else {
		// return "Invalid Username or password!"
		return c.SetResultInfo(true, "Invalid Username or password!", nil)
	}

	return c.SetResultInfo(false, "login success", nil)
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func (c *LoginController) ForgotPswSendEmail(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := struct {
		Email string
	}{}

	err := k.GetPayload(&payload)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	randStr := String(15)

	// tk.Println("---------------", randStr)

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*SysUserModel, 0)

	dbFilter = append(dbFilter, db.Eq("email", payload.Email))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewSysUserModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return c.SetResultInfo(true, "error query", nil)
	}
	// defer crs.Close()

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	err = crs.Fetch(&data, 0, false)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if len(data) == 0 {
		return c.SetResultInfo(true, "no data email found", nil)
	}

	// else {
	// 	// data[0].Password = helper.GetMD5Hash(data[0].Password)
	// 	// err = c.Ctx.Save(data[0])
	// 	if err != nil {
	// 		return c.SetResultInfo(true, err.Error(), nil)
	// 	}
	// }

	c.SendEmailForgotPsw(k, data, payload.Email, randStr)

	return c.SetResultInfo(false, "Please check your email", data)

}

func (c *LoginController) SignUpSendEmail(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := struct {
		Email string
	}{}

	err := k.GetPayload(&payload)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	randStr := String(15)

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*SysUserModel, 0)

	usr := new(SysUserModel)

	dbFilter = append(dbFilter, db.Eq("username", payload.Email))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewSysUserModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return c.SetResultInfo(true, "error query", nil)
	}
	// defer crs.Close()

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	err = crs.Fetch(&data, 0, false)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if len(data) != 0 {
		return c.SetResultInfo(true, "Your email already registered", nil)
	}

	usr.Id = bson.NewObjectId().Hex()
	usr.EmpId = ""
	usr.Designation = ""
	usr.Username = payload.Email
	usr.Fullname = payload.Email
	usr.Enable = true
	usr.PhoneNumber = ""
	usr.Email = ""
	usr.YearLeave = 12
	usr.PublicLeave = 12
	usr.Roles = "dasboard-user"
	Password := helper.GetMD5Hash(string(randStr))
	usr.Password = Password
	usr.Location = ""
	usr.IsChangePassword = false
	usr.AccesRight = ""

	//check condition exist in registeruser -- start
	queryRegUser := c.Ctx.Connection.NewQuery()
	resultRegUser := []RegisterUserModel{}
	csr, e := queryRegUser.
		Select().From("RegisterUser").Where(db.Eq("IsRegister", true)).Cursor(nil)
	e = csr.Fetch(&resultRegUser, 0, false)

	if e != nil {

	}
	defer csr.Close()
	var stsRegisterUser = false
	if len(resultRegUser) > 0 {
		for _, val := range resultRegUser {
			if val.UserID == usr.Username {
				stsRegisterUser = true
			}
		}
	}

	//check condition exist in registeruser -- end
	if stsRegisterUser {
		err = c.Ctx.Save(usr)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
		c.SendEmailSignUp(k, usr, payload.Email, string(randStr))
		return c.SetResultInfo(true, "Please check your email", data)
	}
	_ = new(RegisterUserController).SaveRegisterUserByLogin(usr.Username, c.Ctx)

	return c.SetResultInfo(true, "Please contact your administrator \nfor get confirmation email", data)

}

func (c *LoginController) SendEmailForgotPsw(k *knot.WebContext, dataUser []*SysUserModel, mail string, randomString string) interface{} {
	k.Config.OutputType = knot.OutputJson
	sConf := helper.ReadConfig()

	// sUrl := k.Server.Address

	urlResetPsw := sConf.GetString("BaseUrlEmail") + "/mail/resetpassword"

	mailParam := struct {
		UserId string
	}{}

	// config := ReadConfig()

	mailParam.UserId = dataUser[0].Id
	paramApp, _ := json.Marshal(mailParam)

	urlReset, _ := http.NewRequest("Get", urlResetPsw, nil)
	inReset := urlReset.URL.Query()
	inReset.Add("param", GCMEncrypter(string(paramApp)))
	urlReset.URL.RawQuery = inReset.Encode()
	// mailServer := k.Config
	conf := gomail.NewPlainDialer("smtp.office365.com", 587, "admin.support@creativelab.com", "DFOP4vfsiOw1roNZ")

	m := gomail.NewMessage()
	mailsubj := tk.Sprintf("%v", "Reset Password")
	m.SetHeader("Subject", mailsubj)
	m.SetHeader("From", "admin.support@creativelab.com")
	m.SetHeader("To", mail)

	mailContain := map[string]string{
		"to": dataUser[0].Fullname, "urlLink": urlReset.URL.String(),
	}
	mailController := MailController(*c)
	bd, er := mailController.File("resetpassword.html", mailContain)
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

func (c *LoginController) SendEmailSignUp(k *knot.WebContext, dataUser *SysUserModel, mail string, randomString string) interface{} {
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

func (c *LoginController) ProcessReset(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Password string
		Param    string
	}{}

	payload := new(ParameterResetPassword)

	err := k.GetPayload(&p)
	if err != nil {
		c.SetResultInfo(true, err.Error(), nil)
	}

	decript := GCMDecrypter(p.Param)
	json.Unmarshal([]byte(decript), payload)
	// tk.Println("------------", payload)
	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*SysUserModel, 0)

	dbFilter = append(dbFilter, db.Eq("_id", payload.UserId))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewSysUserModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return c.SetResultInfo(true, "error query", nil)
	}
	// defer crs.Close()

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	err = crs.Fetch(&data, 0, false)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if len(data) == 0 {
		return c.SetResultInfo(true, "no data email found", nil)
	}
	p.Password = helper.GetMD5Hash(p.Password)
	data[0].Password = p.Password

	err = c.Ctx.Save(data[0])
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	return c.SetResultInfo(false, "Data successfully reset", nil)
}
