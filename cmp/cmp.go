package cmp

import (
	"bytes"
	"encoding/json"
	"microtest/vars"
	"strings"

	"reflect"
)

type Comparator struct {
	IsRaw     bool `yaml:"is_raw"`
	IsLeast   bool `yaml:"is_least"`
	IsOrdered bool `yaml:"is_ordered"`

	OverrideVars []string `yaml:"override"`

	overrideMap map[string]struct{}

	vars vars.Map
}

func (c *Comparator) SetVars(m vars.Map) {
	c.vars = m
}

func (c *Comparator) CmpBody(r, ex []byte) error {
	if c.IsRaw {
		if !bytes.Equal(r, ex) {
			return NewErrNotEqual(string(r), string(ex))
		}
		return nil
	}
	result := map[string]interface{}{}
	err := json.Unmarshal(r, &result)
	if err != nil {
		return err
	}

	expect := map[string]interface{}{}
	err = json.Unmarshal(ex, &expect)
	if err != nil {
		return err
	}
	return c.Compare(result, expect)
}

func (c *Comparator) Compare(result, expect interface{}) error {
	if v, ok := expect.(string); ok && strings.HasPrefix(v, "$") {
		if c.isOverride(v[1:]) {
			c.vars.Add(v[1:], result)
			return nil
		}

		if r, ok := result.(string); ok && r == v {
			return nil
		}

		if variable, ok := c.vars[v[1:]]; ok {
			expect = variable
		}
	}

	if reflect.TypeOf(result) != reflect.TypeOf(expect) {
		return NewErrDifferentTypes(result, expect)
	}
	switch r := result.(type) {
	case map[string]interface{}:
		return c.cmpMap(r, expect.(map[string]interface{}))
	case []interface{}:
		return c.cmpSlice(r, expect.([]interface{}))
	case int:
		if r != expect.(int) {
			return NewErrNotEqual(result, expect)
		}
	case int64:
		if r != expect.(int64) {
			return NewErrNotEqual(result, expect)
		}
	case string:
		if r != expect.(string) {
			return NewErrNotEqual(result, expect)
		}
	case float64:
		if r != expect.(float64) {
			return NewErrNotEqual(result, expect)
		}
	case nil:
		if expect != nil {
			return NewErrNotEqual(result, expect)
		}
	default:
		return ErrUnknownType
	}
	// if result != expect {
	// 	return NewErrNotEqual(result, expect)
	// }
	return nil
}

func (c *Comparator) cmpMap(result, expect map[string]interface{}) error {
	if !c.IsLeast && len(result) != len(expect) {
		return ErrLenMapsNotEquals
	}
	var r interface{}
	var ok bool
	var err error
	for k, v := range expect {
		if r, ok = result[k]; !ok {
			return NewErrFieldNotFound(k)
		}
		err = c.Compare(r, v)
		if err != nil {
			return NewErrCmpField(k, err)
		}
	}
	return nil
}

// func (c *Comparator) cmpSlice(result, expect []interface{}) error {
// 	if !c.IsLeast && len(result) != len(expect) {
// 		return ErrLenSlicesNotEqueals
// 	}

// 	if !c.IsOrdered {
// 		return c.cmpSliceUnordered(result, expect)
// 	}

// 	var err error
// 	for i, exp := range expect {
// 		err = c.Compare(result[i], exp)
// 		if err != nil {
// 			return NewErrCmpIndex(i, err)
// 		}
// 	}

// 	return nil
// }

func (c *Comparator) cmpSlice(result, expect []interface{}) error {
	if !c.IsLeast && len(result) != len(expect) {
		return ErrLenSlicesNotEquals
	}

	var err error
	var idx int
	var marks = make([]bool, len(result))

	for ie, exp := range expect {
		idx = -1

		ir := 0
		if c.IsOrdered {
			ir = ie
		}
		for ir < len(result) {

			res := result[ir]

			if marks[ir] {
				ir++
				continue
			}
			err = c.Compare(res, exp)
			if err == nil {
				idx = ir
				break
			}

			if c.IsOrdered {
				if err != nil {
					return NewErrCmpIndex(ir, err)
				}
				break
			}

			ir++
		}

		if idx == -1 {
			return ErrExpectNotFoundInArray
		}
		marks[idx] = true
	}

	return nil
}

func (c *Comparator) isOverride(key string) bool {
	if c.overrideMap == nil {
		c.overrideMap = map[string]struct{}{}
		for _, o := range c.OverrideVars {
			c.overrideMap[o] = struct{}{}
		}
	}
	_, ok := c.overrideMap[key]
	return ok
}
