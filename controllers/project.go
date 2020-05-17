package controllers

import (
	. "creativelab/ecleave-dev/models"
	"fmt"
	"time"

	// "creativelab/ecleave-dev/services"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	db "github.com/creativelab/dbox"

	knot "github.com/creativelab/knot/knot.v1"
	tk "github.com/creativelab/toolkit"
	"gopkg.in/mgo.v2/bson"

	// tk "github.com/creativelab/toolkit"
	"github.com/tealeg/xlsx"
	xl "github.com/tealeg/xlsx"
)

type ProjectController struct {
	*BaseController
}

func (c *ProjectController) Default(k *knot.WebContext) interface{} {
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
func (c *ProjectController) DefaultV2(k *knot.WebContext) interface{} {
	c.LoadBase(k)
	access := c.LoadBase(k)
	// k.Config.NoLog = true
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
	if k.Session("jobrolename") != nil {
		DataAccess.JobRoleName = k.Session("jobrolename").(string)
		DataAccess.JobRoleLevel = k.Session("jobrolelevel").(int)
	}

	return DataAccess
}

func (c *ProjectController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	dataProject := []ProjectModel{}
	query := tk.M{}
	crs, err := c.Ctx.Find(NewListProject(), query)
	if crs != nil {
		defer crs.Close()
	} else if crs == nil {
		return c.SetResultInfo(true, "Error when build query", nil)
	}

	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	err = crs.Fetch(&dataProject, 0, false)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	return dataProject
}

func (c *ProjectController) SaveProject(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	payload := new(ProjectModel)
	err := k.GetPayload(&payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}

	if payload.Id == "" {
		payload.Id = bson.NewObjectId().Hex()
	}

	err = c.Ctx.Save(payload)
	if err != nil {
		return c.SetResultInfo(true, err.Error(), nil)
	}
	return c.SetResultInfo(false, "Data Project has been save", nil)
}

func (c *ProjectController) UploadFile(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson

	//get data Location -- start
	dataLocation := make([]LocationModel, 0)
	crsLocation, errLocation := c.Ctx.Find(NewLocationModel(), nil)

	if crsLocation != nil {
		defer crsLocation.Close()
	}
	defer crsLocation.Close()
	if errLocation != nil {
		tk.Println(errLocation)
	}

	errLocation = crsLocation.Fetch(&dataLocation, 0, false)
	if errLocation != nil {
		tk.Println(errLocation)
	}
	//get data Location -- end

	reader, err := k.Request.MultipartReader()
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	var fileLocation, fileName string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		fileName = tk.RandomString(32) + filepath.Ext(part.FileName())
		fileLocation = filepath.Join("assets/doc/", fileName)
		dst, err := os.Create(fileLocation)
		if dst != nil {
			defer dst.Close()
		}
		if err != nil {
			return c.SetResultError(err.Error(), nil)
		}

		if _, err := io.Copy(dst, part); err != nil {
			return c.SetResultError(err.Error(), nil)
		}
	}

	//get data sysuser -- start
	data := make([]SysUserModel, 0)
	crs2, errdata := c.Ctx.Find(NewSysUserModel(), nil)

	if crs2 != nil {
		defer crs2.Close()
	}
	defer crs2.Close()
	if errdata != nil {
		tk.Println(errdata)
	}

	errdata = crs2.Fetch(&data, 0, false)
	if errdata != nil {
		tk.Println(errdata)
	}
	//get data sysuser -- end

	xlFile, err := xl.OpenFile(fileLocation)
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	if len(xlFile.Sheets) == 0 {
		return c.SetResultError("Excel contains no sheet", nil)
	}

	isNoRows := true
	var listAllUnidentifedName = make([]ListUnidentifiedName, 0)
	for iRow, row := range xlFile.Sheets[0].Rows {
		var stsUnidentified = false
		var iAllUnidentifedName ListUnidentifiedName
		if iRow == 0 {
			continue
		}

		if len(row.Cells) == 0 {
			continue
		}

		isNoRows = false

		var ID, ProjectKey, ProjectName, ProjectManager, BusinessAnalyst, ProjectLeader, Developer, LocationName, Address, Uri, Active string

		if len(row.Cells) > 0 {
			ID = row.Cells[0].String()

		}
		if len(row.Cells) > 1 {
			ProjectKey = row.Cells[1].String()
		}
		if len(row.Cells) > 2 {
			ProjectName = row.Cells[2].String()
		}
		if len(row.Cells) > 3 {
			ProjectManager = row.Cells[3].String()
		}
		if len(row.Cells) > 4 {
			BusinessAnalyst = row.Cells[4].String()
		}
		if len(row.Cells) > 5 {
			ProjectLeader = row.Cells[5].String()
		}
		if len(row.Cells) > 6 {
			Developer = row.Cells[6].String()
		}
		if len(row.Cells) > 7 {
			LocationName = row.Cells[7].String()
		}
		if len(row.Cells) > 8 {
			Address = row.Cells[8].String()
		}
		if len(row.Cells) > 9 {
			Uri = row.Cells[9].String()
		}
		if len(row.Cells) > 10 {
			Active = row.Cells[10].String()
		}

		if strings.TrimSpace(ID) == "" && strings.TrimSpace(ProjectName) != "" {
			ID = bson.NewObjectId().Hex()
		}

		if strings.TrimSpace(ID) == "" && strings.TrimSpace(ProjectName) == "" {
			continue
		}
		iAllUnidentifedName.ProjectName = ProjectName
		var findPM OcupationData
		var valNamePM = strings.ToLower(strings.TrimSpace(ProjectManager))
		var unidentifiedPM = "\nUnidentified Project Manager Name : "
		var stsNamePM = false
		if valNamePM != "" {
			for _, valData := range data {
				if valData.EmpId == valNamePM || strings.ToLower(valData.Fullname) == valNamePM {

					findPM.UserId = valData.Id
					findPM.IdEmp = valData.EmpId
					findPM.Name = valData.Fullname
					findPM.Location = valData.Location
					findPM.Email = valData.Email
					findPM.PhoneNumber = valData.PhoneNumber
					stsNamePM = true
					continue
				}
			}
		}
		if !stsNamePM {
			unidentifiedPM = unidentifiedPM + strings.TrimSpace(ProjectManager) + "; "
			iAllUnidentifedName.ProjectManager = iAllUnidentifedName.ProjectManager + strings.TrimSpace(ProjectManager) + "; "
			stsUnidentified = true
		}

		splitBA := strings.Split(BusinessAnalyst, ";")
		var listFindBA = make([]OcupationData, 0)
		var unidentifiedBA = "\nUnidentified Business Analyst Name : "

		if len(splitBA) > 0 {
			for _, val := range splitBA {
				var stsNameBA = false
				var valName = strings.ToLower(strings.TrimSpace(val))
				if valName != "" {
					for _, valData := range data {
						if valData.EmpId == valName || strings.ToLower(valData.Fullname) == valName {
							var valOcupationData OcupationData
							valOcupationData.UserId = valData.Id
							valOcupationData.IdEmp = valData.EmpId
							valOcupationData.Name = valData.Fullname
							valOcupationData.Location = valData.Location
							valOcupationData.Email = valData.Email
							valOcupationData.PhoneNumber = valData.PhoneNumber
							listFindBA = append(listFindBA, valOcupationData)
							stsNameBA = true
							continue

						}
					}
				}
				if !stsNameBA && valName != "" {
					unidentifiedBA = unidentifiedBA + strings.TrimSpace(val) + "; "
					iAllUnidentifedName.BusinessAnalist = iAllUnidentifedName.BusinessAnalist + strings.TrimSpace(val) + "; "
					stsUnidentified = true
				}
			}
		}

		var findTL OcupationData
		var valNameTL = strings.ToLower(strings.TrimSpace(ProjectLeader))
		var unidentifiedTL = "\nUnidentified Team Leader Name : "
		var stsNameTL = false
		if valNameTL != "" {
			for _, valData := range data {
				if valData.EmpId == valNameTL || strings.ToLower(valData.Fullname) == valNameTL {

					findTL.UserId = valData.Id
					findTL.IdEmp = valData.EmpId
					findTL.Name = valData.Fullname
					findTL.Location = valData.Location
					findTL.Email = valData.Email
					findTL.PhoneNumber = valData.PhoneNumber
					stsNameTL = true
					continue
				}
			}
		}
		if !stsNameTL {
			unidentifiedTL = unidentifiedTL + strings.TrimSpace(ProjectLeader) + "; "
			iAllUnidentifedName.ProjectLeader = iAllUnidentifedName.ProjectLeader + strings.TrimSpace(ProjectLeader) + "; "
			stsUnidentified = true
		}
		splitDev := strings.Split(Developer, ";")
		var listFindDev = make([]OcupationData, 0)
		var unidentifiedDev = "\nUnidentified Developer Name : "

		if len(splitDev) > 0 {
			for _, val := range splitDev {
				var stsNameDev = false
				var valName = strings.ToLower(strings.TrimSpace(val))
				if valName != "" {
					for _, valData := range data {
						if valData.EmpId == valName || strings.ToLower(valData.Fullname) == valName {
							var valOcupationData OcupationData
							valOcupationData.UserId = valData.Id
							valOcupationData.IdEmp = valData.EmpId
							valOcupationData.Name = valData.Fullname
							valOcupationData.Location = valData.Location
							valOcupationData.Email = valData.Email
							valOcupationData.PhoneNumber = valData.PhoneNumber
							listFindDev = append(listFindDev, valOcupationData)
							stsNameDev = true
							continue

						}
					}
				}

				if !stsNameDev && valName != "" {
					unidentifiedDev = unidentifiedDev + strings.TrimSpace(val) + "; "
					iAllUnidentifedName.Developer = iAllUnidentifedName.Developer + strings.TrimSpace(val) + "; "
					stsUnidentified = true
				}

			}
		}
		var findLocation Location
		var valNameLocation = strings.ToLower(strings.TrimSpace(LocationName))
		if valNameLocation != "" {
			for _, valData := range dataLocation {
				if valData.Id.Hex() == valNameLocation || strings.ToLower(valData.Location) == valNameLocation {

					findLocation.Id = valData.Id.Hex()
					findLocation.Name = valData.Location

					continue

				}
			}
		}
		var stsActive bool
		if s := strings.ToLower(Active); s == "yes" || s == "true" || s == "1" {
			stsActive = true
		}
		tk.Println(unidentifiedPM, unidentifiedBA, unidentifiedTL, unidentifiedDev)
		if stsUnidentified {
			listAllUnidentifedName = append(listAllUnidentifedName, iAllUnidentifedName)
		}
		tk.Println(findLocation)

		query := tk.M{}
		var dbFilter []*db.Filter
		dbFilter = append(dbFilter, db.Eq("ProjectKey", strings.TrimSpace(ProjectKey)))
		projects := []ProjectModel{}
		query.Set("where", db.And(dbFilter...))
		crs, _ := c.Ctx.Find(NewListProject(), query)
		if crs != nil {
			defer crs.Close()
		}
		err := crs.Fetch(&projects, 0, false)

		if len(projects) > 0 {
			ID = projects[0].Id
		}

		payload := new(ProjectModel)
		payload.Id = ID
		payload.ProjectKey = ProjectKey
		payload.ProjectName = ProjectName
		payload.ProjectManager = findPM
		payload.BusinessAnalist = listFindBA
		payload.ProjectLeader = findTL
		payload.Developer = listFindDev
		payload.Location = findLocation.Name
		payload.Address = Address
		payload.Uri = Uri
		payload.Active = stsActive
		err = c.Ctx.Save(payload)
		if err != nil {

		}

	}

	if isNoRows {
		return c.SetResultError("Excel contains no data", nil)
	}

	return c.SetResultOK(listAllUnidentifedName)
}

var lock sync.Mutex

func (c *ProjectController) DownloadProjectFile(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	//get data Location -- start
	dataLocation := make([]LocationModel, 0)
	crsLocation, errLocation := c.Ctx.Find(NewLocationModel(), nil)

	if crsLocation != nil {
		defer crsLocation.Close()
	}
	defer crsLocation.Close()
	if errLocation != nil {
		tk.Println(errLocation)
	}

	errLocation = crsLocation.Fetch(&dataLocation, 0, false)
	if errLocation != nil {
		tk.Println(errLocation)
	}
	//get data Location -- end

	dataProject := []ProjectModel{}
	query := tk.M{}

	crs, err := c.Ctx.Find(NewListProject(), query)
	defer crs.Close()
	if err != nil {
		return c.SetResultError(err.Error(), nil)
	}

	if crs != nil {
		defer crs.Close()
	} else if crs == nil {
	}
	if err != nil {
	}
	err = crs.Fetch(&dataProject, 0, false)
	if err != nil {
	}

	fileName := "List-ProjectProfile_" + time.Now().Format("2006-01-02_15-04-05") + ".xlsx"
	fileLocation := "assets/doc/" + fileName
	// os.Remove(fileLocation)

	file := xl.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	style := *xlsx.NewStyle()
	styleHeader := *xlsx.NewStyle()
	if err != nil {
		tk.Println(err.Error())
		return nil
	}

	style.Border.Bottom = "thin"
	style.Border.Top = "thin"
	style.Border.Right = "thin"
	style.Border.Left = "thin"

	styleHeader.Border.Bottom = "thin"
	styleHeader.Border.Top = "thin"
	styleHeader.Border.Right = "thin"
	styleHeader.Border.Left = "thin"
	styleHeader.Font.Bold = true

	row := sheet.AddRow()

	cell1 := row.AddCell()
	cell1.Value = "Project ID"
	cell1.SetStyle(&styleHeader)
	cell2 := row.AddCell()
	cell2.Value = "Project Key"
	cell2.SetStyle(&styleHeader)
	cell3 := row.AddCell()
	cell3.Value = "Project Name"
	cell3.SetStyle(&styleHeader)
	cell4 := row.AddCell()
	cell4.Value = "Project Manager"
	cell4.SetStyle(&styleHeader)
	cell5 := row.AddCell()
	cell5.Value = "Business Analyst"
	cell5.SetStyle(&styleHeader)
	cell6 := row.AddCell()
	cell6.Value = "Project Leader"
	cell6.SetStyle(&styleHeader)
	cell7 := row.AddCell()
	cell7.Value = "Developer"
	cell7.SetStyle(&styleHeader)
	cell8 := row.AddCell()
	cell8.Value = "Location"
	cell8.SetStyle(&styleHeader)
	cell9 := row.AddCell()
	cell9.Value = "Address"
	cell9.SetStyle(&styleHeader)
	cell10 := row.AddCell()
	cell10.Value = "Uri"
	cell10.SetStyle(&styleHeader)
	cell11 := row.AddCell()
	cell11.Value = "Active"
	cell11.SetStyle(&styleHeader)
	// tk.Println("masuk bos 2")

	autoNumber := 1
	for _, each := range dataProject {
		row := sheet.AddRow()

		id := fmt.Sprintf("%06d", autoNumber)
		cell1 := row.AddCell()
		cell1.Value = id
		cell1.SetStyle(&style)

		cell2 := row.AddCell()
		cell2.Value = each.ProjectKey
		cell2.SetStyle(&style)

		cell3 := row.AddCell()
		cell3.Value = strings.ToUpper(each.ProjectName)
		cell3.SetStyle(&style)

		cell4 := row.AddCell()
		cell4.Value = each.ProjectManager.Name
		cell4.SetStyle(&style)

		var joinBA string
		if len(each.BusinessAnalist) > 0 {
			for _, val := range each.BusinessAnalist {
				joinBA = joinBA + val.Name + "; "
			}
		}
		cell5 := row.AddCell()
		cell5.Value = joinBA
		cell5.SetStyle(&style)

		cell6 := row.AddCell()
		cell6.Value = each.ProjectLeader.Name
		cell6.SetStyle(&style)

		var joinDev string
		if len(each.Developer) > 0 {
			for _, val := range each.Developer {
				joinDev = joinDev + val.Name + "; "
			}
		}
		cell7 := row.AddCell()
		cell7.Value = joinDev
		cell7.SetStyle(&style)

		cell8 := row.AddCell()
		cell8.Value = each.Location
		cell8.SetStyle(&style)

		cell9 := row.AddCell()
		cell9.Value = each.Address
		cell9.SetStyle(&style)

		cell10 := row.AddCell()
		cell10.Value = each.Uri
		cell10.SetStyle(&style)

		cell11 := row.AddCell()
		if each.Active == true {
			cell11.Value = "yes"
		} else {
			cell11.Value = "no"
		}
		cell11.SetStyle(&style)

		autoNumber++
	}

	lock.Lock()
	err = file.Save(fileLocation)
	lock.Unlock()

	if err != nil {
		tk.Println(err.Error())
		return nil
	}

	// param := "?time=" + time.Now().Format("2006-01-02-15-04-05")
	// http.Redirect(k.Writer, k.Request, "/static/doc/List-ProjectProfile.xlsx"+param, http.StatusTemporaryRedirect)

	return fileName
}

func (c *ProjectController) DeleteProject(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	p := struct {
		Id string
	}{}
	e := k.GetPayload(&p)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	result := new(ProjectModel)
	e = c.Ctx.GetById(result, p.Id)
	if e != nil {
		return c.SetResultInfo(true, e.Error(), nil)
	}

	e = c.Ctx.Delete(result)

	return c.SetResultInfo(false, "OK", nil)
}
