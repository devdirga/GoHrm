package controllers

import (
	. "creativelab/ecleave-dev/models"
	"strconv"
	"strings"

	db "github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	"gopkg.in/mgo.v2/bson"
)

type SysRolesController struct {
	*BaseController
}

func (c *SysRolesController) Default(k *knot.WebContext) interface{} {
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

func (d *SysRolesController) GetData(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}

	oo := struct {
		Name   []interface{}
		Status bool
		Take   int
		Skip   int
		Sort   []tk.M
	}{}

	err := r.GetPayload(&oo)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}
	var dbFilter []*db.Filter

	dbFilter = append(dbFilter, db.Eq("status", oo.Status))

	if len(oo.Name) != 0 {
		dbFilter = append(dbFilter, db.In("name", oo.Name...))
	}

	sort := ""
	dir := ""
	if len(oo.Sort) > 0 {
		sort = strings.ToLower(oo.Sort[0].Get("field").(string))
		dir = oo.Sort[0].Get("dir").(string)
	}

	if sort == "" {
		sort = "acc_no_map"
	}
	if dir == "" {
		dir = "-"
	}
	if dir != "" && dir != "asc" {
		sort = "-" + sort
	}

	queryTotal := tk.M{}
	query := tk.M{}
	data := make([]SysRolesModel, 0)
	total := make([]SysRolesModel, 0)
	retModel := tk.M{}
	query.Set("limit", oo.Take)
	query.Set("skip", oo.Skip)
	query.Set("order", []string{sort})

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
		queryTotal.Set("where", db.And(dbFilter...))
	}

	crsData, errData := d.Ctx.Find(NewSysRolesModel(), query)
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
		retModel.Set("Records", data)
	}
	crsTotal, errTotal := d.Ctx.Find(NewSysRolesModel(), queryTotal)
	if crsTotal != nil {
		defer crsTotal.Close()
	} else {
		return d.SetResultInfo(true, "error on query", nil)
	}
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

func (c *SysRolesController) GetMenu(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	filterForm := struct {
		Id   string
		Take int
		Skip int
		Sort []tk.M
	}{}
	err := k.GetPayload(&filterForm)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	sort := ""
	dir := ""
	if len(filterForm.Sort) > 0 {
		sort = strings.ToLower(filterForm.Sort[0].Get("field").(string))
		dir = filterForm.Sort[0].Get("dir").(string)
	}

	if sort == "" {
		sort = "name"
	}

	if dir != "" && dir != "asc" {
		sort = "-" + sort
	}

	var dbFilter []*db.Filter
	if filterForm.Id != "" {
		id, _ := strconv.Atoi(filterForm.Id)
		dbFilter = append(dbFilter, db.Eq("_id", id))
	}

	queryTotal := tk.M{}
	query := tk.M{}
	query.Set("limit", filterForm.Take)
	query.Set("skip", filterForm.Skip)
	query.Set("order", []string{sort})
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
		queryTotal.Set("where", db.And(dbFilter...))
	}

	data := make([]TopMenuModel, 0)
	total := make([]TopMenuModel, 0)
	retModel := tk.M{}

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

	if errData != nil {
		return c.SetResultInfo(true, errData.Error(), nil)
	} else {
		retModel.Set("Records", data)
	}
	crsTotal, errTotal := c.Ctx.Find(NewTopMenuModel(), queryTotal)
	// defer crsTotal.Close()
	if errTotal != nil {
		return c.SetResultInfo(true, errTotal.Error(), nil)
	}
	errTotal = crsTotal.Fetch(&total, 0, false)

	if errTotal != nil {
		return c.SetResultInfo(true, errTotal.Error(), nil)
	} else {
		retModel.Set("Count", len(total))
	}
	ret.Data = retModel

	return ret
}

func (c *SysRolesController) GetMenuEdit(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson
	ret := ResultInfo{}
	filterForm := struct {
		Id   string
		Name string
	}{}
	err := k.GetPayload(&filterForm)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	var dbFilter []*db.Filter

	if filterForm.Id != "" {
		dbFilter = append(dbFilter, db.Eq("_id", bson.ObjectIdHex(filterForm.Id)))
	}

	if filterForm.Name != "" {
		dbFilter = append(dbFilter, db.Eq("name", filterForm.Name))
	}

	query := tk.M{}
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	data := make([]SysRolesModel, 0)
	retModel := tk.M{}

	crsData, errData := c.Ctx.Find(NewSysRolesModel(), query)
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
	} else {
		retModel.Set("Records", data)
	}

	ret.Data = retModel

	return ret
}

func (d *SysRolesController) SaveData(r *knot.WebContext) interface{} {
	r.Config.OutputType = knot.OutputJson
	oo := struct {
		Id     string
		Name   string
		Status bool
		Menu   []tk.M
	}{}
	ret := ResultInfo{}
	o := NewSysRolesModel()

	err := r.GetPayload(&oo)
	if err != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	}

	if oo.Id != "" {
		o.Id = bson.ObjectIdHex(oo.Id)
	} else {
		o.Id = bson.NewObjectId()
	}

	o.Name = oo.Name
	o.Status = oo.Status

	xs := []string{}

	tempMenu := o.Menu
	for _, det := range oo.Menu {
		dtl := Detailsmenu{}

		access := det["access"].(bool)
		parent := det["parent"].(string)

		if access == true && parent != "" {
			xs = append(xs, parent)
		}
		dtl.Url = det["Url"].(string)
		dtl.Access = det["access"].(bool)
		dtl.Approve = det["approve"].(bool)
		dtl.Create = det["create"].(bool)
		dtl.Delete = det["delete"].(bool)
		dtl.Edit = det["edit"].(bool)
		dtl.Enable = det["enable"].(bool)
		dtl.Haschild = det["haschild"].(bool)
		dtl.Menuid = det["menuid"].(string)
		dtl.Menuname = det["menuname"].(string)
		dtl.Process = det["process"].(bool)
		dtl.View = det["view"].(bool)
		dtl.Parent = det["parent"].(string)
		dtl.Checkall = det["checkall"].(bool)

		tempMenu = append(tempMenu, dtl)
	}

	d.DistinctArray(&xs)

	for _, dt := range tempMenu {
		menuid := dt.Menuid

		for _, op := range xs {
			if menuid == op {
				dt.Access = true
			}
		}
		o.Menu = append(o.Menu, dt)
	}

	e := d.Ctx.Save(o)
	if e != nil {
		return d.SetResultInfo(true, err.Error(), nil)
	} else {
		ret.IsError = false
		ret.Message = "Saving Data Successfully"
		ret.Data = ""
	}

	return ret
}

func (d *SysRolesController) DistinctArray(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}
