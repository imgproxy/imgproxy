package structdiff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var builderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

type Entry struct {
	Name  string
	Value interface{}
}

type Entries []Entry

func (d Entries) String() string {
	buf := builderPool.Get().(*strings.Builder)
	last := len(d) - 1

	buf.Reset()

	for i, e := range d {
		buf.WriteString(e.Name)
		buf.WriteString(": ")

		if dd, ok := e.Value.(Entries); ok {
			buf.WriteByte('{')
			buf.WriteString(dd.String())
			buf.WriteByte('}')
		} else {
			fmt.Fprint(buf, e.Value)
		}

		if i != last {
			buf.WriteString("; ")
		}
	}

	return buf.String()
}

func (d Entries) MarshalJSON() ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	last := len(d) - 1

	buf.Reset()

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

		if !fieldA.CanInterface() || !fieldB.CanInterface() {
			continue
		}

		intA := fieldA.Interface()
		intB := fieldB.Interface()

		if !reflect.DeepEqual(intA, intB) {
			name := valB.Type().Field(i).Name
			value := intB

			if fieldB.Kind() == reflect.Struct {
				value = Diff(intA, intB)
			}

			d = append(d, Entry{name, value})
		}
	}

	return d
}
