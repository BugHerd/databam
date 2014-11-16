package databam

import (
	"fmt"
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

type queryCondition interface{}

type selectQuery struct {
	model      *Repository
	counter    int
	aliases    map[*Repository]string
	fields     []queryField
	joins      []queryJoin
	conditions []queryCondition
	offset     int
	limit      int
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
	q.conditions = append(q.conditions, c)

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
			fmt.Sprintf("%s.%s", q.aliasFor(j.remote.model), q.model.d.mapper.FieldToColumn(j.remote.field)),
			fmt.Sprintf("%s.%s", q.aliasFor(j.source.model), q.model.d.mapper.FieldToColumn(j.source.field)),
		))
	}

	if q.offset != 0 {
		bits = append(bits, fmt.Sprintf("OFFSET %d", q.offset))
	}
	if q.limit != 0 {
		bits = append(bits, fmt.Sprintf("LIMIT %d", q.limit))
	}

	pretty.Println(q)

	return strings.Join(bits, " "), []interface{}{}, nil
}
