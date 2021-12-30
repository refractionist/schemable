package schemable

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

type Client interface {
	Exec(ctx context.Context, q string, args ...any) (sql.Result, error)
	Query(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	Builder() *sq.StatementBuilderType
}

type QueryLogger interface {
	LogQuery(q string, args []any)
}

type DBClient struct {
	db      *sql.DB
	logger  QueryLogger
	builder *sq.StatementBuilderType
}

func New(driver, conn string) (*DBClient, error) {
	db, err := sql.Open(driver, conn)
	if err != nil {
		return nil, err
	}

	builder := sq.StatementBuilder.RunWith(db)
	return &DBClient{db: db, logger: nilLogger, builder: &builder}, nil
}

func (c *DBClient) SetLogger(l QueryLogger) {
	c.logger = l
}

func (c *DBClient) Builder() *sq.StatementBuilderType {
	return c.builder
}

func (c *DBClient) Exec(ctx context.Context, q string, args ...any) (sql.Result, error) {
	c.logger.LogQuery(q, args)
	return c.db.ExecContext(ctx, q, args...)
}

func (c *DBClient) Query(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	c.logger.LogQuery(q, args)
	return c.db.QueryContext(ctx, q, args...)
}

func (c *DBClient) Close() error {
	return c.db.Close()
}

func (c *DBClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

type TxnClient struct {
	tx      *sql.Tx
	logger  QueryLogger
	builder *sq.StatementBuilderType
}

func (c *DBClient) Begin() (*TxnClient, error) {
	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}

	builder := sq.StatementBuilder.RunWith(tx)
	return &TxnClient{tx: tx, logger: c.logger, builder: &builder}, nil
}

func (c *TxnClient) Commit() error {
	return c.tx.Commit()
}

func (c *TxnClient) Rollback() error {
	return c.tx.Rollback()
}

func (c *TxnClient) Exec(ctx context.Context, q string, args ...any) (sql.Result, error) {
	c.logger.LogQuery(q, args)
	return c.tx.ExecContext(ctx, q, args...)
}

func (c *TxnClient) Query(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	c.logger.LogQuery(q, args)
	return c.tx.QueryContext(ctx, q, args...)
}

func (c *TxnClient) Builder() *sq.StatementBuilderType {
	return c.builder
}

type noLogger struct{}

func (l *noLogger) LogQuery(q string, args []any) {
}

var nilLogger = &noLogger{}
