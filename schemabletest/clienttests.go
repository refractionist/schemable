package schemabletest

import (
	"context"
	"testing"

	"github.com/refractionist/schemable_sqlitetest/internal/schemable"
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
}
