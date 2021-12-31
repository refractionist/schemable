# Schemable

Schemable provides basic struct mapping against a database, using the
[squirrel][sq] package.

[sq]: https://github.com/Masterminds/squirrel

NOTE: Only works on go 1.18, since it uses generics.

## How to Use

Schemable works with annotated structs, and schemers that bind those structs
to tables.

```go
type ComicTitle struct {
	ID     int64  `db:"id, PRIMARY KEY, AUTO INCREMENT"`
	Name   string `db:"name"`
	Volume int    `db:"num"`
}

var ComicTitles = schemable.Bind[ComicTitle]("comic_titles")
```

Initialize the client and store in a context. This lets queries take advantage
of context cancellation and timeouts.

```go
// calls sql.Open(...)
client := schemable.New("sqlite3", "connection")
client.SetLogger(...) // optional, for logging queries
ctx := schemable.WithClient(context.Background(), client)
```

Schemers can list and delete multiple records:

```go
import sq "github.com/Masterminds/squirrel"

titleRecs, err := ComicTitles.ListWhere(ctx, func(q sq.SelectBuilder) sq.SelectBuilder {
  return q.Limit(10)
})

// Target is the actual *ComicTitle instance
titleRecs[0].Target

sqlResult, err := ComicTitles.DeleteWhere(ctx, func(q sq.DeleteBuilder) sq.DeleteBuilder {
  return q.Where(sq.Eq{"id": 1})
})
```

Records are managed in Recorders that can Load, Insert, Update, and Delete.
Updating only updates fields that have changed.

```go
// initialize an empty instance
newRec := ComicTitles.Record(nil)
newRec.Target.Name = "The X-Men"
newRec.Target.Volume = 1

err := newRec.Insert(ctx)

// load record by primary key
rec := ComicTitles.Record(&ComicTitle{ID: 1})
ok, err := rec.Exists(ctx)
err = rec.Load(ctx)

// only updates name column
rec.Target.Name = "The Uncanny X-Men"
err = rec.Update(ctx)

// deletes record
err = rec.Delete(ctx)
```

Schemable works with db transactions too:

```go
// TxOptions is optional and can be nil
txc, err := client.Begin(ctx, &sql.TxOptions{...})

tctx := schemable.WithClient(ctx, txc)
txRec := ComicTitles.Record(nil)
txRec.Target.Title = "The Immortal X-Men"
err = txRec.Insert(tctx)

err = txc.Commit() // or txc.Rollback()
```

## TODO

- [x] verify sqlite support
- [ ] verify mysql support
- [ ] verify postgres support
- [ ] Automatic CI tests schemable test repos
  - [sqlitetest](https://github.com/refractionist/schemable_sqlitetest)

## Inspiration

Heavily inspired by the [structable][st] package, and this [Facilitator][f]
pattern for Go generics.

[st]: https://github.com/Masterminds/structable
[f]: https://rakyll.org/generics-facilititators