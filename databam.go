// package databam is an ORM designed to make querying highly-relational
// datasets easy
package databam

import (
	"database/sql"
)

type DebugLogger func(string, ...interface{})

type Databam struct {
	debugLogger DebugLogger
	db          *sql.DB
	mapper      Mapper
}

func New(db *sql.DB) *Databam {
	return NewWithMapper(db, DefaultMapper{})
}

func NewWithMapper(db *sql.DB, mapper Mapper) *Databam {
	return &Databam{
		db:     db,
		mapper: mapper,
	}
}

func (d *Databam) SetDebugLogger(debugLogger DebugLogger) {
	d.debugLogger = debugLogger
}

func (d *Databam) Repository(i interface{}) (*Repository, error) {
	return NewRepository(d, i)
}

func (d *Databam) MustRepository(i interface{}) *Repository {
	if m, err := d.Repository(i); err != nil {
		panic(err)
	} else {
		return m
	}
}

func (d *Databam) Load(out interface{}) error {
	if m, err := d.Repository(out); err != nil {
		return err
	} else {
		return m.Load(out)
	}
}
