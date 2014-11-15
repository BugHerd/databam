package databam

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/lann/squirrel"
	"github.com/serenize/snaker"
)

type Repository struct {
	d *Databam
	t reflect.Type

	table string
	alias string

	joinc int
	joins map[*Repository]string
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

		alias: "t0",
		table: table,
		joins: map[*Repository]string{},
	}

	return &r, nil
}

func (r Repository) Load(out interface{}) error {
	return r.Find(out, out)
}

func (r Repository) Find(out interface{}, where interface{}) error {
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

	q := squirrel.Select(r.alias + ".*").From(r.table + " " + r.alias)

	if count != 0 {
		q = q.Limit(uint64(count))
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

		joins, wheres, err := r.addFieldsToQuery(&r, where)
		if err != nil {
			return err
		}

		for _, j := range joins {
			q = q.Join(j)
		}
		for _, w := range wheres {
			q = q.Where(w)
		}
	}

	query, parameters, err := q.ToSql()
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

func (r *Repository) nextAlias() string {
	r.joinc++

	return "t" + strconv.Itoa(r.joinc)
}

func (r *Repository) addFieldsToQuery(source *Repository, where interface{}) ([]string, []squirrel.Eq, error) {
	v := reflect.ValueOf(where)

	joins := []string{}
	wheres := []squirrel.Eq{}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	w := squirrel.Eq{}
	for i, j := 0, v.NumField(); i < j; i++ {
		f := v.Field(i)

		if !reflect.DeepEqual(f.Interface(), reflect.Zero(f.Type()).Interface()) {
			if f.Kind() == reflect.Ptr {
				f = reflect.Indirect(f)
			}

			if f.Kind() == reflect.Int || f.Kind() == reflect.String {
				w[r.fieldToColumn(source, source.t.Field(i).Name)] = f.Interface()
			}
		}
	}
	if len(w) > 0 {
		wheres = append(wheres, w)
	}

	for i, j := 0, v.NumField(); i < j; i++ {
		f := v.Field(i)

		if !reflect.DeepEqual(f.Interface(), reflect.Zero(f.Type()).Interface()) {
			if f.Kind() == reflect.Ptr {
				f = reflect.Indirect(f)
			}

			if f.Kind() == reflect.Slice {
				for k := 0; k < f.Len(); k++ {
					f := f.Index(k)

					if f.Kind() == reflect.Ptr {
						f = reflect.Indirect(f)
					}

					join := r.join(source, source.t.Field(i).Name)

					joins = append(joins, fmt.Sprintf(
						"%s %s ON %s = %s",
						join.repository.table,
						r.joins[join.repository],
						r.fieldToColumn(join.repository, join.rfield),
						r.fieldToColumn(source, join.sfield),
					))

					if j, w, err := r.addFieldsToQuery(join.repository, f.Interface()); err != nil {
						return nil, nil, err
					} else {
						joins = append(joins, j...)
						wheres = append(wheres, w...)
					}
				}
			}

			if f.Kind() == reflect.Struct {
				join := r.join(source, source.t.Field(i).Name)

				joins = append(joins, fmt.Sprintf(
					"%s %s ON %s = %s",
					join.repository.table,
					r.joins[join.repository],
					r.fieldToColumn(join.repository, join.rfield),
					r.fieldToColumn(source, join.sfield),
				))

				if j, w, err := r.addFieldsToQuery(join.repository, f.Interface()); err != nil {
					return nil, nil, err
				} else {
					joins = append(joins, j...)
					wheres = append(wheres, w...)
				}
			}
		}
	}

	return joins, wheres, nil
}

func (r *Repository) fieldToColumn(repository *Repository, field string) string {
	return fmt.Sprintf("%s.%s", repository.alias, r.d.mapper.FieldToColumn(field))
}

type joinInfo struct {
	repository *Repository
	sfield     string
	rfield     string
}

func (r *Repository) join(source *Repository, name string) *joinInfo {
	f, ok := source.t.FieldByName(name)
	if !ok {
		return nil
	}

	t := f.Type

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
		return nil
	}

	v := reflect.Indirect(reflect.New(t))

	other := r.d.MustRepository(v.Interface())
	other.alias = r.nextAlias()

	r.joins[other] = other.alias

	remoteField := ""
	sourceField := ""

	for i := 0; i < t.NumField(); i++ {
		rf := t.Field(i).Name
		sf := name + rf

		if _, ok := source.t.FieldByName(sf); ok {
			remoteField = rf
			sourceField = sf
			break
		}
	}

	if remoteField == "" && sourceField == "" {
		if _, ok := t.FieldByName(source.t.Name() + "Id"); ok {
			remoteField = source.t.Name() + "Id"
			sourceField = "Id"
		}
	}

	return &joinInfo{
		repository: other,
		sfield:     sourceField,
		rfield:     remoteField,
	}
}
