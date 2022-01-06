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
			for i, t := range targets {
				t.Errorf("target %d: %T %+v", i, t, t)
			}
		}
	})
}
