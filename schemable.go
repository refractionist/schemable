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

// DeleteFunc modifies a basic delete operation to add conditions.
//
// Technically, conditions are not limited to adding where clauses. It will receive
// a select statement with the 'SELECT ... FROM tablename' portion composed already.
type DeleteFunc func(query sq.DeleteBuilder) sq.DeleteBuilder

// WithClient returns a modified variant of the given context with an embedded
// client.
func WithClient(ctx context.Context, c Client) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

// ClientFrom fetches the embedded client from the given context.
func ClientFrom(ctx context.Context) Client {
	c, _ := ctx.Value(clientKey).(Client)
	return c
}

type key int

var clientKey = key(1)
