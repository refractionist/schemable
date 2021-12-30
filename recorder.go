package schemable

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type Recorder[T any] struct {
	Schemer *Schemer[T]
	Target *T
	fields map[string]any
}

func (r *Recorder[T]) Load(ctx context.Context) error {
	c := ClientFrom(ctx)
	if c == nil {
		return errors.New("no client in context")
	}

	q := c.Builder().Select(r.Schemer.Columns(false)...).From(r.Schemer.table).Where(r.WhereIDs())
	qu, args, err := q.ToSql()
	if err != nil {
		return err
	}

	refs := r.fieldRefs(false)
	err = c.QueryRow(ctx, qu, args...).Scan(refs...)
	if err == nil {
		r.setFields()
	}
	return err
}

func (r *Recorder[T]) LoadWhere(ctx context.Context, pred interface{}, args ...interface{}) error {
	c := ClientFrom(ctx)
	if c == nil {
		return errors.New("no client in context")
	}

	q := c.Builder().Select(r.Schemer.Columns(true)...).From(r.Schemer.table).Where(pred, args...)
	qu, args, err := q.ToSql()
	if err != nil {
		return err
	}

	refs := r.fieldRefs(true)
	err = c.QueryRow(ctx, qu, args...).Scan(refs...)
	if err == nil {
		r.setFields()
	}
	return err
}

func (r *Recorder[T]) Exists(ctx context.Context) (bool, error) {
	return r.ExistsWhere(ctx, r.WhereIDs())
}

func (r *Recorder[T]) ExistsWhere(ctx context.Context, pred any, args ...any) (bool, error) {
	c := ClientFrom(ctx)
	if c == nil {
		return false, errors.New("no client in context")
	}

	q := c.Builder().Select("COUNT(*) > 0").From(r.Schemer.table).Where(pred, args...)
	qu, args, err := q.ToSql()
	if err != nil {
		return false, err
	}

	has := false
	return has, c.QueryRow(ctx, qu, args...).Scan(&has)
}

func (r *Recorder[T]) Insert(ctx context.Context) error {
	c := ClientFrom(ctx)
	if c == nil {
		return errors.New("no client in context")
	}

	cols, vals := r.colValLists(true, false)
	q := c.Builder().Insert(r.Schemer.table).Columns(cols...).Values(vals...)
	qu, args, err := q.ToSql()
	if err != nil {
		return err
	}

	res, err := c.Exec(ctx, qu, args...)
	if err != nil {
		return err
	}

	rt := reflect.Indirect(reflect.ValueOf(r.Target))
	for _, f := range r.Schemer.fields {
		if !f.isAuto {
			continue
		}

		id, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("could not get last insert ID. did you set the db driver? %s", err)
		}

		field := rt.FieldByName(f.name)
		if !field.CanSet() {
			return fmt.Errorf("could not set %s to returned value", f.name)
		}
		field.SetInt(id)
	}
	r.setFields()
	return nil
}

func (r *Recorder[T]) Update(ctx context.Context) error {
	updates := r.UpdatedFields()
	if len(updates) == 0 {
		return nil
	}

	c := ClientFrom(ctx)
	if c == nil {
		return errors.New("no client in context")
	}

	q := c.Builder().Update(r.Schemer.table).SetMap(updates).Where(r.WhereIDs())
	qu, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = c.Exec(ctx, qu, args...)
	if err == nil {
		r.setFields()
	}
	return err
}

func (r *Recorder[T]) Delete(ctx context.Context) error {
	c := ClientFrom(ctx)
	if c == nil {
		return errors.New("no client in context")
	}

	q := c.Builder().Delete(r.Schemer.table).Where(r.WhereIDs())
	qu, args, err := q.ToSql()

	_, err = c.Exec(ctx, qu, args...)
	return err
}

func (r *Recorder[T]) UpdatedFields() map[string]any {
	fields := r.AllFields()
	if r.fields == nil {
		return fields
	}

	for col, val := range fields {
		orig, ok := r.fields[col]
		if ok && orig == val {
			delete(fields, col)
		}
	}
	return fields
}

func (r *Recorder[T]) WhereIDs() map[string]any {
	clause := make(map[string]any, len(r.Schemer.keys))

	rt := reflect.Indirect(reflect.ValueOf(r.Target))

	for _, f := range r.Schemer.keys {
		clause[f.selectcolumn] = rt.FieldByName(f.name).Interface()
	}

	return clause
}

func (r *Recorder[T]) fieldRefs(withKeys bool) []any {
	refs := make([]any, 0, len(r.Schemer.fields))

	ar := reflect.Indirect(reflect.ValueOf(r.Target))
	for _, field := range r.Schemer.fields {
		if !withKeys && field.isKey {
			continue
		}

		fv := ar.FieldByName(field.name)
		var ref reflect.Value
		switch fv.Kind() {
		default:
			// we want the address of field
			ref = fv.Addr()
		}
		refs = append(refs, ref.Interface())
	}

	return refs
}

func (r *Recorder[T]) setFields() {
	r.fields = r.AllFields()
}

func (r *Recorder[T]) AllFields() map[string]any {
	cols, vals := r.colValLists(false, true)
	update := make(map[string]any, len(cols))
	for i, col := range cols {
		update[col] = vals[i]
	}
	return update
}

// colValLists returns 2 lists, the column names and values.
// If withKeys is false, columns and values of fields designated as primary keys
// will not be included in those lists. Also, if withAutos is false, the returned
// lists will not include fields designated as auto-increment.
func (r *Recorder[T]) colValLists(withKeys, withAutos bool) (columns []string, values []any) {
	rt := reflect.Indirect(reflect.ValueOf(r.Target))

	for _, field := range r.Schemer.fields {
		switch {
		case !withKeys && field.isKey:
			continue
		case !withAutos && field.isAuto:
			continue
		}

		// Get the value of the field we are going to store.
		f := rt.FieldByName(field.name)
		var v reflect.Value
		switch f.Kind() {
		case reflect.Ptr:
			if f.IsNil() {
				// nothing to store
				v = reflect.Zero(f.Type())
			} else {
				// no indirection: the field is already a reference to its value
				v = f.Elem()
			}
		case reflect.Map, reflect.Slice, reflect.Struct:
			by, _ := json.Marshal(f.Interface())
			v = reflect.ValueOf(string(by))
		default:
			// get the value pointed to by the field
			v = reflect.Indirect(f)
		}

		values = append(values, v.Interface())
		columns = append(columns, field.column)
	}

	return
}