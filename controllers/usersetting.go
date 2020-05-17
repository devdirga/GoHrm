package controllers

import (
	"creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"creativelab/ecleave-dev/repositories"
	"creativelab/ecleave-dev/services"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	db "github.com/creativelab/dbox"
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	xl "github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2/bson"
)

type UserSettingController struct {
	*BaseController
}

type SortPaging struct {
	Take int
	Skip int
}

func (c *UserSettingController) Default(k *knot.WebContext) interface{} {
	access := c.LoadBase(k)
	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	DataAccess := Previlege{}

	for _, o := range access {
		DataAccess.Create = o["Create"].(bool)
		DataAccess.View = o["View"].(bool)
		DataAccess.Delete = o["Delete"].(bool)
		DataAccess.Process = o["Process"].(bool)
		DataAccess.Delete = o["Delete"].(bool)
		DataAccess.Edit = o["Edit"].(bool)
		DataAccess.Menuid = o["Menuid"].(string)
		DataAccess.Menuname = o["Menuname"].(string)
		DataAccess.Approve = o["Approve"].(bool)
		DataAccess.Username = o["Username"].(string)
	}

	return DataAccess
}

func (c *UserSettingController) EditProfile(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputTemplate

	viewData := c.SetViewData(k, nil)
	return viewData
}

func (c *UserSettingController) GetUserLogin(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.IsAuthenticate(k)

	userId := k.Session("userid")
	if userId == nil {
		return c.SetResultError("not have credential", nil)
	}

	user, err := new(services.UserService).GetByID(userId.(string))
	if err != nil {
		return c.SetResultError(err.Error(), user)
	}

	return c.SetResultOK(user)
}
func (c *UserSettingController) GetUserProfileLogin(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.IsAuthenticate(k)

	userId := k.Session("userid")
	if userId == nil {
		return c.SetResultError("not have credential", nil)
	}

	user, err := new(services.UserService).GetProfileByID(userId.(string))
	if err != nil {
		return c.SetResultError(err.Error(), user)
	}

	return c.SetResultOK(user)
}

func (d *UserSettingController) GetData_byId(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	//ret := ResultInfo{}

	p := struct {
		Id string
	}{}

	err := r.GetPayload(&p)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	var dbFilter []*db.Filter
	if p.Id != "" {
		dbFilter = append(dbFilter, db.Eq("_id", p.Id))
	}

	query := tk.M{}
	data := make([]*SysUserModel, 0)
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, errdata := d.Ctx.Find(NewSysUserModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return d.SetResultInfo(true, "error on query", nil)
	}
	// defer crs.Close()
	if errdata != nil {
		return d.SetResultInfo(true, errdata.Error(), nil)
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return d.SetResultInfo(true, errdata.Error(), nil)
	}

	return data
}

func (d *UserSettingController) GetDataUser(r *knot.WebContext, id string) ([]*SysUserModel, error) {
	r.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	if id != "" {
		dbFilter = append(dbFilter, db.Eq("_id", id))
	}

	query := tk.M{}
	data := []*SysUserModel{}
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	} else {
		query = nil
	}

	crs, errdata := d.Ctx.Find(NewSysUserModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil, nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil, errdata
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil, errdata
	}

	return data, nil
}

func (d *UserSettingController) GetDataOptionUser(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	p := struct {
		Id string
	}{}

	err := r.GetPayload(&p)

	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	data, err := d.GetOptionUser(r, p.Id)

	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	if len(data) > 0 {
		return d.SetResultInfo(false, "success", data)
	}

	return d.SetResultInfo(true, "data empty", nil)
}

func (d *UserSettingController) GetOptionUser(r *knot.WebContext, userid string) ([]*ChangeOptionModel, error) {
	r.Config.OutputType = knot.OutputJson

	var dbFilter []*db.Filter
	if userid != "" {
		dbFilter = append(dbFilter, db.Eq("userid", userid))
	}

	query := tk.M{}
	data := []*ChangeOptionModel{}
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	} else {
		query = nil
	}

	crs, errdata := d.Ctx.Find(NewChangeOptionModel(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil, nil
	}
	// defer crs.Close()
	if errdata != nil {
		return nil, errdata
	}

	errdata = crs.Fetch(&data, 0, false)
	if errdata != nil {
		return nil, errdata
	}

	return data, nil
}

func (d *UserSettingController) SaveAllOptRemote(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	p := struct {
		Remote      bool
		Monthly     bool
		FullMonth   bool
		Conditional int
	}{}

	err := r.GetPayload(&p)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	tk.Println("---------- ", p)

	dataUser, err := d.GetDataUser(r, "")
	tk.Println("---------- ", dataUser)

	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	for _, dt := range dataUser {
		dataOption, err := d.GetOptionUser(r, dt.Id)

		if err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		}

		tk.Println("---------- ", dataOption)

		if len(dataOption) > 0 {
			tk.Println("--------- masuk ")
			for _, dop := range dataOption {
				//
				dop.Name = dt.Fullname
				dop.Email = dt.Email
				dop.Remote.RemoteActive = p.Remote
				dop.Remote.Monthly = p.Monthly
				dop.Remote.FullMonth = p.FullMonth
				dop.Remote.ConditionalRemote = p.Conditional

				err = d.Ctx.Save(dop)

				if err != nil {
					return d.SetResultInfo(true, err.Error(), nil)
				}
			}
		} else {
			opt := ChangeOptionModel{}

			opt.Id = bson.NewObjectId().Hex()
			opt.UserId = dt.Id
			opt.Name = dt.Fullname
			opt.Email = dt.Email
			opt.Remote.RemoteActive = p.Remote
			opt.Remote.Monthly = p.Monthly
			opt.Remote.FullMonth = p.FullMonth
			opt.Remote.ConditionalRemote = p.Conditional

			err = d.Ctx.Save(&opt)

			if err != nil {
				return d.SetResultInfo(true, err.Error(), nil)
			}
		}

	}

	optUser, err := d.GetOptionUser(r, "")
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	mail := MailController(*d)
	mail.SendMailUserOption(r, optUser)

	return d.SetResultInfo(false, "Remote Option Successfully saved", nil)
}

func (d *UserSettingController) SaveOptionRemote(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	p := struct {
		Name        string
		Userid      string
		Email       string
		Remote      bool
		Monthly     bool
		FullMonth   bool
		Conditional int
	}{}

	err := r.GetPayload(&p)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	optionUser, err := d.GetOptionUser(r, p.Userid)

	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	tk.Println("--------- ", p)
	if len(optionUser) > 0 {
		optionUser[0].Remote.RemoteActive = p.Remote
		optionUser[0].Remote.Monthly = p.Monthly
		optionUser[0].Remote.FullMonth = p.FullMonth
		optionUser[0].Remote.ConditionalRemote = p.Conditional

		err = d.Ctx.Save(optionUser[0])

		if err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		}

		mail := MailController(*d)
		mail.SendMailUserOption(r, optionUser)

	} else {
		opt := ChangeOptionModel{}
		opted := []*ChangeOptionModel{}

		opt.Id = bson.NewObjectId().Hex()
		opt.Name = p.Name
		opt.UserId = p.Userid
		opt.Email = p.Email
		opt.Remote.RemoteActive = p.Remote
		opt.Remote.Monthly = p.Monthly
		opt.Remote.FullMonth = p.FullMonth
		opt.Remote.ConditionalRemote = p.Conditional

		err = d.Ctx.Save(&opt)

		if err != nil {
			return d.SetResultInfo(true, err.Error(), nil)
		}

		opted = append(opted, &opt)

		tk.Println("------------- opted ", opted)

		mail := MailController(*d)
		mail.SendMailUserOption(r, opted)
	}

	return d.SetResultInfo(false, "Remote Option Successfully saved", nil)
}

func (d *UserSettingController) GetData(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}

	oo := struct {
		Id       string
		Username []interface{}
		Role     []interface{}
		Status   bool
		Take     int
		Skip     int
		Sort     []tk.M
	}{}
	//get data Project Rule -- start
	dataProjRule := make([]ProjectRuleModel, 0)
	crsProjRule, errProjRule := d.Ctx.Find(NewProjectRuleModel(), nil)

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

	err := r.GetPayload(&oo)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}
	var dbFilter []*db.Filter
	if oo.Id != "" {
		dbFilter = append(dbFilter, db.Eq("_id", bson.ObjectIdHex(oo.Id)))
	} else {
		dbFilter = append(dbFilter, db.Eq("enable", oo.Status))
	}

	if len(oo.Username) != 0 {
		dbFilter = append(dbFilter, db.In("username", oo.Username...))
	}

	if len(oo.Role) != 0 {
		dbFilter = append(dbFilter, db.In("roles", oo.Role...))
	}

	sort := ""
	dir := ""
	if len(oo.Sort) > 0 {
		sort = strings.ToLower(oo.Sort[0].Get("field").(string))
		dir = oo.Sort[0].Get("dir").(string)
	}

	if sort == "" {
		sort = "username"
	}
	if dir != "" && dir != "asc" {
		sort = "-" + sort
	}

	queryTotal := tk.M{}
	query := tk.M{}
	data := make([]SysUserModel, 0)
	total := make([]SysUserModel, 0)
	retModel := tk.M{}
	query.Set("limit", oo.Take)
	query.Set("skip", oo.Skip)
	query.Set("order", []string{sort})

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
		queryTotal.Set("where", db.And(dbFilter...))
	}

	crsData, errData := d.Ctx.Find(NewSysUserModel(), query)
	if crsData != nil {
		defer crsData.Close()
	} else {
		return d.SetResultInfo(true, "error on query", nil)
	}
	// defer crsData.Close()
	if errData != nil {
		return d.SetResultInfo(true, errData.Error(), nil)
	}
	errData = crsData.Fetch(&data, 0, false)

	if errData != nil {
		return d.SetResultInfo(true, errData.Error(), nil)
	} else {
		for idx, val := range data {
			projRuleID := val.ProjectRuleID

			var findProjectRuleName string
			if projRuleID != "" {
				for _, valData := range dataProjRule {
					if valData.Id.Hex() == projRuleID {
						findProjectRuleName = valData.Name
						continue

					}
				}
			}
			data[idx].ProjectRuleName = findProjectRuleName
			data[idx].Password = ""

		}
		retModel.Set("Records", data)
	}
	crsTotal, errTotal := d.Ctx.Find(NewSysUserModel(), queryTotal)
	// defer crsTotal.Close()
	if errTotal != nil {
		return d.SetResultInfo(true, errTotal.Error(), nil)
	}
	errTotal = crsTotal.Fetch(&total, 0, false)

	if errTotal != nil {
		return d.SetResultInfo(true, errTotal.Error(), nil)
	} else {
		retModel.Set("Count", len(total))
	}

	ret.Data = retModel

	return ret
}

func (c *UserSettingController) SaveData(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	p := new(SysUserModel)
	e := k.GetPayload(&p)

	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	if p.Id == "" {
		p.Id = bson.NewObjectId().Hex()
		p.Password = helper.GetMD5Hash(p.Password)
		p.AddLeave = ""

		er := c.AddNewOpt(k, p.Id, p.Email, p.Fullname)

		if er != nil {
			return c.SetResultInfo(true, er.Error(), nil)
		}
	} else {
		userrepo := repositories.UserOrmRepo{}
		user, err := userrepo.GetByID(p.Id)
		if err != nil {
			return err
		}

		if user.Password != p.Password {
			p.Password = helper.GetMD5Hash(p.Password)
		}

		p.Departement = user.Departement
		p.Address = user.Address
		p.Gender = user.Gender
		tm := time.Now()
		ms, _ := time.Parse("2006-01-02", p.JointDate)
		// if tm.Year() == ms.Year() {
		next := tm.AddDate(0, 1, 0)
		// snow := next.Format("2006-01-02")
		temp := time.Date(tm.Year(), next.Month(), ms.Day(), 0, 0, 0, 0, time.UTC)
		p.AddLeave = temp.Format("2006-01-02")
		// }

		p.DecYear = float64(p.YearLeave)
	}

	err := c.Ctx.Save(p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	return c.SetResultInfo(false, "data has been saved", nil)
}

func (c *UserSettingController) AddNewOpt(k *knot.WebContext, id string, email string, name string) error {
	k.Config.OutputType = knot.OutputJson

	p := ChangeOptionModel{}

	p.Id = bson.NewObjectId().Hex()
	p.UserId = id
	p.Email = email
	p.Name = name
	p.Remote.RemoteActive = false
	p.Remote.Monthly = false
	p.Remote.FullMonth = false
	p.Remote.ConditionalRemote = 0

	err := c.Ctx.Save(&p)

	if err != nil {
		return err
	}

	return nil
}

func (c *UserSettingController) UploadImage(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	path, err := new(services.UserService).UploadImage(k)
	if err != nil {
		return c.SetResultError(err.Error(), err)
	}

	return c.SetResultOK(tk.M{}.Set("ImagePath", path))
}

func (c *UserSettingController) CheckImgExist(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		Filename string
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	pathImg := new(services.UserService).CheckImageExist(payload.Filename)
	return c.SetResultOK(pathImg)
}

func (c *UserSettingController) Save(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	payload := SysUserModel{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	user, err := new(services.UserService).Save(payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(user)
}

func (c *UserSettingController) SaveUserProfile(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	payload := SysUserModel{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	user, err := new(services.UserService).SaveUserProfile(payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(user)
}
func (c *UserSettingController) UpdateUserProfile(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := SysUserModel{}
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}
	payload.Id = k.Session("userid").(string)

	user, err := new(services.UserService).SaveUserProfileByClient(payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(user)
}

func (c *UserSettingController) Try(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	return new(repositories.LeaveDboxRepo).GetLastLeaveDate()
}
func (c *UserSettingController) UploadFile(k *knot.WebContext) interface{} {
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
	tk.Println("------------------ ", xlFile.Sheets[0].Rows)
	isNoRows := true
	for iRow, row := range xlFile.Sheets[0].Rows {
		if iRow == 0 {
			continue
		}

		if len(row.Cells) == 0 {
			continue
		}

		isNoRows = false

		var EmpID, Username, Fullname, Email, Jobrole, Roles, Designation, Jointdate, Location string
		var Publicleave, Yearleave int
		var Status bool
		if len(row.Cells) > 0 {
			EmpID = row.Cells[0].String()
		}
		if len(row.Cells) > 1 {
			Username = row.Cells[1].String()
		}
		if len(row.Cells) > 2 {
			Fullname = row.Cells[2].String()
		}
		if len(row.Cells) > 3 {
			Email = row.Cells[3].String()
		}
		if len(row.Cells) > 4 {
			Jobrole = row.Cells[4].String()
		}
		var findProjRuleID string
		var valProjRule = strings.ToLower(strings.TrimSpace(Jobrole))
		if Jobrole != "" {
			for _, valData := range dataProjRule {
				if strings.ToLower(valData.Name) == valProjRule || strings.ToLower(valData.AliasName) == valProjRule || valData.Id.Hex() == valProjRule {
					findProjRuleID = valData.Id.Hex()
					continue
				}
			}
		}
		if len(row.Cells) > 5 {
			Roles = row.Cells[5].String()
		}
		if len(row.Cells) > 6 {
			Designation = row.Cells[6].String()
		}
		if len(row.Cells) > 7 {
			tm, err := row.Cells[7].GetTime(false)
			if err != nil {
				if row.Cells[7].String() != "" {
					t, _ := time.Parse("2006-01-02", row.Cells[7].String())
					Jointdate = t.Format("2006-01-02")
				} else {
					Jointdate = ""
				}
			} else {
				Jointdate = tm.Format("2006-01-02")
			}
		}
		if len(row.Cells) > 8 {
			Location = row.Cells[8].String()
		}
		if len(row.Cells) > 9 {
			Publicleave, _ = row.Cells[9].Int()
		}
		if len(row.Cells) > 10 {
			Status = row.Cells[10].Bool()
		}
		if len(row.Cells) > 11 {
			Yearleave, _ = row.Cells[11].Int()
		}

		query := tk.M{}
		var dbFilter []*db.Filter
		dbFilter = append(dbFilter, db.Eq("empid", strings.TrimSpace(EmpID)))
		users := []SysUserModel{}
		query.Set("where", db.And(dbFilter...))
		crs, _ := c.Ctx.Find(NewSysUserModel(), query)
		if crs != nil {
			defer crs.Close()
		}
		err := crs.Fetch(&users, 0, false)

		if len(users) > 0 {
			users[0].Fullname = Fullname
			users[0].Username = Username
			users[0].Email = Email
			// handle empty ProjectRuleId
			if strings.TrimSpace(Jobrole) != "" {
				users[0].ProjectRuleID = findProjRuleID
			}
			// handle empty Roles
			if strings.TrimSpace(Roles) != "" {
				users[0].Roles = Roles
			}
			// handle empty JointDate
			if Jointdate != "" {
				users[0].JointDate = Jointdate
			}
			// handle empty Designation
			if strings.TrimSpace(Designation) != "" {
				users[0].Designation = Designation
			}
			// handle empty Location
			if strings.TrimSpace(Location) != "" {
				users[0].Location = Location
			}
			users[0].PublicLeave = Publicleave
			users[0].IsChangePassword = Status
			users[0].YearLeave = Yearleave
			users[0].DecYear = float64(Yearleave)
			err = c.Ctx.Save(&users[0])
			if err != nil {
			}
		} else {
			user := NewSysUserModel()
			user.JointDate = Jointdate
			user.EmpId = EmpID
			user.Username = Username
			user.Fullname = Fullname
			user.Email = Email
			user.Enable = true
			user.IsChangePassword = false
			user.YearLeave = Yearleave
			user.PublicLeave = Publicleave
			user.Roles = Roles
			user.ProjectRuleID = findProjRuleID
			user.YearLeave = Yearleave
			user.DecYear = float64(Yearleave)
			user.Password = "a8b6bb3b34339f5a6d439a0ef7fc7878"
			user.Roles = "dasboard-user"
			user.Designation = Designation
			user.Location = Location
			err = c.Ctx.Insert(user)
			if err != nil {
			}
		}

	}

	if isNoRows {
		return c.SetResultError("Excel contains no data", nil)
	}

	return c.SetResultOK("Ok")
}

// GetUsers ...
func (d *UserSettingController) GetUsers(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}
	pipe = append(pipe, tk.M{}.Set("$match", tk.M{}.Set("enable", true)))
	csr, e := d.Ctx.Connection.NewQuery().Select().From(NewSysUserModel().TableName()).Command("pipe", pipe).Cursor(nil)
	if e != nil {
		return d.SetResultInfo(true, e.Error(), nil)
	}
	defer csr.Close()
	data := []SysUserModel{}
	e = csr.Fetch(&data, 0, false)
	if e != nil {
		return d.SetResultInfo(true, e.Error(), nil)
	}
	return d.SetResultInfo(false, "Success", data)
}
