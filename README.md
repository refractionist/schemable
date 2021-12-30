# Schemable

Schemable provides basic struct mapping against a database, using the
[squirrel][sq] package.

[sq]: https://github.com/Masterminds/squirrel

NOTE: Only works on go 1.18, since it uses generics.

## How to Use

Schemable works with annotated structs, and schemers that bind those structs
to tables.

```go
type Thing struct {
	ID   int64  `db:"id, PRIMARY KEY, AUTO INCREMENT"`
	Name string `db:"name"`
	Num  int    `db:"num"`
}

var Things = schemable.Bind[Thing]("things")
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

thingRecs, err := Things.List(ctx, func(q sq.SelectBuilder) sq.SelectBuilder {
  return q.Limit(10)
})

// Target is the actual *Thing instance
thingRecs[0].Target

sqlResult, err := Things.DeleteWhere(ctx, func(q sq.DeleteBuilder) sq.DeleteBuilder {
  return q.Where(sq.Eq{"id": 1})
})
```

Records are managed in Recorders that can Load, Insert, Update, and Delete.
Updating only updates fields that have changed.

```go
newRec := ThingScanner.Record(nil)
newRec.Target.Name = "test"

err := newRec.Insert(ctx)

// load record by primary key
rec := ThingScanner.Record(&Thing{ID: 1})
ok, err := rec.Exists(ctx)
err = rec.Load(ctx)

// only updates name column
rec.Target.Name = "updated"
err = rec.Update(ctx)

// deletes record
err = rec.Delete(ctx)
```

## TODO

- [x] verify sqlite support
- [ ] verify mysql support
- [ ] verify postgres support
- [ ] Automatic CI tests schemable test repos
  - [sqlitetest][https://github.com/refractionist/schemable_sqlitetest]

## Inspiration

Heavily inspired by the [structable][st] package, and this [Facilitator][f]
pattern for Go generics.

[st]: https://github.com/Masterminds/structable
[f]: https://rakyll.org/generics-facilititators