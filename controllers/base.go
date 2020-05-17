package controllers

import (
	_ "creativelab/ecleave-dev/helper"
	. "creativelab/ecleave-dev/models"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "github.com/creativelab/dbox"
	"github.com/creativelab/knot/knot.v1"
	"github.com/creativelab/orm"
	tk "github.com/creativelab/toolkit"
)

type IBaseController interface {
}
type BaseController struct {
	base       IBaseController
	Ctx        *orm.DataContext
	UploadPath string
	PdfPath    string
	LogoFile   string
	DocPath    string
}

type PageInfo struct {
	PageTitle    string
	SelectedMenu string
	Breadcrumbs  map[string]string
}

type ResultInfo struct {
	IsError bool
	Message string
	Data    interface{}
}

type Previlege struct {
	View         bool
	Create       bool
	Edit         bool
	Delete       bool
	Approve      bool
	Process      bool
	Menuid       string
	Menuname     string
	Username     string
	JobRoleName  string
	JobRoleLevel int
}

func (b *BaseController) LoadBase(k *knot.WebContext) []tk.M {
	k.Config.NoLog = true
	b.IsAuthenticate(k)
	access := b.AccessMenu(k)
	return access
}

func (b *BaseController) IsAuthenticate(k *knot.WebContext) {
	if k.Session("userid") == nil {
		b.Redirect(k, "login", "default")
	}
	return
}

func (b *BaseController) SetViewData(k *knot.WebContext, viewData tk.M) tk.M {
	access := b.LoadBase(k)
	if viewData == nil {
		viewData = tk.M{}
	}

	for _, o := range access {
		viewData.Set("Create", o["Create"].(bool))
		viewData.Set("View", o["View"].(bool))
		viewData.Set("Delete", o["Delete"].(bool))
		viewData.Set("Process", o["Process"].(bool))
		viewData.Set("Edit", o["Edit"].(bool))
		viewData.Set("Menuid", o["Menuid"].(string))
		viewData.Set("Menuname", o["Menuname"].(string))
		viewData.Set("Approve", o["Approve"].(bool))
		viewData.Set("Username", o["Username"].(string))
		viewData.Set("JobRoleName", o["JobRoleName"].(string))
		viewData.Set("JobRoleLevel", o["JobRoleLevel"].(int))
	}

	viewData.Set("UserTemplate", SysUserModel{})
	viewData.Set("UserId", k.Session("userid"))
	return viewData
}

func (b *BaseController) SetResultOK(data interface{}) *tk.Result {
	r := tk.NewResult()
	r.Data = data

	return r
}

func (b *BaseController) SetResultError(msg string, data interface{}) *tk.Result {
	tk.Println(msg)

	r := tk.NewResult()
	r.SetError(errors.New(msg))
	r.Data = data

	return r
}

func (b *BaseController) AccessMenu(k *knot.WebContext) []tk.M {
	url := k.Request.URL.String()
	if strings.Index(url, "?") > -1 {
		url = url[:strings.Index(url, "?")]
		//		tk.Println("URL_PARSED,", url)
	}
	sessionRoles := k.Session("roles")
	access := []tk.M{}
	if sessionRoles != nil {
		accesMenu := sessionRoles.([]SysRolesModel)
		if len(accesMenu) > 0 {
			for _, o := range accesMenu[0].Menu {
				if o.Url == url {
					obj := tk.M{}
					obj.Set("View", o.View)
					obj.Set("Create", o.Create)
					obj.Set("Approve", o.Approve)
					obj.Set("Delete", o.Delete)
					obj.Set("Process", o.Process)
					obj.Set("Edit", o.Edit)
					obj.Set("Menuid", o.Menuid)
					obj.Set("Menuname", o.Menuname)
					if k.Session("username") != nil {
						obj.Set("Username", k.Session("username").(string))
					} else {
						obj.Set("Username", "")
					}
					if k.Session("jobrolename") != nil {
						obj.Set("JobRoleName", k.Session("jobrolename").(string))
						obj.Set("JobRoleLevel", k.Session("jobrolelevel").(int))
					} else {
						obj.Set("JobRoleName", "")
						obj.Set("JobRoleLevel", "")
					}

					tk.Println("-------------", k.Session("jobrolename"))

					access = append(access, obj)
					return access
				}

			}
		}
	}
	return access
}

func (b *BaseController) IsLoggedIn(k *knot.WebContext) bool {
	if k.Session("userid") == nil {
		return false
	}
	return true
}
func (b *BaseController) GetCurrentUser(k *knot.WebContext) string {
	if k.Session("userid") == nil {
		return ""
	}

	if k.Session("username") == nil {
		return ""
	}

	return k.Session("username").(string)
}
func (b *BaseController) Redirect(k *knot.WebContext, controller string, action string) {
	log.Println("ecleave -->> redirecting to " + controller + "/" + action)
	http.Redirect(k.Writer, k.Request, "/"+controller+"/"+action, http.StatusTemporaryRedirect)
}

func (b *BaseController) WriteLog(msg interface{}) {
	log.Printf("%#v\n\r", msg)
	return
}
func (b *BaseController) SetResultInfo(isError bool, msg string, data interface{}) ResultInfo {
	r := ResultInfo{}
	r.IsError = isError
	r.Message = msg
	r.Data = data
	return r
}

func (b *BaseController) ErrorResultInfo(msg string, data interface{}) ResultInfo {
	r := ResultInfo{}
	r.IsError = true
	r.Message = msg
	r.Data = data
	return r
}
func (b *BaseController) Round(f float64) float64 {
	return math.Floor(f + .5)
}
func (b *BaseController) RoundPlus(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return b.Round(f*shift) / shift
}

//val is float value, roundon is what value must go up e.g: .6 will convert 1.23457 to 1.2346, places: how many digits do you want at the backyard :)s
func (b *BaseController) Round64Set(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	_div := math.Copysign(div, val)
	_roundOn := math.Copysign(roundOn, val)
	if _div >= _roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
func (b *BaseController) FirstMonday(year int, mn int) int {
	month := time.Month(mn)
	t := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	d0 := (8-int(t.Weekday()))%7 + 1
	s := strconv.Itoa(year) + fmt.Sprintf("%02d", mn) + fmt.Sprintf("%02d", d0)
	ret, _ := strconv.Atoi(s)
	return ret
}

func (b *BaseController) FirstWorkDay(ym string) int {
	t, err := time.Parse("2006-01-02", ym+"-01")
	if err != nil {
		fmt.Println(err.Error())
	}
	for t.Weekday() == 0 || t.Weekday() == 6 {
		if t.Weekday() == 0 {
			t = t.AddDate(0, 0, 1)
		} else if t.Weekday() == 6 {
			t = t.AddDate(0, 0, 2)
		}
	}
	ret, _ := strconv.Atoi(t.Format("20060102"))
	return ret
}

func (b *BaseController) GetNextIdSeq(collName string) (int, error) {
	ret := 0
	mdl := NewSequenceModel()
	crs, err := b.Ctx.Find(NewSequenceModel(), tk.M{}.Set("where", db.Eq("collname", collName)))
	if err != nil {
		return -9999, err
	}
	defer crs.Close()
	err = crs.Fetch(mdl, 1, false)
	if err != nil {
		return -9999, err
	}
	ret = mdl.Lastnumber + 1
	mdl.Lastnumber = ret
	b.Ctx.Save(mdl)
	return ret, nil
}
