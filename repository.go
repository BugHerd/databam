package databam

import (
	"reflect"

	"github.com/serenize/snaker"
)

type Repository struct {
	d *Databam
	t reflect.Type

	table string
}

func NewRepository(d *Databam, i interface{}) (*Repository, error) {
	t := reflect.TypeOf(i)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	if t.Kind() != reflect.Struct {
		return nil, ErrNotMappable
	}

	table := ""

	for i, j := 0, t.NumField(); i < j; i++ {
		if t.Field(i).Tag.Get("table") != "" {
			table = t.Field(i).Tag.Get("table")
			break
		}
	}
	if table == "" {
		table = snaker.CamelToSnake(t.Name())
	}

	r := Repository{
		d: d,
		t: t,

		table: table,
	}

	return &r, nil
}

func (r *Repository) Select() selectQuery {
	return newSelectQuery(r)
}

func (r Repository) Load(out interface{}) error {
	return r.Fetch(out, out)
}

func (r Repository) Fetch(out interface{}, where interface{}) error {
	t := reflect.TypeOf(out)

	if t.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	t = t.Elem()

	single := true
	count := 0

	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		single = false

		if t.Kind() == reflect.Array {
			count = t.Len()
		}

		t = t.Elem()

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	if t.Kind() != reflect.Struct {
		return ErrNotMappable
	}

	q := r.Select()

	if count != 0 {
		q = q.Limit(count)
	} else if single {
		q = q.Limit(1)
	}

	v := reflect.ValueOf(where)

	if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && !v.IsNil()) {
		if v.Kind() == reflect.Ptr {
			v = reflect.Indirect(v)
		}

		if v.Type() != t {
			return ErrIncompatibleType
		}

		q.addModelToQuery(&r, where)
	}

	query, parameters, err := q.Compile()
	if err != nil {
		return err
	}

	if r.d.debugLogger != nil {
		r.d.debugLogger("Query: %s", query)
		r.d.debugLogger("Parameters: %#v", parameters)
	}

	rows, err := r.d.db.Query(query, parameters...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return r.d.mapper.RowsTo(rows, out)
}
