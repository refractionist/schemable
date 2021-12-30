package schemable

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

// WhereFunc modifies a basic select operation to add conditions.
//
// Technically, conditions are not limited to adding where clauses. It will receive
// a select statement with the 'SELECT ... FROM tablename' portion composed already.
type WhereFunc func(query sq.SelectBuilder) sq.SelectBuilder

func WithClient(ctx context.Context, c Client) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

func ClientFrom(ctx context.Context) Client {
	c, _ := ctx.Value(clientKey).(Client)
	return c
}

type key int

var clientKey = key(1)
