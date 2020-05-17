package controllers

import (
	. "creativelab/ecleave-dev/models"

	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
)

type DesignationController struct {
	*BaseController
}

func (c *DesignationController) GetDataDesignation(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	data := make([]DesignationModel, 0)
	query := tk.M{}
	crs, err := c.Ctx.Find(NewDesignationModel(), query)
	if err != nil {
		return nil
	}

	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return nil
	}

	return data
}

func (c *DesignationController) GetDataNDesignation(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	data := make([]DesignationModel, 0)
	query := tk.M{}
	crs, err := c.Ctx.Find(NewDesignationModel(), query)
	if err != nil {
		return nil
	}
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	err = crs.Fetch(&data, 0, false)
	if err != nil {
		return nil
	}
	datas := []DesignationList{}
	for _, each := range data {
		datas = append(datas, DesignationList{
			Code:        each.Designation,
			Designation: each.Designation,
		})
	}
	return datas
}
