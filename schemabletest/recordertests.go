package schemabletest

import (
	"context"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func RecorderTests(t *testing.T, ctx context.Context) {
	t.Run("Recorder", func(t *testing.T) {
		t.Run("Insert()", func(t *testing.T) {
			t.Run("nil target", func(t *testing.T) {
				rec := TestSchemer.Record(nil)
				rec.Target.ID2 = 1
				rec.Target.Name = "one"
				rec.Target.Num = 100
				if err := rec.Insert(ctx); err != nil {
					t.Fatal(err)
				}
			})

			rec := TestSchemer.Record(&TestStruct{
				ID2:  2,
				Name: "two",
				Num:  200,
			})
			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("Exists()", func(t *testing.T) {
			t.Run("nil target", func(t *testing.T) {
				rec := TestSchemer.Record(nil)
				ok, err := rec.Exists(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if ok {
					t.Error("empty record exists")
				}
			})

			t.Run("invalid ID", func(t *testing.T) {
				rec := TestSchemer.Record(&TestStruct{
					ID:  100,
					ID2: 1000,
				})
				ok, err := rec.Exists(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if ok {
					t.Error("record exists")
				}
			})

			rec := TestSchemer.Record(&TestStruct{
				ID:  1,
				ID2: 1,
			})
			ok, err := rec.Exists(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Error("record does not exist")
			}
		})

		t.Run("Load()", func(t *testing.T) {
			t.Run("nil target", func(t *testing.T) {
				rec := TestSchemer.Record(nil)
				if err := rec.Load(ctx); err == nil {
					t.Error("loaded nil record")
				}
			})

			rec := TestSchemer.Record(&TestStruct{
				ID:  1,
				ID2: 1,
			})

			if err := rec.Load(ctx); err != nil {
				t.Fatal(err)
			}

			if rec.Target.Name != "one" {
				t.Errorf("unexpected Name: %q", rec.Target.Name)
			}

			if rec.Target.Num != 100 {
				t.Errorf("unexpected Num: %d", rec.Target.Num)
			}

			if t.Failed() {
				t.Logf("Loaded: %+v", rec.Target)
			}
		})

		t.Run("LoadWhere()", func(t *testing.T) {
			rec := TestSchemer.Record(nil)

			t.Run("invalid where clause", func(t *testing.T) {
				if err := rec.LoadWhere(ctx, sq.Eq{"name": "invalid"}); err == nil {
					t.Fatalf("invalid query. target: %+v", rec.Target)
				}
			})

			if err := rec.LoadWhere(ctx, sq.Eq{"name": "one"}); err != nil {
				t.Fatal(err)
			}

			if rec.Target.ID != 1 {
				t.Errorf("unexpected ID: %d", rec.Target.ID)
			}

			if rec.Target.ID2 != 1 {
				t.Errorf("unexpected ID2: %d", rec.Target.ID2)
			}

			if rec.Target.Name != "one" {
				t.Errorf("unexpected Name: %q", rec.Target.Name)
			}

			if rec.Target.Num != 100 {
				t.Errorf("unexpected Num: %d", rec.Target.Num)
			}
		})

		t.Run("Update()", func(t *testing.T) {
			rec := TestSchemer.Record(&TestStruct{
				ID:  2,
				ID2: 2,
			})

			if err := rec.Load(ctx); err != nil {
				t.Fatal(err)
			}

			if rec.Target.Name != "two" {
				t.Errorf("unexpected Name: %q", rec.Target.Name)
			}

			if rec.Target.Num != 200 {
				t.Errorf("unexpected Num: %d", rec.Target.Num)
			}

			rec.Target.Num = 201
			if err := rec.Update(ctx); err != nil {
				t.Fatal(err)
			}

			rec2 := TestSchemer.Record(&TestStruct{
				ID:  2,
				ID2: 2,
			})
			if err := rec2.Load(ctx); err != nil {
				t.Fatal(err)
			}
			if rec2.Target.Name != "two" {
				t.Errorf("unexpected Name: %q", rec2.Target.Name)
			}

			if rec2.Target.Num != 201 {
				t.Errorf("unexpected Num: %d", rec2.Target.Num)
			}
		})

		t.Run("UpdatedFields()", func(t *testing.T) {
			rec := TestSchemer.Record(&TestStruct{
				ID2:  3,
				Name: "three",
				Num:  300,
			})

			inserting := rec.UpdatedFields()
			if val := inserting["name"]; val != "three" {
				t.Errorf("unexpected Name: %q", val)
			}
			if val := inserting["num"]; val != 300 {
				t.Errorf("unexpected Num: %T %+v", val, val)
			}
			if val := inserting["id"]; val != nil {
				t.Errorf("ID is set: %T %+v", val, val)
			}
			if val := inserting["id_two"]; val != nil {
				t.Errorf("ID2 is set: %T %+v", val, val)
			}

			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}

			postinsert := rec.UpdatedFields()
			if l := len(postinsert); l != 0 {
				t.Error("updated fields should be empty")
			}

			rec.Target.Num = 301
			preupdate := rec.UpdatedFields()
			if val := preupdate["name"]; val != nil {
				t.Errorf("Name is set: %T %+v", val, val)
			}
			if val := preupdate["num"]; val != 301 {
				t.Errorf("unexpected Num: %T %+v", val, val)
			}
			if val := preupdate["id"]; val != nil {
				t.Errorf("ID is set: %T %+v", val, val)
			}
			if val := preupdate["id_two"]; val != nil {
				t.Errorf("ID2 is set: %T %+v", val, val)
			}

			if err := rec.Update(ctx); err != nil {
				t.Fatal(err)
			}

			postupdate := rec.UpdatedFields()
			if l := len(postupdate); l != 0 {
				t.Error("updated fields should be empty")
			}
		})

		t.Run("AllFields()", func(t *testing.T) {
			rec := TestSchemer.Record(&TestStruct{
				ID:   1,
				ID2:  2,
				Name: "FieldMap",
				Num:  7,
			})

			fmap := rec.AllFields()
			if val := fmap["name"]; val != "FieldMap" {
				t.Errorf("unexpected Name: %q", val)
			}
			if val := fmap["num"]; val != 7 {
				t.Errorf("unexpected Num: %T %+v", val, val)
			}
			if val := fmap["id"]; val != nil {
				t.Errorf("ID is set: %T %+v", val, val)
			}
			if val := fmap["id_two"]; val != nil {
				t.Errorf("ID2 is set: %T %+v", val, val)
			}
		})

		t.Run("WhereIDs()", func(t *testing.T) {
			rec := TestSchemer.Record(&TestStruct{
				ID:  1,
				ID2: 2,
			})

			clause := rec.WhereIDs()
			if val := clause["test_structs.id"]; val != int64(1) {
				t.Errorf("unexpected ID: %T %+v", val, val)
			}

			if val := clause["test_structs.id_two"]; val != int64(2) {
				t.Errorf("unexpected ID2: %T %+v", val, val)
			}

			if t.Failed() {
				t.Logf("clause: %+v", clause)
			}
		})
	})
}
