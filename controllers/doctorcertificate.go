package controllers

import (
	. "creativelab/ecleave-dev/models"
	// "time"
	db "github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type DoctorCertificateController struct {
	*BaseController
}

func (c *DoctorCertificateController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	viewData := tk.M{}
	if k.Session("jobrolename") != nil {
		viewData.Set("JobRoleName", k.Session("jobrolename").(string))
		viewData.Set("JobRoleLevel", k.Session("jobrolelevel").(int))
	} else {
		viewData.Set("JobRoleName", "")
		viewData.Set("JobRoleLevel", "")
	}

	DataAccess := c.SetViewData(k, viewData)
	// k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	k.Config.IncludeFiles = []string{
		"_loader.html",
	}

	return DataAccess

}

func (c *DoctorCertificateController) GetDataCertificate(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.LoadBase(k)

	var dbFilter []*db.Filter
	query := tk.M{}
	data := make([]*RequestLeaveModel, 0)

	dbFilter = append(dbFilter, db.Eq("isemergency", true))
	dbFilter = append(dbFilter, db.Eq("isattach", true))
	dbFilter = append(dbFilter, db.Eq("isreset", false))
	dbFilter = append(dbFilter, db.Eq("resultrequest", "Approved"))

	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	crs, err := c.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	err = crs.Fetch(&data, 0, false)

	if len(data) > 0 {
		return c.SetResultInfo(false, "success", data)
	}
	return c.SetResultInfo(true, "data empty", data)
}

// Deletefile is ...
func (c *DoctorCertificateController) Deletefile(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	c.LoadBase(k)
	p := struct {
		Id string
	}{}
	err := k.GetPayload(&p)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	query := tk.M{}
	var dbFilter []*db.Filter
	dbFilter = append(dbFilter, db.Eq("_id", (p.Id)))
	leaves := []RequestLeaveModel{}
	query.Set("where", db.And(dbFilter...))
	crs, _ := c.Ctx.Find(NewRequestLeave(), query)
	if crs != nil {
		defer crs.Close()
	}
	err = crs.Fetch(&leaves, 0, false)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	if len(leaves) > 0 {
		leaves[0].FileLocation = "-"
		leaves[0].IsAttach = false
		err = c.Ctx.Save(&leaves[0])
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}
		return c.SetResultInfo(false, "success delete attachment", nil)
	} else {
		return c.SetResultInfo(true, "failed delete attachment", nil)
	}
}
