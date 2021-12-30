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
	return nil
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

// colValLists returns 2 lists, the column names and values.
// If withKeys is false, columns and values of fields designated as primary keys
// will not be included in those lists. Also, if withAutos is false, the returned
// lists will not include fields designated as auto-increment.
func (r *Recorder[T]) colValLists(withKeys, withAutos bool) (columns []string, values []any) {
	rt := reflect.Indirect(reflect.ValueOf(r.Target))
	/*
	if r.updates == nil {
		r.updates = make(map[string]interface{})
	}
	*/

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