# Storm

[![Build Status](https://travis-ci.org/asdine/storm.svg)](https://travis-ci.org/asdine/storm)
[![GoDoc](https://godoc.org/github.com/asdine/storm?status.svg)](https://godoc.org/github.com/asdine/storm)

Storm is a simple and powerful toolkit for [BoltDB](https://github.com/etcd-io/bbolt) and [Badger](https://github.com/dgraph-io/badger). Basically, Storm provides indexes, a wide range of methods to store and fetch data, support for SQL, and much more. To do so, it relies on [Genji](https://github.com/genjidb/genji), a powerful embedded, document-oriented, SQL database.

In addition to the examples below, see also the [examples in the GoDoc](https://godoc.org/github.com/asdine/storm#pkg-examples).

## Getting Started

```bash
GO111MODULE=on go get -u github.com/asdine/storm/v4
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

`Open` can receive multiple options to customize the way it behaves. See [Options](#options) below.

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

`Insert` serializes the record, check for constraints, updates indexes and stores the record in the selected store.

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

// Fetching the first record
var user User
err = users.First(&user)

// Fetching records
var users []User
err = users.Find(&users)

// Deleting records
err = users.Delete()

// Updating records
err = users.Update("age", 10, "address.zipcode", "Paris")
```

These methods can be combined with clauses to filter records:

```go
// Skip records
users.Offset(10)

// Limit the number of records
users.Limit(10)

// Order records by a certain field
users.OrderBy("age")
users.OrderBy("age", "desc")

// Filter by predicate
users.Where("age", ">", 18)
users.Eq("address.city", "Paris")
users.Gt("age", 21)
users.Gte("age", 21)
users.Lt("age", 21)
users.Lte("age", 21)
```

Example:

```go
var u User
err = db.Store("user").Eq("address.city", "Paris").Limit(10).Offset(2).OrderBy("age").First(&user)

q := db.Store("user").Eq("address.city", "Paris").Limit(10).Offset(2).OrderBy("age")

var users []*User
err = q.Find(&users)
err = q.Delete()
err = q.Update("age", 10, "address.zipcode", "Paris")
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

### Options

Storm options are functions that can be passed when constructing you Storm instance. You can pass it any number of options.

#### BoltOptions

By default, Storm opens a database with the mode `0600` and a timeout of one second.
You can change this behavior by using `BoltOptions`

```go
db, err := storm.Open("my.db", storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}))
```

#### Use existing Bolt connection

You can use an existing connection and pass it to Storm

```go
bDB, _ := bolt.Open(filepath.Join(dir, "bolt.db"), 0600, &bolt.Options{Timeout: 10 * time.Second})
db := storm.Open("my.db", storm.UseDB(bDB))
```

#### Batch mode

Batch mode can be enabled to speed up concurrent writes (see [Batch read-write transactions](https://github.com/coreos/bbolt#batch-read-write-transactions))

```go
db := storm.Open("my.db", storm.Batch())
```

## BoltDB

BoltDB is still easily accessible and can be used as usual

```go
db.Bolt().View(func(tx *bolt.Tx) error {
  bucket := tx.Bucket([]byte("my bucket"))
  val := bucket.Get([]byte("any id"))
  fmt.Println(string(val))
  return nil
})
```

A transaction can be also be passed to Storm

```go
db.Bolt.Update(func(tx *bolt.Tx) error {
  ...
  dbx := db.WithTransaction(tx)
  err = dbx.Save(&user)
  ...
  return nil
})
```

## License

MIT

## Credits

- [Asdine El Hrychy](https://github.com/asdine)
- [Bjørn Erik Pedersen](https://github.com/bep)
