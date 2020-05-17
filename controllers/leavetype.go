package controllers

import (
	. "creativelab/ecleave-dev/models"

	"github.com/creativelab/knot/knot.v1"
)

type LeaveTypeController struct {
	*BaseController
}

func (c *LeaveTypeController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	dataLeave := make([]LeaveTypeModel, 0)
	crsLeave, errLeave := c.Ctx.Find(NewLeaveTypeModel(), nil)

	if crsLeave != nil {
		defer crsLeave.Close()
	} else if crsLeave == nil {
		return c.SetResultInfo(true, "Error when build query", nil)
	}
	defer crsLeave.Close()
	if errLeave != nil {
		return c.SetResultInfo(true, errLeave.Error(), nil)
	}

	errLeave = crsLeave.Fetch(&dataLeave, 0, false)
	if errLeave != nil {
		return c.SetResultInfo(true, errLeave.Error(), nil)
	}

	return dataLeave
}
