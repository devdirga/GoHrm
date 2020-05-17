package crowd

import (
	"errors"

	"github.com/creativelab/toolkit"
)

type FnCrowd func(x interface{}) interface{}

var Self FnCrowd = func(x interface{}) interface{} {
	return x
}

func _fn(f FnCrowd) FnCrowd {
	if f == nil {
		return Self
	} else {
		return f
	}
}

type CrowdResult struct {
	Min  interface{}
	Max  interface{}
	Avg  float64
	Sum  float64
	data interface{}
}

func (cr *CrowdResult) Data() interface{} {
	return cr.data
}

type Crowd struct {
	SliceBase
	Error  error
	Result *CrowdResult

	commands []*Command
}

func From(data interface{}) *Crowd {
	c := new(Crowd)
	c.Result = &CrowdResult{}
	e := c.SetData(data)
	if e != nil {
		c.Error = errors.New("From: " + e.Error())
	}
	return c
}

func (c *Crowd) Len() int {
	if c.data == nil {
		return 0
	}
	return toolkit.SliceLen(c.data)
}

func (c *Crowd) Sort(sortDirection SortDirection, fn FnCrowd) *Crowd {
	fn = _fn(fn)
	cmdSort := newCommand(CommandSort, fn)
	cmdSort.Parms.Set("direction", sortDirection)
	c.commands = append(c.commands, cmdSort)
	return c
}

func (c *Crowd) Clone() *Crowd {
	ccopy := new(Crowd)
	*ccopy = *c
	return ccopy
}

func (c *Crowd) Exec() *Crowd {
	defer func() {
		c.commands = []*Command{}
	}()
	if len(c.commands) == 0 {
		c.Error = errors.New("Exec: no command")
		return c
	}
	for _, cmd := range c.commands {
		e := cmd.Exec(c)
		if e != nil {
			c.Error = e
			return c
		}
	}
	return c
	/*
		    var e error
			if !cmd.isFrom {
				return c, errors.New("From data not defined.")
			}
			if cmd.isAvg {
				c.Result.Avg = 0
				l := c.Len()
				if l == 0 {
					return c, nil
				}
				ret, _ := toolkit.GetEmptySliceElement(c.data)
				//toolkit.Println("Value: ", ret, reflect.TypeOf(ret).String())
				if !toolkit.IsNumber(ret) {
					return c, nil
				}

				fn := _fn(cmd.fnAvg)
				sum := float64(0)
				for i := 0; i < l; i++ {
					item := toolkit.ToFloat64(fn(c.Item(i)), 4, toolkit.RoundingAuto)
					sum += item
				}
				//e := toolkit.Serde(sum, &ret, "json")
				c.Result.Avg = sum / float64(l)
				//		return c, nil
			}
			if cmd.isMin {
				var min interface{}
				l := c.Len()

				//min, _ = toolkit.GetEmptySliceElement(c.data)
				fn := _fn(cmd.fnMin)
				for i := 0; i < l; i++ {
					item := fn(c.Item(i))
					if item == int(0) {
						toolkit.Println("Item ", i, "=0")
					}
					if i == 0 {
						min = item
					} else if toolkit.Compare(min, item, "$gt") {
						min = item
					}
				}
				c.Result.Min = min
				//		return c, nil
			}
			if cmd.isMax {
				var max interface{}
				l := c.Len()

				max, _ = toolkit.GetEmptySliceElement(c.data)
				fn := _fn(cmd.fnMax)
				for i := 0; i < l; i++ {
					item := fn(c.Item(i))
					if i == 0 {
						max = item
					} else if toolkit.Compare(max, item, "$lt") {
						max = item
					}
				}
				c.Result.Max = max
			}
			if cmd.isSum {
				l := c.Len()

				ret, _ := toolkit.GetEmptySliceElement(c.data)
				//toolkit.Println("Value: ", ret, reflect.TypeOf(ret).String())
				if !toolkit.IsNumber(ret) {
					c.Result.Sum = 0
				}

				fn := _fn(cmd.fnSum)
				sum := float64(0)
				for i := 0; i < l; i++ {
					item := toolkit.ToFloat64(fn(c.Item(i)), 4, toolkit.RoundingAuto)
					sum += item
				}
				//e := toolkit.Serde(sum, &ret, "json")

				c.Result.Sum = sum
			}
			if cmd.isGroup {
				ret := map[interface{}][]interface{}{}
				l := c.Len()
				fnKey := _fn(cmd.fnGroupKey)
				fnChild := _fn(cmd.fnGroupChild)
				for i := 0; i < l; i++ {
					item := c.Item(i)
					k := fnKey(item)
					v := fnChild(item)
					_, has := ret[k]
					if !has {
						ret[k] = []interface{}{}
					}
					ret[k] = append(ret[k], v)
				}
				c.Result.Group = ret
			}
			if cmd.isSort {
				l := c.Len()
				if l == 0 {
					c.Result.Sort = 0
				}

				type sk struct {
					Index   int
					SortKey interface{}
				}
				c.Result.Sort = c.data
				fn := _fn(cmd.fnSort)
				keysorter, esort := NewSorter(c.Result.Sort, fn)
				if esort != nil {
					e = errors.New("crowd.Sort: " + esort.Error())
				}
				keysorter.Sort(cmd.sortDir)
				//		c.Result.Sort = c.data
			}
			return c, e
	*/
}
