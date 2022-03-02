package schemable

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// Schemer maintains the table and column mapping of the generic type T. It
// parses the column and primary key info from the struct field tags.
type Schemer[T any] struct {
	table string
	fields []*field
	keys []*field
	kind reflect.Type
}

// Bind creates a Schemer table/column mapping for the given generic type T.
func Bind[T any](table string) *Schemer[T] {
	tgt := new(T)
	fields, keys := scanFields(table, tgt)
	return &Schemer[T]{
		table: table,
		fields: fields,
		keys: keys,
		kind: reflect.TypeOf(tgt).Elem(),
	}
}

// First returns a *Recorder[T] of the first row, filtered by the given
// WhereFunc. The context must have a client embedded with WithClient().
func (s *Schemer[T]) First(ctx context.Context, fn WhereFunc) (*Recorder[T], error) {
	c := ClientFrom(ctx)
	if c == nil {
		return nil, ErrNoClient
	}

	q := fn(c.Builder().Select(s.Columns(true)...).From(s.table)).Limit(1)
	qu, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rec := s.Record(nil)
	start := time.Now()
	err = c.QueryRow(ctx, qu, args...).Scan(rec.fieldRefs(true)...)
	rec.setValues()
	c.LogQuery(WithDBDuration(ctx, start), qu, args)
	return rec, err
}

// List returns rows of type T embedded in Recorders, using the given limit
// and offset values. The context must have a client embedded with WithClient().
func (s *Schemer[T]) List(ctx context.Context, limit, offset uint64) ([]*Recorder[T], error) {
	return s.ListWhere(ctx, func(q sq.SelectBuilder) sq.SelectBuilder {
		return q.Limit(limit).Offset(offset)
	})
}

// ListWhere returns rows of type T embedded in Recorders, filtered by the
// given WhereFunc. The context must have a client embedded with WithClient().
func (s *Schemer[T]) ListWhere(ctx context.Context, fn WhereFunc) ([]*Recorder[T], error) {
	c := ClientFrom(ctx)
	if c == nil {
		return nil, ErrNoClient
	}

	q := fn(c.Builder().Select(s.Columns(true)...).From(s.table))
	qu, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	rows, err := c.Query(ctx, qu, args...)
	if err != nil || rows == nil {
		return nil, err
	}
	defer rows.Close()

	recs := make([]*Recorder[T], 0)
	var rec *Recorder[T]
	for rows.Next() {
		rec = s.Record(nil)
		err = rows.Scan(rec.fieldRefs(true)...)
		if err != nil {
			return recs, err
		}
		rec.setValues()
		recs = append(recs, rec)
	}
	c.LogQuery(WithDBDuration(ctx, start), qu, args)
	return recs, rows.Err()
}

// DeleteWhere deletes rows filtered by the given DeleteFunc. The context must
// have a client embedded with WithClient().
func (s *Schemer[T]) DeleteWhere(ctx context.Context, fn DeleteFunc) (sql.Result, error) {
	c := ClientFrom(ctx)
	if c == nil {
		return nil, ErrNoClient
	}

	q := fn(c.Builder().Delete(s.table))
	qu, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	res, err := c.Exec(ctx, qu, args...)
	c.LogQuery(WithDBDuration(ctx, start), qu, args)
	return res, err
}

// Exists checks if any Recorder Target exists using the given predicate
// args for a Where clause on the squirrel query builder.
func (s *Schemer[T]) Exists(ctx context.Context, pred any, args ...any) (bool, error) {
	c := ClientFrom(ctx)
	if c == nil {
		return false, ErrNoClient
	}

	q := c.Builder().Select("COUNT(*) > 0").From(s.table).Where(pred, args...)
	qu, args, err := q.ToSql()
	if err != nil {
		return false, err
	}

	has := false
	start := time.Now()
	err = c.QueryRow(ctx, qu, args...).Scan(&has)
	c.LogQuery(WithDBDuration(ctx, start), qu, args)
	return has, err
}

// Record returns a Recorder for the given instance, creating a new one if nil
// is provided. For updating purposes, the returned recorder's UpdatedValues()
// do not take the given target's fields into account. Be aware how default
// values (like "" or 0) will affect your database. Use one of the Load or List
// funcs before setting properties in order to take advantage of partial
// updates.
func (s *Schemer[T]) Record(tgt *T) *Recorder[T] {
	if tgt == nil {
		tgt = new(T)
	}
	return &Recorder[T]{Schemer: s, Target: tgt}
}

// Table returns the table name that the Schemer's type T uses.
func (s *Schemer[T]) Table() string {
	return s.table
}

// Columns returns the column names for the Schemer's type T, optionally with
// or without primary key columns. Columns are disambiguated with the table
// name for join queries.
func (s *Schemer[T]) Columns(withKeys bool) []string {
	names := make([]string, 0, len(s.fields))
	for _, f := range s.fields {
		if !withKeys && f.isKey {
			continue
		}
		names = append(names, f.selectcolumn)
	}
	return names
}

// Internal representation of a field on a database table, and its
// relation to a struct field.
type field struct {
	// name = Struct field name
	// column = db column name
	// selectcolumn = "table.column"
	name, column, selectcolumn string
	// Is a primary key
	isKey bool
	// Is an auto increment
	isAuto bool
	// Is optional
	isOptional bool
}

func scanFields(table string, obj any) (fields []*field, keys []*field) {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()
	count := t.NumField()
	keys = make([]*field, 0, 2)
	fields = make([]*field, 0, count)

	for i := 0; i < count; i++ {
		f := t.Field(i)
		if len(f.Tag) == 0 {
			continue
		}

		stag := f.Tag.Get("db")
		if len(stag) == 0 {
			continue
		}

		parts := parseTag(f.Name, stag)
		field := &field{
			name:         f.Name,
			column:       parts[0],
			selectcolumn: table + "." + parts[0],
			isOptional:   f.Type.Kind() == reflect.Ptr,
		}

		for _, part := range parts[1:] {
			switch strings.TrimSpace(part) {
				case pkey:
					field.isKey = true
					keys = append(keys, field)
				case autoinc:
					field.isAuto = true
			}
		}

		fields = append(fields, field)
	}

	return
}

// parseTag parses the contents of a stbl tag.
func parseTag(fieldName, tag string) []string {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return []string{fieldName}
	}
	return parts
}

const (
	pkey = "PRIMARY KEY"
	autoinc = "AUTO INCREMENT"
)