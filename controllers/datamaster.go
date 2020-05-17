package controllers

import (
	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type DatamasterController struct {
	*BaseController
}

func (c *DatamasterController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	return ""
}

func (c *DatamasterController) Accounts(k *knot.WebContext) interface{} {
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
	// //c.InsertActivityLog("Master-Account", "VIEW", k)
	return DataAccess
}

func (d *DatamasterController) GetUsername(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	query := d.Ctx.Connection.NewQuery()
	result := []tk.M{}
	csr, e := query.
		Select("username").From("SysUsers").Order("username").Cursor(nil)
	e = csr.Fetch(&result, 0, false)

	if e != nil {
		return result
	}
	defer csr.Close()

	return result
}

func (d *DatamasterController) GetRoles(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	query := d.Ctx.Connection.NewQuery()
	result := []tk.M{}
	csr, e := query.
		Select("name").From("SysRoles").Order("name").Cursor(nil)
	e = csr.Fetch(&result, 0, false)

	if e != nil {
		return result
	}
	defer csr.Close()

	return result
}

func (d *DatamasterController) GetFullname(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson

	query := d.Ctx.Connection.NewQuery()
	result := []tk.M{}
	csr, e := query.
		Select("fullname").From("SysUsers").Order("fullname").Cursor(nil)
	e = csr.Fetch(&result, 0, false)

	if e != nil {
		return result
	}
	defer csr.Close()

	return result
}
