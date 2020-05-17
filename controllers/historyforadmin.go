package controllers

import ( // "fmt"

	// . "creativelab/ecleave-dev/models"

	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	// "gopkg.in/gomail.v2"
	// "gopkg.in/mgo.v2/bson"
	"creativelab/ecleave-dev/repositories"

	db "github.com/creativelab/dbox"
)

type HistoryForAdminController struct {
	*BaseController
}

func (c *HistoryForAdminController) Default(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	// k.Config.NoLog = true
	k.Config.OutputType = knot.OutputTemplate
	// DataAccess := Previlege{}

	DataAccess := c.SetViewData(k, nil)

	// for _, o := range access {
	// 	DataAccess.Create = o["Create"].(bool)
	// 	DataAccess.View = o["View"].(bool)
	// 	DataAccess.Delete = o["Delete"].(bool)
	// 	DataAccess.Process = o["Process"].(bool)
	// 	DataAccess.Delete = o["Delete"].(bool)
	// 	DataAccess.Edit = o["Edit"].(bool)
	// 	DataAccess.Menuid = o["Menuid"].(string)
	// 	DataAccess.Menuname = o["Menuname"].(string)
	// 	DataAccess.Approve = o["Approve"].(bool)
	// 	DataAccess.Username = o["Username"].(string)
	// }

	return DataAccess
}

func (c *HistoryForAdminController) GetDataRemote(k *knot.WebContext) (interface{}, error) {
	k.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}

	match := tk.M{}.Set("$match", tk.M{}.Set("projects.ismanagersend", false).Set("projects.isprojectsend", true))
	pipe = append(pipe, match)
	pipe = append(pipe, tk.M{"$group": tk.M{"_id": tk.M{"idop": "$idop", "name": "$name", "reason": "$reason", "project": "$projects"}, "date": tk.M{"$push": "$dateleave"}}})

	data, err := new(repositories.RemoteDboxRepo).GetByPipe(pipe)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *HistoryForAdminController) GetDataRemote2(k *knot.WebContext) (interface{}, error) {
	k.Config.OutputType = knot.OutputJson
	pipe := []tk.M{}

	// match := tk.M{}.Set("$match", tk.M{}.Set("projects.ismanagersend", false).Set("projects.isprojectsend", true))
	// pipe = append(pipe, match)
	pipe = append(pipe, tk.M{"$match": tk.M{"projects.ismanagersend": false}})
	pipe = append(pipe, tk.M{"$match": tk.M{"projects.isapprovalleader": true}})
	pipe = append(pipe, tk.M{"$group": tk.M{"_id": tk.M{"idop": "$idop", "name": "$name", "reason": "$reason", "project": "$projects"}, "date": tk.M{"$push": "$dateleave"}}})

	// data, err := new(repositories.RemoteDboxRepo).GetByPipe(pipe)
	// if err != nil {
	// 	return nil, err
	// }

	datas := []tk.M{}
	crs, err := c.Ctx.Connection.NewQuery().From("remote").Command("pipe", pipe).Cursor(nil)
	if crs != nil {
		defer crs.Close()
	}

	if err != nil {
		return datas, err
	}

	err = crs.Fetch(&datas, 0, false)
	if err != nil {
		return datas, err
	}

	// tk.Println("----------- datas ", datas)

	// return datas, err

	return datas, nil
}

func (c *HistoryForAdminController) AdminGetHistory(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	// filter := tk.M{}.Set("where", dbox.And(dbox.Eq("statusmanagerproject.statusrequest", "Pending"), dbox.Eq("statusrequest", "Pending")))

	var dbFilter []*db.Filter

	dbFilter = append(dbFilter, db.Eq("statusprojectleader.statusrequest", "Approved"))
	dbFilter = append(dbFilter, db.Eq("statusmanagerproject.statusrequest", "Pending"))
	dbFilter = append(dbFilter, db.Eq("resultrequest", "Pending"))
	// dbFilter = append(dbFilter, db.Eq("isdelete", false))
	// dbFilter = append(dbFilter, db.Eq("statusrequest", "Pending"))
	query := tk.M{}
	if len(dbFilter) > 0 {
		query.Set("where", db.And(dbFilter...))
	}

	dataLeave, err := new(repositories.LeaveOrmRepo).GetByParamMasterLeave(query)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// dataRemote := []RemoteModel{}
	data := tk.M{}
	data.Set("Leave", dataLeave)

	dataRemote, err := c.GetDataRemote2(k)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	data.Set("Remote", dataRemote)

	return c.SetResultInfo(false, "success", data)
}
