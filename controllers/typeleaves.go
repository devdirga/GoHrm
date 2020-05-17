package controllers

import (
	. "creativelab/ecleave-dev/models"

	knot "github.com/creativelab/knot/knot.v1"
)

type TypeLeavesController struct {
	*BaseController
}

func (c *TypeLeavesController) GetTypeLeaves(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	data := make([]*TypeLeavesModel, 0)

	crs, err := c.Ctx.Find(NewTypeLeavesModel(), nil)
	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	err = crs.Fetch(&data, 0, false)

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if len(data) == 0 {
		return c.SetResultInfo(true, "no data", nil)
	}

	return c.SetResultInfo(false, "success", data)
}
