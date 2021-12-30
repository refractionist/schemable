package schemabletest

import (
	"context"
	"testing"

	"github.com/refractionist/schemable_sqlitetest/internal/schemable"
)

func TransactionTests(t *testing.T, dc *schemable.DBClient) {
	t.Run("TxnClient", func(t *testing.T) {
		dctx := schemable.WithClient(context.Background(), dc)
		rec := TestSchemer.Record(&TestStruct{
			ID2:  5,
			Name: "transaction",
			Num:  5,
		})
		if err := rec.Insert(dctx); err != nil {
			t.Fatal(err)
		}

		t.Run("Rollback", func(t *testing.T) {
			tc, err := dc.Begin()
			if err != nil {
				t.Fatal(err)
			}
			tctx := schemable.WithClient(context.Background(), tc)

			assertExists(t, tctx, rec)
			assertExists(t, dctx, rec)

			if err := rec.Load(tctx); err != nil {
				t.Fatal(err)
			}
			if err := rec.Delete(tctx); err != nil {
				t.Fatal(err)
			}
			refuteExists(t, tctx, rec)

			if err := tc.Rollback(); err != nil {
				t.Fatal(err)
			}

			assertExists(t, dctx, rec)
		})

		t.Run("Commit", func(t *testing.T) {
			tc, err := dc.Begin()
			if err != nil {
				t.Fatal(err)
			}
			tctx := schemable.WithClient(context.Background(), tc)

			assertExists(t, tctx, rec)
			assertExists(t, dctx, rec)

			if err := rec.Load(tctx); err != nil {
				t.Fatal(err)
			}
			if err := rec.Delete(tctx); err != nil {
				t.Fatal(err)
			}
			refuteExists(t, tctx, rec)

			if err := tc.Commit(); err != nil {
				t.Fatal(err)
			}

			refuteExists(t, dctx, rec)
		})
	})
}
