package schemable

import (
	"context"
	"errors"
	"reflect"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

type Schemer[T any] struct {
	table string
	fields []*field
	keys []*field
	kind reflect.Type
}

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

func (s *Schemer[T]) List(ctx context.Context, limit, offset uint64) ([]*Recorder[T], error) {
	return s.ListWhere(ctx, func(q sq.SelectBuilder) sq.SelectBuilder {
		return q.Limit(limit).Offset(offset)
	})
}

func (s *Schemer[T]) ListWhere(ctx context.Context, fn WhereFunc) ([]*Recorder[T], error) {
	c := ClientFrom(ctx)
	if c == nil {
		return nil, errors.New("no client in context")
	}

	q := c.Builder().Select(s.Columns(true)...).From(s.table)
	if fn != nil {
		q = fn(q)
	}

	qu, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

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
		recs = append(recs, rec)
	}
	return recs, rows.Err()
}

func (s *Schemer[T]) Record(tgt *T) *Recorder[T] {
	if tgt == nil {
		tgt = new(T)
	}
	return &Recorder[T]{Schemer: s, Target: tgt}
}

func (s *Schemer[T]) Table() string {
	return s.table
}

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