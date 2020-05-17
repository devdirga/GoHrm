package controllers

import (
	. "creativelab/ecleave-dev/models"
	//"encoding/json"
	//"fmt"
	db "github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type MenuSettingController struct {
	*BaseController
}

func (c *MenuSettingController) Default(k *knot.WebContext) interface{} {
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

func (c *MenuSettingController) GetMenuTop(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	payLoad := struct {
		Id string
	}{}
	err := k.GetPayload(&payLoad)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	var menuAccess []interface{}
	accesMenu := []SysRolesModel{}
	if k.Session("roles") != nil {
		accesMenu = k.Session("roles").([]SysRolesModel)
	}

	if accesMenu != nil && len(accesMenu) > 0 {
		for _, o := range accesMenu[0].Menu {
			if o.Access == true {
				menuAccess = append(menuAccess, o.Menuid)
			}
		}
	}

	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("Enable", true))
	dbFilter = append(dbFilter, db.In("_id", menuAccess...))

	queryTotal := tk.M{}
	query := tk.M{}
	data := make([]TopMenuModel, 0)
	total := make([]TopMenuModel, 0)
	retModel := tk.M{}

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
		queryTotal.Set("where", db.And(dbFilter...))
	}

	crsData, errData := c.Ctx.Find(NewTopMenuModel(), query)
	if crsData != nil {
		defer crsData.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	// defer crsData.Close()
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	}
	errData = crsData.Fetch(&data, 0, false)

	//	log.Printf("Data => %#v\n", len(data))
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	} else {
		retModel.Set("Records", data)
	}
	crsTotal, errTotal := c.Ctx.Find(NewTopMenuModel(), queryTotal)
	if crsTotal != nil {
		defer crsTotal.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	// defer crsTotal.Close()
	if errTotal != nil {
		return c.SetResultInfo(true, errTotal.Error(), nil)
	}
	errTotal = crsTotal.Fetch(&total, 0, false)

	//	log.Printf("Total => %#v\n", len(total))
	if errTotal != nil {
		return c.SetResultInfo(true, errTotal.Error(), nil)
	} else {
		retModel.Set("Count", len(total))
	}
	ret.Data = retModel

	return ret
}

func (c *MenuSettingController) GetAccessMenu(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	url := k.Request.URL.String()
	accesMenu := k.Session("roles").([]SysRolesModel)

	access := []tk.M{}
	for _, o := range accesMenu[0].Menu {

		if o.Url == url {
			obj := tk.M{}
			obj.Set("view", o.View)
			obj.Set("create", o.Create)
			obj.Set("approve", o.Approve)
			obj.Set("delete", o.Delete)
			obj.Set("process", o.Process)
			obj.Set("edit", o.Edit)
			access = append(access, obj)
		}

	}

	return access
}

func (c *MenuSettingController) GetSelectMenu(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	payLoad := struct {
		Id string
	}{}
	err := k.GetPayload(&payLoad)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	var dbFilter []*db.Filter
	if payLoad.Id != "" {
		dbFilter = append(dbFilter, db.Eq("_id", payLoad.Id))
	}
	//	log.Printf("QUERY=> %#v\n", query)
	queryTotal := tk.M{}
	query := tk.M{}
	data := make([]TopMenuModel, 0)
	total := make([]TopMenuModel, 0)
	retModel := tk.M{}

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
		queryTotal.Set("where", db.And(dbFilter...))
	}

	crsData, errData := c.Ctx.Find(NewTopMenuModel(), query)
	if crsData != nil {
		defer crsData.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	// defer crsData.Close()
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	}
	errData = crsData.Fetch(&data, 0, false)

	//	log.Printf("Data => %#v\n", len(data))
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	} else {
		retModel.Set("Records", data)
	}
	crsTotal, errTotal := c.Ctx.Find(NewTopMenuModel(), queryTotal)
	if crsTotal != nil {
		defer crsTotal.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	// defer crsTotal.Close()
	if errTotal != nil {
		return c.SetResultInfo(true, errTotal.Error(), nil)
	}
	errTotal = crsTotal.Fetch(&total, 0, false)

	//	log.Printf("Total => %#v\n", len(total))
	if errTotal != nil {
		return c.SetResultInfo(true, errTotal.Error(), nil)
	} else {
		retModel.Set("Count", len(total))
	}
	ret.Data = retModel

	return ret
}

type submenu struct {
	Id   string
	Name string
	Menu map[string]string
}

type addMenu struct {
	Url      string
	Access   bool
	Approve  bool
	Create   bool
	Delete   bool
	Edit     bool
	Enable   bool
	Haschild bool
	Menuid   string
	Menuname string
	Process  bool
	View     bool
	Parent   string
	Checkall bool
}

func (c *MenuSettingController) SaveMenuTop(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payLoad := struct {
		Id        string
		PageId    string
		Parent    string
		Title     string
		Url       string
		IndexMenu int
		Enable    bool
	}{}
	err := k.GetPayload(&payLoad)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	mt := NewTopMenuModel()
	mt.Id = payLoad.Id
	mt.PageId = payLoad.PageId
	mt.Parent = payLoad.Parent
	mt.Title = payLoad.Title
	mt.Url = payLoad.Url
	mt.IndexMenu = payLoad.IndexMenu
	mt.Enable = payLoad.Enable
	c.Ctx.Save(mt)

	// =============================== save menu on sysrole

	data := make([]SysRolesModel, 0)

	crsData, errData := c.Ctx.Find(NewSysRolesModel(), nil)
	if crsData != nil {
		defer crsData.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	// defer crsData.Close()
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	}

	errData = crsData.Fetch(&data, 0, false)

	menu := Detailsmenu{
		Menuid:   payLoad.Id,
		Menuname: payLoad.Title,
		Access:   false,
		View:     false,
		Create:   false,
		Approve:  false,
		Delete:   false,
		Process:  false,
		Edit:     false,
		Parent:   payLoad.Parent,
		Haschild: false,
		Enable:   false,
		Url:      payLoad.Url,
		Checkall: false,
	}

	w := NewSysRolesModel()

	for _, dt := range data {
		w.Id = dt.Id
		w.Name = dt.Name
		w.Menu = append(dt.Menu, menu)
		w.Status = dt.Status
		c.Ctx.Save(w)
	}

	return c.SetResultInfo(false, "Menu has been successfully created.", nil)
}

func (c *MenuSettingController) DeleteMenuTop(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payLoad := struct {
		Id string
	}{}

	err := k.GetPayload(&payLoad)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	u := NewTopMenuModel()
	err = c.Ctx.GetById(u, payLoad.Id)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	err = c.Ctx.Delete(u)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	//--------------- Delete SysRole -----------------
	data := make([]SysRolesModel, 0)
	crsData, errData := c.Ctx.Find(NewSysRolesModel(), nil)
	if crsData != nil {
		defer crsData.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	// defer crsData.Close()
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	}
	errData = crsData.Fetch(&data, 0, false)
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	}
	for _, dt := range data {
		NewMenu := []Detailsmenu{}
		ModelRole := NewSysRolesModel()
		ModelRole.Id = dt.Id
		ModelRole.Name = dt.Name
		for _, arrMenu := range dt.Menu {
			mVal, _ := tk.ToM(arrMenu)
			Menuid := mVal["Menuid"].(string)
			if Menuid != payLoad.Id {
				NewMenu = append(NewMenu, arrMenu)
			}
		}
		ModelRole.Menu = NewMenu
		ModelRole.Status = dt.Status
		c.Ctx.Save(ModelRole)
	}

	return c.SetResultInfo(false, "Menu has been successfully created.", nil)
}

func (c *MenuSettingController) UpdateMenuTop(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	payLoad := struct {
		Id        string
		PageId    string
		Parent    string
		Title     string
		Url       string
		IndexMenu int
		Enable    bool
	}{}
	err := k.GetPayload(&payLoad)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	mt := NewTopMenuModel()
	mt.Id = payLoad.Id
	mt.PageId = payLoad.PageId
	mt.Parent = payLoad.Parent
	mt.Title = payLoad.Title
	mt.Url = payLoad.Url
	mt.IndexMenu = payLoad.IndexMenu
	mt.Enable = payLoad.Enable
	c.Ctx.Save(mt)
	//--------------- Delete Update -----------------
	data := make([]SysRolesModel, 0)
	crsData, errData := c.Ctx.Find(NewSysRolesModel(), nil)
	if crsData != nil {
		defer crsData.Close()
	} else {
		return c.SetResultInfo(true, "error on query", nil)
	}
	// defer crsData.Close()
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	}
	errData = crsData.Fetch(&data, 0, false)
	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	}
	for _, dt := range data {
		NewMenu := []Detailsmenu{}
		ModelRole := NewSysRolesModel()
		ModelRole.Id = dt.Id
		ModelRole.Name = dt.Name
		for _, arrMenu := range dt.Menu {
			mVal, _ := tk.ToM(arrMenu)
			Menuid := mVal["Menuid"].(string)
			if Menuid != payLoad.Id {
				NewMenu = append(NewMenu, arrMenu)
			} else {
				UpdateMenu := Detailsmenu{
					Menuid:   mVal["Menuid"].(string),
					Menuname: payLoad.Title,
					Access:   mVal["Access"].(bool),
					View:     mVal["View"].(bool),
					Create:   mVal["Create"].(bool),
					Approve:  mVal["Approve"].(bool),
					Delete:   mVal["Delete"].(bool),
					Process:  mVal["Process"].(bool),
					Edit:     mVal["Edit"].(bool),
					Parent:   payLoad.Parent,
					Haschild: mVal["Haschild"].(bool),
					Enable:   payLoad.Enable,
					Url:      payLoad.Url,
					Checkall: mVal["Checkall"].(bool),
				}
				NewMenu = append(NewMenu, UpdateMenu)
			}
		}
		ModelRole.Menu = NewMenu
		ModelRole.Status = dt.Status
		c.Ctx.Save(ModelRole)
	}
	return c.SetResultInfo(false, "Menu has been successfully update.", nil)
}
