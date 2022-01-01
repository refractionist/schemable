package schemable

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// Recorder records changes to Target of type T using its Schemer.
type Recorder[T any] struct {
	Schemer *Schemer[T]
	Target *T
	values map[string]any
}

// Load reloads the Recorder Target's columns (except primary keys) from the
// database.
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
		r.setValues()
	}
	return err
}

// LoadWhere loads a single Recorder Target using the given predicate args for
// a Where clause on the squirrel query builder.
func (r *Recorder[T]) LoadWhere(ctx context.Context, pred any, args ...any) error {
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
		r.setValues()
	}
	return err
}

// Exists checks if the given Recorder Target exists in the db according to its
// primary keys.
func (r *Recorder[T]) Exists(ctx context.Context) (bool, error) {
	return r.Schemer.Exists(ctx, r.WhereIDs())
}

// Insert uses the Recorder's Schemer to insert the Target into the database.
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
	r.setValues()
	return nil
}

// Insert uses the Recorder's Schemer to update the Target in the database,
// skipping if no values were updated since this Recorder was instantiated.
func (r *Recorder[T]) Update(ctx context.Context) error {
	updates := r.UpdatedValues()
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
		r.setValues()
	}
	return err
}

// Delete removes this Recorder's Target from its Schemer's table in the
// database.
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

// UpdatedValues returns an updated column/value map since this Recorder was
// instantiated.
func (r *Recorder[T]) UpdatedValues() map[string]any {
	values := r.Values()
	if r.values == nil {
		return values
	}

	for col, val := range values {
		orig, ok := r.values[col]
		if ok && orig == val {
			delete(values, col)
		}
	}
	return values
}

// Values returns a column/value map of this Recorder Target's values, except
// the primary keys.
func (r *Recorder[T]) Values() map[string]any {
	cols, vals := r.colValLists(false, true)
	update := make(map[string]any, len(cols))
	for i, col := range cols {
		update[col] = vals[i]
	}
	return update
}

// WhereIDs returns a column/value map of this Recorder Target's primary keys.
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
		// we want the address of field
		refs = append(refs, fv.Addr().Interface())
	}

	return refs
}

func (r *Recorder[T]) setValues() {
	r.values = r.Values()
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
		default:
			// get the value pointed to by the field
			v = reflect.Indirect(f)
		}

		values = append(values, v.Interface())
		columns = append(columns, field.column)
	}

	return
}