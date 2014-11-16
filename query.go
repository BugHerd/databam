package databam

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kr/pretty"
)

type queryField struct {
	model *Repository
	field string
}

type queryJoin struct {
	source queryField
	remote queryField
}

type queryCondition interface {
	Compile(*selectQuery) (string, []interface{})
}

type queryConditionAnd []queryCondition

func (c *queryConditionAnd) Compile(q *selectQuery) (string, []interface{}) {
	bits := []string{}
	vals := []interface{}{}

	for _, c := range *c {
		s, v := c.Compile(q)

		bits = append(bits, s)
		vals = append(vals, v...)
	}

	if len(bits) > 1 {
		return "(" + strings.Join(bits, " AND ") + ")", vals
	} else {
		return bits[0], vals
	}
}

type queryConditionEq struct {
	field queryField
	value interface{}
}

func (c *queryConditionEq) Compile(q *selectQuery) (string, []interface{}) {
	return fmt.Sprintf("%s = ?", q.formatField(c.field)), []interface{}{c.value}
}

type selectQuery struct {
	model     *Repository
	counter   int
	aliases   map[*Repository]string
	fields    []queryField
	joins     []queryJoin
	condition queryConditionAnd
	offset    int
	limit     int
}

func newSelectQuery(model *Repository) selectQuery {
	return selectQuery{
		model: model,
		aliases: map[*Repository]string{
			model: "t",
		},
	}
}

func (q *selectQuery) aliasFor(model *Repository) string {
	if alias, ok := q.aliases[model]; ok {
		return alias
	} else {
		q.aliases[model] = fmt.Sprintf("j%d", q.counter)
		q.counter++
		return q.aliases[model]
	}
}

func (q selectQuery) Field(model *Repository, field string) selectQuery {
	q.fields = append(q.fields, queryField{
		model: model,
		field: field,
	})

	return q
}

func (q selectQuery) Join(j queryJoin) selectQuery {
	q.joins = append(q.joins, j)

	return q
}

func (q selectQuery) Where(c queryCondition) selectQuery {
	q.condition = append(q.condition, c)

	return q
}

func (q selectQuery) Offset(offset int) selectQuery {
	q.offset = offset

	return q
}

func (q selectQuery) Limit(limit int) selectQuery {
	q.limit = limit

	return q
}

func (q selectQuery) Compile() (string, []interface{}, error) {
	bits := []string{"SELECT"}
	vals := []interface{}{}

	fields := []string{}
	for _, f := range q.fields {
		fields = append(bits, fmt.Sprintf("%s.%s", q.aliasFor(f.model), f.field))
	}
	if len(fields) == 0 {
		fields = []string{
			fmt.Sprintf("%s.*", q.aliasFor(q.model)),
		}
	}

	bits = append(bits, strings.Join(fields, ", "))

	bits = append(bits, fmt.Sprintf("FROM %s %s", q.model.table, q.aliasFor(q.model)))

	for _, j := range q.joins {
		bits = append(bits, fmt.Sprintf(
			"JOIN %s %s ON %s = %s",
			j.remote.model.table,
			q.aliasFor(j.remote.model),
			q.formatField(j.remote),
			q.formatField(j.source),
		))
	}

	b, v := q.condition.Compile(&q)
	bits = append(bits, "WHERE", b)
	vals = append(vals, v...)

	if q.offset != 0 {
		bits = append(bits, fmt.Sprintf("OFFSET %d", q.offset))
	}
	if q.limit != 0 {
		bits = append(bits, fmt.Sprintf("LIMIT %d", q.limit))
	}

	pretty.Println(q)

	return strings.Join(bits, " "), vals, nil
}

func (q selectQuery) formatField(f queryField) string {
	return fmt.Sprintf(
		"%s.%s",
		q.aliasFor(f.model),
		q.model.d.mapper.FieldToColumn(f.field),
	)
}

func (q *selectQuery) addModelToQuery(source *Repository, model interface{}) error {
	v := reflect.ValueOf(model)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	w := queryConditionAnd{}
	for i, j := 0, v.NumField(); i < j; i++ {
		f := v.Field(i)

		if !reflect.DeepEqual(f.Interface(), reflect.Zero(f.Type()).Interface()) {
			if f.Kind() == reflect.Ptr {
				f = reflect.Indirect(f)
			}

			if f.Kind() == reflect.Int || f.Kind() == reflect.String {
				w = append(w, &queryConditionEq{
					field: queryField{model: source, field: source.t.Field(i).Name},
					value: f.Interface(),
				})
			}
		}
	}
	if len(w) > 0 {
		q.condition = append(q.condition, &w)
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

					join := q.makeJoinForRelation(source, source.t.Field(i).Name)

					q.joins = append(q.joins, *join)

					if err := q.addModelToQuery(join.remote.model, f.Interface()); err != nil {
						return err
					}
				}
			}

			if f.Kind() == reflect.Struct {
				join := q.makeJoinForRelation(source, source.t.Field(i).Name)

				q.joins = append(q.joins, *join)

				if err := q.addModelToQuery(join.remote.model, f.Interface()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (q *selectQuery) makeJoinForRelation(source *Repository, fieldName string) *queryJoin {
	f, ok := source.t.FieldByName(fieldName)
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

	other := q.model.d.MustRepository(v.Interface())

	remoteJoinField := ""
	sourceJoinField := ""

	for i := 0; i < t.NumField(); i++ {
		rf := t.Field(i).Name
		sf := fieldName + rf

		if _, ok := source.t.FieldByName(sf); ok {
			remoteJoinField = rf
			sourceJoinField = sf
			break
		}
	}

	if remoteJoinField == "" && sourceJoinField == "" {
		if _, ok := t.FieldByName(source.t.Name() + "Id"); ok {
			remoteJoinField = source.t.Name() + "Id"
			sourceJoinField = "Id"
		}
	}

	return &queryJoin{
		source: queryField{
			model: source,
			field: sourceJoinField,
		},
		remote: queryField{
			model: other,
			field: remoteJoinField,
		},
	}
}
