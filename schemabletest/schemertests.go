package schemabletest

import (
	"context"
	"testing"
)

func SchemerTests(t *testing.T, ctx context.Context) {
	t.Run("Schemer", func(t *testing.T) {
		t.Run("ListWhere()", func(t *testing.T) {
			t.Run("nil WhereFunc", func(t *testing.T) {
				recs, err := TestSchemer.ListWhere(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}

				if len(recs) == 0 {
					t.Fatal("no records")
				}

				if recs[0].Target.Name != "one" {
					recorderErr(t, recs[0])
				}

				if recs[1].Target.Name != "two" {
					recorderErr(t, recs[1])
				}
			})
		})

		t.Run("List()", func(t *testing.T) {
			recs, err := TestSchemer.List(ctx, 10, 1)
			if err != nil {
				t.Fatal(err)
			}

			if len(recs) == 0 {
				t.Fatal("no records")
			}

			if recs[0].Target.Name != "two" {
				recorderErr(t, recs[0])
			}
		})

		t.Run("DeleteWhere()", func(t *testing.T) {
			rec := TestSchemer.Record(&TestStruct{
				ID2:  50,
				Name: "Deleting",
			})
			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("Table()", func(t *testing.T) {
			if tbl := TestSchemer.Table(); tbl != "test_structs" {
				t.Errorf("unexpected table: %q", tbl)
			}
		})

		t.Run("Columns()", func(t *testing.T) {
			t.Run("with keys", func(t *testing.T) {
				cols := TestSchemer.Columns(true)
				if len(cols) != 4 {
					t.Fatalf("invalid columns: %+v", cols)
				}

				if cols[0] != "test_structs.id" {
					t.Errorf("invalid col 0: %q", cols[0])
				}

				if cols[1] != "test_structs.id_two" {
					t.Errorf("invalid col 1: %q", cols[1])
				}

				if cols[2] != "test_structs.name" {
					t.Errorf("invalid col 2: %q", cols[2])
				}

				if cols[3] != "test_structs.num" {
					t.Errorf("invalid col 3: %q", cols[3])
				}
			})

			t.Run("without keys", func(t *testing.T) {
				cols := TestSchemer.Columns(false)
				if l := len(cols); l != 2 {
					t.Fatalf("invalid columns: %+v", cols)
				}

				if cols[0] != "test_structs.name" {
					t.Errorf("invalid col 0: %q", cols[0])
				}

				if cols[1] != "test_structs.num" {
					t.Errorf("invalid col 1: %q", cols[1])
				}
			})
		})
	})
}
