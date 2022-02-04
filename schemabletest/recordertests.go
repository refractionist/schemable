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
				rec := ComicTitles.Record(nil)
				rec.Target.ID2 = 1
				rec.Target.Name = "one"
				rec.Target.Volume = 100
				if err := rec.Insert(ctx); err != nil {
					t.Fatal(err)
				}
			})

			rec := ComicTitles.Record(&ComicTitle{
				ID2:    2,
				Name:   "two",
				Volume: 200,
			})
			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("Exists()", func(t *testing.T) {
			t.Run("nil target", func(t *testing.T) {
				rec := ComicTitles.Record(nil)
				refuteExists(t, ctx, rec)
			})

			t.Run("invalid ID", func(t *testing.T) {
				rec := ComicTitles.Record(&ComicTitle{
					ID:  100,
					ID2: 1000,
				})
				refuteExists(t, ctx, rec)
			})

			rec := ComicTitles.Record(&ComicTitle{
				ID:  1,
				ID2: 1,
			})
			assertExists(t, ctx, rec)
		})

		t.Run("Load()", func(t *testing.T) {
			t.Run("nil target", func(t *testing.T) {
				rec := ComicTitles.Record(nil)
				if err := rec.Load(ctx); err == nil {
					t.Error("loaded nil record")
				}
			})

			rec := ComicTitles.Record(&ComicTitle{
				ID:  1,
				ID2: 1,
			})

			if err := rec.Load(ctx); err != nil {
				t.Fatal(err)
			}

			if rec.Target.Name != "one" {
				t.Errorf("unexpected Name: %q", rec.Target.Name)
			}

			if rec.Target.Volume != 100 {
				t.Errorf("unexpected Volume: %d", rec.Target.Volume)
			}

			if t.Failed() {
				t.Logf("Loaded: %+v", rec.Target)
			}
		})

		t.Run("LoadWhere()", func(t *testing.T) {
			rec := ComicTitles.Record(nil)

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

			if rec.Target.Volume != 100 {
				t.Errorf("unexpected Volume: %d", rec.Target.Volume)
			}
		})

		t.Run("Update()", func(t *testing.T) {
			rec := ComicTitles.Record(&ComicTitle{
				ID:  2,
				ID2: 2,
			})

			if err := rec.Load(ctx); err != nil {
				t.Fatal(err)
			}

			if rec.Target.Name != "two" {
				t.Errorf("unexpected Name: %q", rec.Target.Name)
			}

			if rec.Target.Volume != 200 {
				t.Errorf("unexpected Volume: %d", rec.Target.Volume)
			}

			rec.Target.Volume = 201
			if err := rec.Update(ctx); err != nil {
				t.Fatal(err)
			}

			rec2 := ComicTitles.Record(&ComicTitle{
				ID:  2,
				ID2: 2,
			})
			if err := rec2.Load(ctx); err != nil {
				t.Fatal(err)
			}
			if rec2.Target.Name != "two" {
				t.Errorf("unexpected Name: %q", rec2.Target.Name)
			}

			if rec2.Target.Volume != 201 {
				t.Errorf("unexpected Volume: %d", rec2.Target.Volume)
			}
		})

		t.Run("Delete()", func(t *testing.T) {
			rec := ComicTitles.Record(&ComicTitle{
				ID2:    500,
				Name:   "Delete",
				Volume: 500,
			})
			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}

			assertExists(t, ctx, rec)

			if err := rec.Delete(ctx); err != nil {
				t.Fatal(err)
			}

			refuteExists(t, ctx, rec)
		})

		t.Run("UpdatedValues()", func(t *testing.T) {
			rec := ComicTitles.Record(&ComicTitle{
				ID2:    3,
				Name:   "three",
				Volume: 300,
			})

			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}

			postinsert := rec.UpdatedValues()
			if l := len(postinsert); l != 0 {
				t.Error("updated fields should be empty")
			}

			rec.Target.Volume = 301
			preupdate := rec.UpdatedValues()
			if val := preupdate["name"]; val != nil {
				t.Errorf("Name is set: %T %+v", val, val)
			}
			if val := preupdate["volume"]; val != 301 {
				t.Errorf("unexpected Volume: %T %+v", val, val)
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

			postupdate := rec.UpdatedValues()
			if l := len(postupdate); l != 0 {
				t.Error("updated fields should be empty")
			}
		})

		t.Run("Values()", func(t *testing.T) {
			rec := ComicTitles.Record(&ComicTitle{
				ID:     1,
				ID2:    2,
				Name:   "FieldMap",
				Volume: 7,
			})

			fmap := rec.Values()
			if val := fmap["name"]; val != "FieldMap" {
				t.Errorf("unexpected Name: %q", val)
			}
			if val := fmap["volume"]; val != 7 {
				t.Errorf("unexpected Volume: %T %+v", val, val)
			}
			if val := fmap["id"]; val != nil {
				t.Errorf("ID is set: %T %+v", val, val)
			}
			if val := fmap["id_two"]; val != nil {
				t.Errorf("ID2 is set: %T %+v", val, val)
			}
		})

		t.Run("WhereIDs()", func(t *testing.T) {
			rec := ComicTitles.Record(&ComicTitle{
				ID:  1,
				ID2: 2,
			})

			clause := rec.WhereIDs()
			if val := clause["comic_titles.id"]; val != int64(1) {
				t.Errorf("unexpected ID: %T %+v", val, val)
			}

			if val := clause["comic_titles.id_two"]; val != int64(2) {
				t.Errorf("unexpected ID2: %T %+v", val, val)
			}

			if t.Failed() {
				t.Logf("clause: %+v", clause)
			}
		})
	})
}
