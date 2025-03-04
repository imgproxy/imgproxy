package structdiff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Diffable interface {
	Diff() Entries
}

type Entry struct {
	Name  string
	Value interface{}
}

type Entries []Entry

func (d Entries) String() string {
	buf := new(strings.Builder)
	last := len(d) - 1

	for i, e := range d {
		buf.WriteString(e.Name)
		buf.WriteString(": ")

		if dd, ok := e.Value.(Entries); ok {
			buf.WriteByte('{')
			buf.WriteString(dd.String())
			buf.WriteByte('}')
		} else {
			fmt.Fprintf(buf, "%+v", e.Value)
		}

		if i != last {
			buf.WriteString("; ")
		}
	}

	return buf.String()
}

func (d Entries) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	last := len(d) - 1

	buf.WriteByte('{')

	for i, e := range d {
		j, err := json.Marshal(e.Value)
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(buf, "%q:%s", e.Name, j)

		if i != last {
			buf.WriteByte(',')
		}
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func (d Entries) flatten(m map[string]interface{}, prefix string) {
	for _, e := range d {
		key := e.Name
		if len(prefix) > 0 {
			key = prefix + "." + key
		}

		if dd, ok := e.Value.(Entries); ok {
			dd.flatten(m, key)
		} else {
			m[key] = e.Value
		}
	}
}

func (d Entries) Flatten() map[string]interface{} {
	m := make(map[string]interface{})
	d.flatten(m, "")
	return m
}

func valDiff(a, b reflect.Value) (any, bool) {
	if !a.CanInterface() || !b.CanInterface() {
		return nil, false
	}

	typeB := b.Type()

	if a.Type() != typeB {
		return b.Interface(), true
	}

	intA := a.Interface()
	intB := b.Interface()

	if reflect.DeepEqual(intA, intB) {
		return nil, false
	}

	if typeB.Kind() == reflect.Struct {
		return Diff(intA, intB), true
	}

	if typeB.Kind() == reflect.Ptr && typeB.Elem().Kind() == reflect.Struct {
		if !a.IsNil() && !b.IsNil() {
			return Diff(intA, intB), true
		}

		if !b.IsNil() {
			if diffable, ok := intB.(Diffable); ok {
				return diffable.Diff(), true
			}
		}

		return nil, true
	}

	return intB, true
}

func Diff(a, b interface{}) Entries {
	valA := reflect.Indirect(reflect.ValueOf(a))
	valB := reflect.Indirect(reflect.ValueOf(b))

	d := make(Entries, 0, valA.NumField())

	if valA.Type() != valB.Type() {
		return d
	}

	for i := 0; i < valA.NumField(); i++ {
		fieldA := valA.Field(i)
		fieldB := valB.Field(i)

		if v, ok := valDiff(fieldA, fieldB); ok {
			name := valB.Type().Field(i).Name
			d = append(d, Entry{name, v})
		}
	}

	return d
}
