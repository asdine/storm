# Storm

[![Build Status](https://travis-ci.org/asdine/storm.svg)](https://travis-ci.org/asdine/storm)
[![GoDoc](https://godoc.org/github.com/asdine/storm?status.svg)](https://godoc.org/github.com/asdine/storm)

Storm is a simple and powerful toolkit for [BoltDB](https://github.com/etcd-io/bbolt) and [Badger](https://github.com/dgraph-io/badger). Basically, Storm provides indexes, a wide range of methods to store and fetch data, support for SQL, and much more. To do so, it relies on [Genji](https://github.com/genjidb/genji), a powerful embedded, document-oriented, SQL database.

In addition to the examples below, see also the [examples in the GoDoc](https://godoc.org/github.com/asdine/storm#pkg-examples).

## Getting Started

```bash
go get -u github.com/asdine/storm/v4
```

## Import Storm

```go
import "github.com/asdine/storm/v4"
```

## Opening a database

Quick way of opening a database:

```go
db, err := storm.Open("my.db")
defer db.Close()
```

## Storing data

### Creating and creating a store

A store holds records in the underlying storage. To create one, use the `CreateStore` method:

```go
users, err := db.CreateStore("user")
```

`CreateStore` returns a store object with methods to manipulate records.

To get an already existing store, use `GetStore`:

```go
users, err := db.GetStore("user")
```

To quickly get a store and chain a call to a method, use `Store`. If the store doesn't exist, the call to the next method will fail:

```go
id, err = db.Store("user").Insert(...)
```

### Storing data

Storm supports pretty much any structure and data type. Example:

```go
type User struct {
	Name      string
	Age       int
	Email     string
	Group     string
	CreatedAt time.Time
	Address   struct {
		City    string
		ZipCode string
	}
}
```

### Inserting data

```go
user := User{
  Name: "John",
  Age: 21,
  Email: "john@provider.com",
  Group: "staff",
  CreatedAt: time.Now(),
}
user.Address.City = "Lyon"
user.Address.ZipCode = "69001"

id, err := users.Insert(&user)
```

That's it.

`Insert` serializes the record, checks for constraints, updates indexes and stores the record in the selected store.

### Creating indexes

Storm supports regular indexes, unique indexes and composite indexes.

```go
// regular index
err = db.CreateIndex("idx_user_name", "name")

// unique index
err = db.CreateUniqueIndex("idx_user_email", "name")

id, err = users.Insert(&user)
// err == storm.ErrAlreadyExists

// index on nested fields
err = db.CreateIndex("idx_user_zipcode", "address.zipcode")

// composite index
err = db.CreateIndex("idx_user_age_city", "age", "address.city")
```

#### Primary keys

Every time a record is inserted, a `docid`, a unique auto-incremented integer, is generated.
To specify a different primary key for a store, use `storm.StoreOptions` upon store creation.

```go
err = db.CreateStore("user", &storm.Options{
  PrimaryKey: "id",
})
```

Any field named `ID`, `iD` or `Id` will be used as the primary key of the store.

### Simple queries

Storm provides a simple but powerful DSL to run queries. Each query must end with one of the following methods:

```go
users := db.Store("user")

q := users.Query()

// Fetching the first record
var user User
err = q.First(&user)

// Fetching records
var users []User
err = q.Find(&users)

// Deleting records
err = q.Delete()

// Updating records
err = q.Update("age", 10, "address.zipcode", "Paris")
```

These methods can be combined with clauses to filter records:

```go
// Skip records
q.Offset(10)

// Limit the number of records
q.Limit(10)

// Order records by a certain field
q.OrderBy("age")
q.OrderBy("age", "desc")

// Filter by predicate
q.Where("age", ">", 18)
q.Eq("address.city", "Paris")
q.Neq("address.city", "Paris")
q.Gt("age", 21)
q.Gte("age", 21)
q.Lt("age", 21)
q.Lte("age", 21)
```

Example:

```go
var u User
q := db.Store("user").Query()
err = q.Eq("address.city", "Paris").Limit(10).Offset(2).OrderBy("age").First(&user)

q := db.Store("user").Eq("address.city", "Paris").Limit(10).Offset(2).OrderBy("age")

var users []*User
err = q.Find(&users)
err = q.Update("age", 10, "address.zipcode", "Paris")
err = q.Delete()
```

For more information, and to see all of the supported features, checkout the [Go doc](https://pkg.go.dev/github.com/asdine/storm).

### Advanced queries with SQL

For more complex queries, you can use SQL! Storm uses Genji behind the scenes to run SQL queries on Bolt. SQL Documentation can be found on [Genji's documentation website](https://genji.dev/docs/genji-sql/).

```go
res, err := db.Query("SELECT * FROM user WHERE age > 10")
if err != nil {
  return err
}
defer res.Close()

err = res.Scan(&user)
err = res.Iterate(func(d document.Document) error {
    var u User
    err = document.StructScan(d, &u)
    if err != nil {
        return err
    }

    fmt.Println(u)
    return nil
})
```

### Transactions

```go
tx, err := db.Begin(true)
if err != nil {
  return err
}
defer tx.Rollback()

accountA.Amount -= 100
accountB.Amount += 100

id, err = tx.Store("account").Insert(accountA)
if err != nil {
  return err
}

id, err = tx.Store("account").Insert(accountB)
if err != nil {
  return err
}

return tx.Commit()
```

## License

MIT

## Credits

- [Asdine El Hrychy](https://github.com/asdine)
- [Bjørn Erik Pedersen](https://github.com/bep)
