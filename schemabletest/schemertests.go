package schemabletest

import (
	"context"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func SchemerTests(t *testing.T, ctx context.Context) {
	t.Run("Schemer", func(t *testing.T) {
		t.Run("First()", func(t *testing.T) {
			rec, err := ComicTitles.First(ctx, func(q sq.SelectBuilder) sq.SelectBuilder {
				return q.OrderBy("id DESC")
			})
			if err != nil {
				t.Fatal(err)
			}
			if rec.Target.Name != "three" {
				recorderErr(t, rec)
			}
			if v := rec.UpdatedValues(); len(v) > 0 {
				t.Errorf("has updated values: %+v", v)
			}
		})

		t.Run("ListWhere()", func(t *testing.T) {
			recs, err := ComicTitles.ListWhere(ctx, func(q sq.SelectBuilder) sq.SelectBuilder {
				return q.Where(sq.Eq{"name": "one"})
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(recs) != 1 {
				for i, r := range recs {
					t.Logf("record %d: %+v", i, r.Target)
				}
				t.Fatal("wrong records")
			}

			if recs[0].Target.Name != "one" {
				recorderErr(t, recs[0])
			}

			if v := recs[0].UpdatedValues(); len(v) > 0 {
				t.Errorf("has updated values: %+v", v)
			}
		})

		t.Run("List()", func(t *testing.T) {
			recs, err := ComicTitles.List(ctx, 10, 1)
			if err != nil {
				t.Fatal(err)
			}

			if len(recs) == 0 {
				t.Fatal("no records")
			}

			if recs[0].Target.Name != "direct" {
				recorderErr(t, recs[0])
			}

			if v := recs[0].UpdatedValues(); len(v) > 0 {
				t.Errorf("has updated values: %+v", v)
			}
		})

		t.Run("DeleteWhere()", func(t *testing.T) {
			rec := ComicTitles.Record(&ComicTitle{
				ID2:  50,
				Name: "Deleting",
			})
			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}

			assertExists(t, ctx, rec)
			ComicTitles.DeleteWhere(ctx, func(q sq.DeleteBuilder) sq.DeleteBuilder {
				return q.Where(sq.Eq{"name": "Deleting"})
			})
		})

		t.Run("Table()", func(t *testing.T) {
			if tbl := ComicTitles.Table(); tbl != "comic_titles" {
				t.Errorf("unexpected table: %q", tbl)
			}
		})

		t.Run("Columns()", func(t *testing.T) {
			t.Run("with keys", func(t *testing.T) {
				cols := ComicTitles.Columns(true)
				if len(cols) != 4 {
					t.Fatalf("invalid columns: %+v", cols)
				}

				if cols[0] != "comic_titles.id" {
					t.Errorf("invalid col 0: %q", cols[0])
				}

				if cols[1] != "comic_titles.id_two" {
					t.Errorf("invalid col 1: %q", cols[1])
				}

				if cols[2] != "comic_titles.name" {
					t.Errorf("invalid col 2: %q", cols[2])
				}

				if cols[3] != "comic_titles.volume" {
					t.Errorf("invalid col 3: %q", cols[3])
				}
			})

			t.Run("without keys", func(t *testing.T) {
				cols := ComicTitles.Columns(false)
				if l := len(cols); l != 2 {
					t.Fatalf("invalid columns: %+v", cols)
				}

				if cols[0] != "comic_titles.name" {
					t.Errorf("invalid col 0: %q", cols[0])
				}

				if cols[1] != "comic_titles.volume" {
					t.Errorf("invalid col 1: %q", cols[1])
				}
			})
		})

		t.Run("InsertColumns()", func(t *testing.T) {
			cols := ComicTitles.InsertColumns()
			if len(cols) != 3 {
				t.Fatalf("invalid columns: %+v", cols)
			}

			if cols[0] != "id_two" {
				t.Errorf("invalid col 1: %q", cols[1])
			}

			if cols[1] != "name" {
				t.Errorf("invalid col 2: %q", cols[2])
			}

			if cols[2] != "volume" {
				t.Errorf("invalid col 3: %q", cols[3])
			}
		})
	})
}
