package orm

import (
	"fmt"
	"strings"

	"github.com/creativelab/config"
	"github.com/creativelab/dbox"
	err "github.com/creativelab/errorlib"
	tk "github.com/creativelab/toolkit"
)

const (
	ConfigWhere  string = "where"
	ConfigOrder         = "order"
	ConfigSort          = "order"
	ConfigTake          = "limit"
	ConfigLimit         = "limit"
	ConfigTop           = "limit"
	ConfigSkip          = "skip"
	ConfigSelect        = "select"
)

type DataContext struct {
	//Adapter dbox.IAdapter
	ConnectionName string
	Connection     dbox.IConnection

	pooling bool
	//adapters       map[string]dbox.IAdapter
}

func (d *DataContext) NewModel(m IModel) IModel {
	m.SetM(m)
	return m
}

func (d *DataContext) SetPooling(p bool) *DataContext {
	d.pooling = p
	return d
}

func (d *DataContext) Pooling() bool {
	return d.pooling
}

func New(conn dbox.IConnection) *DataContext {
	ctx := new(DataContext)
	ctx.Connection = conn
	//ctx.adapters = map[string]dbox.IAdapter{}
	return ctx
}

func NewFromConfig(name string) (*DataContext, error) {
	ctx := new(DataContext)
	//ctx.adapters = map[string]dbox.IAdapter{}
	eSet := ctx.setConnectionFromConfigFile(name)
	if eSet != nil {
		return ctx, eSet
	}
	return ctx, nil
}

func (d *DataContext) Find(m IModel, parms tk.M) (dbox.ICursor, error) {
	////_ = "breakpoint"
	q := d.Connection.NewQuery().From(m.TableName())
	if qe := parms.Get(ConfigSelect); qe != nil {
		fields := qe.(string)
		selectFields := strings.Split(fields, ",")
		q = q.Select(selectFields...)
	}
	if qe := parms.Get(ConfigWhere, nil); qe != nil {
		q = q.Where(qe.(*dbox.Filter))
	}
	if qe := parms.Get(ConfigOrder, nil); qe != nil {
		q = q.Order(qe.([]string)...)
	}
	if qe := parms.Get(ConfigSkip, nil); qe != nil {
		q = q.Skip(qe.(int))
	}
	if qe := parms.Get(ConfigLimit, nil); qe != nil {
		q = q.Take(qe.(int))
	}
	//fmt.Printf("Debug Q: %s\n", tk.JsonString(q))
	return q.Cursor(nil)
	//return c
}

func (d *DataContext) Get(m IModel, config tk.M) error {
	var e error
	q := d.Connection.NewQuery().SetConfig("pooling", d.Pooling()).From(m.(IModel).TableName())
	if config.Has(ConfigWhere) {
		q = q.Where(config.Get(ConfigWhere).(*dbox.Filter))
	}
	if config.Has(ConfigOrder) {
		q = q.Order(config.Get(ConfigOrder).([]string)...)
	}
	q = q.Take(1)
	//q := d.Connection.NewQuery().From(m.(IModel).TableName()).Where(dbox.Eq("_id", id))
	c, e := q.Cursor(nil)
	if e != nil {
		return err.Error(packageName, modCtx, "Get", "Cursor fail. "+e.Error())
	}
	defer c.Close()
	e = c.Fetch(m, 1, false)
	if e != nil {
		return err.Error(packageName, modCtx, "Get", e.Error())
	}
	return nil
}

func (d *DataContext) GetById(m IModel, id interface{}) error {
	return d.Get(m, tk.M{}.Set(ConfigWhere, dbox.Eq("_id", id)))
}

func (d *DataContext) Insert(m IModel) error {
	q := d.Connection.NewQuery().SetConfig("pooling", d.Pooling()).From(m.TableName()).Insert()
	e := q.Exec(tk.M{"data": m})
	return e
}

func (d *DataContext) InsertOut(m IModel) (int64, error) {
	q := d.Connection.NewQuery().SetConfig("pooling", d.Pooling()).From(m.TableName()).Insert()
	id, e := q.ExecOut(tk.M{"data": m})
	return id, e
}

func (d *DataContext) InsertBulk(m []IModel) (e error) {
	if len(m) > 0 {
		q := d.Connection.NewQuery().SetConfig("pooling", d.Pooling()).From(m[0].TableName()).Insert()
		e = q.Exec(tk.M{"data": m})
	} else {
		e = err.Error(packageName, modCtx, "InsertBulk", "No Data")
	}
	return
}

func (d *DataContext) Save(m IModel) error {
	var e error
	if m.RecordID() == nil {
		m.PrepareID()
		if tk.IsNilOrEmpty(m.RecordID()) {
			return err.Error(packageName, modCtx, "Save", "No ID")
		}
	}
	if e = m.PreSave(); e != nil {
		return err.Error(packageName, modCtx, m.TableName()+".PreSave", e.Error())
	}
	q := d.Connection.NewQuery().SetConfig("pooling", d.Pooling()).SetConfig("multiexec", true).From(m.TableName()).Save()
	defer q.Close()
	e = q.Exec(tk.M{"data": m})
	if e != nil {
		return err.Error(packageName, modCtx, "Save", e.Error())
	}
	if e = m.PostSave(); e != nil {
		return err.Error(packageName, modCtx, m.TableName()+",PostSave", e.Error())
	}
	return e
}

func (d *DataContext) Delete(m IModel) error {
	q := d.Connection.NewQuery().SetConfig("pooling", d.Pooling()).From(m.TableName()).Delete()
	e := q.Exec(tk.M{"data": m})
	return e
}

func (d *DataContext) DeleteMany(m IModel, where *dbox.Filter) error {
	var e error
	q := d.Connection.NewQuery().SetConfig("pooling", d.Pooling()).From(m.TableName()).Delete()
	if where != nil {
		q.Where(where)
	}
	e = q.Exec(nil)
	return e
}

func (d *DataContext) Close() {
	d.Connection.Close()
}

func (d *DataContext) setConnectionFromConfigFile(name string) error {
	d.ConnectionName = name
	if d.ConnectionName == "" {
		d.ConnectionName = fmt.Sprintf("Default")
	}

	connType := strings.ToLower(config.Get("Connection_" + d.ConnectionName + "_Type").(string))
	host := config.Get("Connection_" + d.ConnectionName + "_Host").(string)
	username := config.Get("Connection_" + d.ConnectionName + "_Username").(string)
	password := config.Get("Connection_" + d.ConnectionName + "_Password").(string)
	database := config.Get("Connection_" + d.ConnectionName + "_database").(string)

	ci := new(dbox.ConnectionInfo)
	ci.Host = host
	ci.UserName = username
	ci.Password = password
	ci.Database = database

	conn, eConnect := dbox.NewConnection(connType, ci)
	if eConnect != nil {
		return err.Error(packageName, modCtx, "SetConnectionFromConfigFile", eConnect.Error())
	}
	if eConnect = conn.Connect(); eConnect != nil {
		return err.Error(packageName, modCtx, "SetConnectionFromConfigFile", eConnect.Error())
	}
	d.Connection = conn
	return nil
}
