package schemabletest

import (
	"context"
	"testing"
)

func RecorderTests(t *testing.T, ctx context.Context) {
	t.Run("Recorder", func(t *testing.T) {
		t.Run("insert nil target", func(t *testing.T) {
			rec := TestSchemer.Record(nil)
			rec.Target.ID2 = 1
			rec.Target.Name = "one"
			rec.Target.Num = 100
			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("insert target", func(t *testing.T) {
			rec := TestSchemer.Record(&TestStruct{
				ID2:  2,
				Name: "two",
				Num:  200,
			})
			if err := rec.Insert(ctx); err != nil {
				t.Fatal(err)
			}
		})
	})
}
