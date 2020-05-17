package controllers

import (
	. "creativelab/ecleave-dev/models"

	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	// tk "github.com/creativelab/toolkit"
)

type ProjectRuleController struct {
	*BaseController
}

func (c *ProjectRuleController) Default(k *knot.WebContext) interface{} {
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

func (c *ProjectRuleController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	dataProjRule := make([]ProjectRuleModel, 0)
	crsProjRule, errProjRule := c.Ctx.Find(NewProjectRuleModel(), nil)

	if crsProjRule != nil {
		defer crsProjRule.Close()
	}
	defer crsProjRule.Close()
	if errProjRule != nil {
		return c.SetResultInfo(true, "Error when build query", nil)
	}

	errProjRule = crsProjRule.Fetch(&dataProjRule, 0, false)
	if errProjRule != nil {
		return c.SetResultInfo(true, errProjRule.Error(), nil)
	}

	return dataProjRule
}

func (c *ProjectRuleController) GetProjectRule(k *knot.WebContext) ([]ProjectRuleModel, error) {
	k.Config.OutputType = knot.OutputJson

	dataProjRule := make([]ProjectRuleModel, 0)
	crsProjRule, errProjRule := c.Ctx.Find(NewProjectRuleModel(), nil)

	if crsProjRule != nil {
		defer crsProjRule.Close()
	}
	defer crsProjRule.Close()
	if errProjRule != nil {
		return dataProjRule, nil
	}

	errProjRule = crsProjRule.Fetch(&dataProjRule, 0, false)
	if errProjRule != nil {
		return dataProjRule, errProjRule
	}

	return dataProjRule, nil
}

func (c *ProjectRuleController) GetAllUser(k *knot.WebContext) ([]SysUserModel, error) {
	k.Config.OutputType = knot.OutputJson

	users := make([]SysUserModel, 0)
	crsProjRule, errProjRule := c.Ctx.Find(NewSysUserModel(), nil)

	if crsProjRule != nil {
		defer crsProjRule.Close()
	}
	defer crsProjRule.Close()
	if errProjRule != nil {
		return users, nil
	}

	errProjRule = crsProjRule.Fetch(&users, 0, false)
	if errProjRule != nil {
		return users, errProjRule
	}

	return users, nil
}

func (c *ProjectRuleController) EmployeeProjectRule(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	users, err := c.GetAllUser(k)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	rule, err := c.GetProjectRule(k)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	PC := []SysUserModel{}
	PM := []SysUserModel{}
	Lead := []SysUserModel{}

	res := tk.M{}

	for _, rl := range rule {
		// data[i].GetString
		for _, dt := range users {

			// fmt.Println("------- rl ", rl.Id.Hex())
			// fmt.Println("------- dt ", dt.ProjectRuleID)
			// fmt.Println("------- nm ", dt.ProjectRuleName)

			if rl.Id.Hex() == dt.ProjectRuleID {
				// fmt.Println("------- kk ", dt.ProjectRuleName)
				dt.ProjectRuleName = rl.Name
				if rl.Name == "Project Manager" {
					PM = append(PM, dt)
				} else if rl.Name == "Project Coordinator" {
					PC = append(PC, dt)
				} else if rl.Name == "Project Leader" {
					Lead = append(Lead, dt)
				}
			}
		}

	}

	res.Set("PM", PM).Set("PC", PC).Set("lead", Lead)

	return c.SetResultInfo(false, "success", res)
}

func (c *ProjectRuleController) SaveProjectRule(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	k.Config.OutputType = knot.OutputJson

	payload := []SysUserModel{}

	err := k.GetPayload(&payload)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	// rule, err := c.GetProjectRule(k)

	// if err != nil {
	// 	return c.SetResultInfo(true, err.Error(), nil)
	// }
	loc := LocationController(*c)
	countpc := 0
	pcm := []PClist{}
	// newpc = new(LocationModel)
	newpc := PClist{}
	for _, res := range payload {
		err = c.Ctx.Save(&res)
		if err != nil {
			return c.SetResultInfo(true, err.Error(), nil)
		}

		dc := loc.GetLocationParam(k, res.Location)
		if res.ProjectRuleName == "Project Coordinator" {
			if len(dc.PC) > 0 {
				for _, lc1 := range dc.PC {
					if lc1.UserId != res.Id {
						countpc = countpc + 1
						if len(dc.PC) == countpc {
							newpc.IdEmp = res.EmpId
							newpc.Name = res.Fullname
							newpc.Location = res.Location
							newpc.Email = res.Email
							newpc.PhoneNumber = res.PhoneNumber
							newpc.UserId = res.Id
							dc.PC = append(dc.PC, newpc)
							// id := hex.EncodeToString([]byte(dc.Id))
							// dc.Id = bson.ObjectId(id)

						}
					}
				}
			} else {
				newpc.IdEmp = res.EmpId
				newpc.Name = res.Fullname
				newpc.Location = res.Location
				newpc.Email = res.Email
				newpc.PhoneNumber = res.PhoneNumber
				newpc.UserId = res.Id
				dc.PC = append(dc.PC, newpc)
			}

			tk.Println("---------- dc ", dc.Id)
			err = c.Ctx.Save(dc)
			if err != nil {
				tk.Println("------------- error")
				return c.SetResultInfo(true, err.Error(), nil)
			}

		} else {
			for _, lc2 := range dc.PC {
				if lc2.UserId == res.Id {
					countpc = countpc + 1
				} else {
					pcm = append(pcm, lc2)
				}

			}
			dc.PC = []PClist{}
			dc.PC = pcm
			tk.Println("------------ masuk sama3 ", dc)
			err = c.Ctx.Save(dc)
			if err != nil {
				return c.SetResultInfo(true, err.Error(), nil)
			}
		}

	}

	return c.SetResultInfo(false, "Success", nil)
}
