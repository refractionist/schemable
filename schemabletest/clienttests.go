package schemabletest

import (
	"context"
	"testing"

	"github.com/refractionist/schemable"
)

func Run(t *testing.T, c *schemable.DBClient) {
	t.Run("DBClient", func(t *testing.T) {
		ctx := context.Background()
		if c2 := schemable.ClientFrom(ctx); c2 != nil {
			t.Error("empty ctx has client")
		}

		dbctx := schemable.WithClient(ctx, c)
		if c2 := schemable.ClientFrom(dbctx); c2 != c {
			t.Logf("CLIENT: %+v", c)
			t.Logf("CLIENT 2: %+v", c2)
			t.Error("wrong client in ctx")
		}

		t.Run("Builders from context", func(t *testing.T) {
			t.Run("Select()", func(t *testing.T) {
				c, b := schemable.Select(dbctx, "table", "col1", "col2")
				if c == nil {
					t.Fatal("client is nil")
				}
				sql, _, err := b.ToSql()
				if err != nil {
					t.Fatal(err)
				}
				if sql != "SELECT col1, col2 FROM table" {
					t.Errorf("invalid sql: %s", sql)
				}
			})

			t.Run("Insert()", func(t *testing.T) {
				c, b := schemable.Insert(dbctx, "table")
				if c == nil {
					t.Fatal("client is nil")
				}
				sql, _, err := b.Values(1, 2).ToSql()
				if err != nil {
					t.Fatal(err)
				}
				if sql != "INSERT INTO table VALUES (?,?)" {
					t.Errorf("invalid sql: %s", sql)
				}
			})

			t.Run("Update()", func(t *testing.T) {
				c, b := schemable.Update(dbctx, "table")
				if c == nil {
					t.Fatal("client is nil")
				}
				sql, _, err := b.SetMap(map[string]any{"col1": 1}).ToSql()
				if err != nil {
					t.Fatal(err)
				}
				if sql != "UPDATE table SET col1 = ?" {
					t.Errorf("invalid sql: %s", sql)
				}
			})

			t.Run("Delete()", func(t *testing.T) {
				c, b := schemable.Delete(dbctx, "table")
				if c == nil {
					t.Fatal("client is nil")
				}
				sql, _, err := b.ToSql()
				if err != nil {
					t.Fatal(err)
				}
				if sql != "DELETE FROM table" {
					t.Errorf("invalid sql: %s", sql)
				}
			})
		})

		RecorderTests(t, dbctx)
		SchemerTests(t, dbctx)
	})

	TransactionTests(t, c)

	t.Run("Targets()", func(t *testing.T) {
		recs := []*schemable.Recorder[ComicTitle]{
			ComicTitles.Record(&ComicTitle{ID: 1}),
			ComicTitles.Record(&ComicTitle{ID: 2}),
			ComicTitles.Record(&ComicTitle{ID: 3}),
		}

		targets := schemable.Targets[ComicTitle](recs)
		if len(targets) != 3 {
			t.Fatalf("invalid targets: %T %+v", targets, targets)
		}
		if targets[0].ID != 1 || targets[1].ID != 2 || targets[2].ID != 3 {
			for i, tg := range targets {
				t.Errorf("target %d: %T %+v", i, tg, tg)
			}
		}
	})
}
