package databam

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/serenize/snaker"
)

type Mapper interface {
	RowsTo(rows *sql.Rows, out interface{}) error
	RowTo(rows *sql.Rows, columns []string, out reflect.Value) error
	FieldsFrom(v reflect.Value, names []string) ([]reflect.Value, error)
	FieldToColumn(name string) string
}

type DefaultMapper struct{}

func (m DefaultMapper) RowsTo(rows *sql.Rows, out interface{}) error {
	t := reflect.TypeOf(out)

	if t.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	t = t.Elem()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if t.Kind() == reflect.Struct {
		if !rows.Next() {
			return nil
		}

		return m.RowTo(rows, columns, reflect.Indirect(reflect.ValueOf(out)))
	}

	if t.Kind() == reflect.Slice {
		s := reflect.ValueOf(out)
		if s.Type().Kind() == reflect.Ptr {
			s = s.Elem()
		}

		t := t.Elem()

		for {
			if !rows.Next() {
				return nil
			}

			v := reflect.New(t)

			if err := m.RowTo(rows, columns, reflect.Indirect(v)); err != nil {
				return err
			}

			s.Set(reflect.Append(s, reflect.Indirect(v)))
		}
	}

	return ErrNotMappable
}

func (m DefaultMapper) RowTo(rows *sql.Rows, columns []string, out reflect.Value) error {
	fields, err := m.FieldsFrom(out, columns)
	if err != nil {
		return err
	}

	r := reflect.ValueOf(rows).MethodByName("Scan").Call(fields)

	if !r[0].IsNil() {
		return r[0].Interface().(error)
	}

	return nil
}

func (m DefaultMapper) FieldsFrom(v reflect.Value, names []string) ([]reflect.Value, error) {
	t := v.Type()

	fields := []reflect.Value{}

outer:
	for _, originalName := range names {
		convertedName := snaker.SnakeToCamel(originalName)

		for i, j := 0, t.NumField(); i < j; i++ {
			sfield := t.Field(i)
			vfield := v.Field(i)

			if sfield.Tag.Get("sql") == "-" {
				continue
			}

			if sfield.Tag.Get("sql") == originalName || sfield.Name == originalName || sfield.Name == convertedName || strings.ToLower(sfield.Name) == strings.ToLower(originalName) || strings.ToLower(sfield.Name) == strings.ToLower(convertedName) {
				if vfield.Kind() == reflect.Ptr {
					if vfield.IsNil() {
						vfield.Set(reflect.New(vfield.Type().Elem()))
					}
				}

				fields = append(fields, vfield.Addr())

				continue outer
			}
		}

		return nil, ErrFieldUnmapped
	}

	return fields, nil
}

func (m DefaultMapper) FieldToColumn(name string) string {
	return snaker.CamelToSnake(name)
}
