package controllers

import (
	. "creativelab/ecleave-dev/models"

	"creativelab/ecleave-dev/services"

	"github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	"gopkg.in/mgo.v2/bson"
)

type HRDAdminController struct {
	*BaseController
}

func (c *HRDAdminController) GetDataHR(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	dataHR := make([]*HRDAdminModel, 0)
	query := tk.M{}
	crs, err := c.Ctx.Find(NewHRDAdminModel(), query)
	if err != nil {
		return nil
	}

	if crs != nil {
		defer crs.Close()
	} else {
		return nil
	}
	// defer crs.Close()
	err = crs.Fetch(&dataHR, 0, false)
	if err != nil {
		return nil
	}

	return dataHR
}

func (c *HRDAdminController) SaveDataHR(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := new(HRDAdminModel)
	err := k.GetPayload(&payload)
	if err != nil {
		return err
	}

	if payload.Id == "" {
		payload.Id = bson.NewObjectId().Hex()
	}

	err = c.Ctx.Save(payload)
	if err != nil {
		return err
	}
	return c.SetResultInfo(false, "data has been saved", nil)
}

func (c *HRDAdminController) Delete(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	payload := struct {
		Id string
	}{}

	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	err = new(services.HrdServices).DeleteByID(payload.Id)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	return c.SetResultOK(nil)
}
