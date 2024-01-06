# Rainstorm

[![GoDoc](https://godoc.org/github.com/AndersonBargas/rainstorm?status.svg)](https://pkg.go.dev/github.com/AndersonBargas/rainstorm/v5?tab=doc) [![Go Report Card](https://goreportcard.com/badge/github.com/AndersonBargas/rainstorm)](https://goreportcard.com/report/github.com/AndersonBargas/rainstorm) ![GoLang](https://github.com/AndersonBargas/rainstorm/workflows/GoLang/badge.svg)

Rainstorm is a simple and powerful toolkit for [BoltDB](https://github.com/coreos/bbolt), forked from the great [Storm](https://github.com/asdine/storm).
Basically, Rainstorm provides indexes, a wide range of methods to store and fetch data, an advanced query system, and much more.

In addition to the examples below, see also the [examples in the GoDoc](https://godoc.org/github.com/AndersonBargas/rainstorm#pkg-examples).

## Table of Contents

- [Getting Started](#getting-started)
- [Main differences from "storm"](#main-differences-from-"storm")
- [Import Rainstorm](#import-rainstorm)
- [Open a database](#open-a-database)
- [Simple CRUD system](#simple-crud-system)
  - [Declare your structures](#declare-your-structures)
  - [Save your object](#save-your-object)
    - [Auto Increment](#auto-increment)
  - [Simple queries](#simple-queries)
    - [Fetch one object](#fetch-one-object)
    - [Fetch multiple objects](#fetch-multiple-objects)
    - [Fetch all objects](#fetch-all-objects)
    - [Fetch all objects sorted by index](#fetch-all-objects-sorted-by-index)
    - [Fetch a range of objects](#fetch-a-range-of-objects)
    - [Fetch objects by prefix](#fetch-objects-by-prefix)
    - [Skip, Limit and Reverse](#skip-limit-and-reverse)
    - [Delete an object](#delete-an-object)
    - [Update an object](#update-an-object)
    - [Count the objects](#count-the-objects)
    - [Initialize buckets and indexes before saving an object](#initialize-buckets-and-indexes-before-saving-an-object)
    - [Drop a bucket](#drop-a-bucket)
    - [Re-index a bucket](#re-index-a-bucket)
  - [Advanced queries](#advanced-queries)
  - [Transactions](#transactions)
  - [Options](#options)
    - [BoltOptions](#boltoptions)
    - [MarshalUnmarshaler](#marshalunmarshaler)
      - [Provided Codecs](#provided-codecs)
    - [Use existing Bolt connection](#use-existing-bolt-connection)
    - [Batch mode](#batch-mode)
- [Nodes and nested buckets](#nodes-and-nested-buckets)
  - [Node options](#node-options)
- [Simple Key/Value store](#simple-keyvalue-store)
- [BoltDB](#boltdb)
- [License](#license)
- [Credits](#credits)

## Getting Started

```bash
GO111MODULE=on go get -u github.com/AndersonBargas/rainstorm/v5
```

## Main differences from "storm"

The main differente for now is the primary key indexes. On storm, the PK index (aka ID) is, in fact, a unique index.
By nature, an ID is already unique, so there's no need to indexing the ID.
This way, rainstorm uses the Primary Keys as ID so almost every operations over ID index is executed as fast as possible.

Rainstorm emerged from the fork of storm library at version 3.1.0. After renaming the package and their imports to the rainstorm, version 3.2.0 was generated.
This means that if you want to test the library with the same characteristics that it had at the time of the fork, just use version 3.2.0.
To take advantage of the performance changes made after the fork, just use version 4 and above.

## Import Rainstorm

```go
import "github.com/AndersonBargas/rainstorm/v5"
```

## Open a database

Quick way of opening a database

```go
db, err := rainstorm.Open("my.db")

defer db.Close()
```

`Open` can receive multiple options to customize the way it behaves. See [Options](#options) below

## Simple CRUD system

### Declare your structures

```go
type User struct {
  ID int // primary key
  Group string `rainstorm:"index"` // this field will be indexed
  Email string `rainstorm:"unique"` // this field will be indexed with a unique constraint
  Name string // this field will not be indexed
  Age int `rainstorm:"index"`
}
```

The primary key can be of any type as long as it is not a zero value. Rainstorm will search for the tag `id`, if not present Rainstorm will search for a field named `ID`.

```go
type User struct {
  ThePrimaryKey string `rainstorm:"id"`// primary key
  Group string `rainstorm:"index"` // this field will be indexed
  Email string `rainstorm:"unique"` // this field will be indexed with a unique constraint
  Name string // this field will not be indexed
}
```

Rainstorm handles tags in nested structures with the `inline` tag

```go
type Base struct {
  Ident bson.ObjectId `rainstorm:"id"`
}

type User struct {
  Base      `rainstorm:"inline"`
  Group     string `rainstorm:"index"`
  Email     string `rainstorm:"unique"`
  Name      string
  CreatedAt time.Time `rainstorm:"index"`
}
```

### Save your object

```go
user := User{
  ID: 10,
  Group: "staff",
  Email: "john@provider.com",
  Name: "John",
  Age: 21,
  CreatedAt: time.Now(),
}

err := db.Save(&user)
// err == nil

user.ID++
err = db.Save(&user)
// err == rainstorm.ErrAlreadyExists
```

That's it.

`Save` creates or updates all the required indexes and buckets, checks the unique constraints and saves the object to the store.

#### Auto Increment

Rainstorm can auto increment integer values so you don't have to worry about that when saving your objects. Also, the new value is automatically inserted in your field.

```go

type Product struct {
  Pk                  int `rainstorm:"id,increment"` // primary key with auto increment
  Name                string
  IntegerField        uint64 `rainstorm:"increment"`
  IndexedIntegerField uint32 `rainstorm:"index,increment"`
  UniqueIntegerField  int16  `rainstorm:"unique,increment=100"` // the starting value can be set
}

p := Product{Name: "Vaccum Cleaner"}

fmt.Println(p.Pk)
fmt.Println(p.IntegerField)
fmt.Println(p.IndexedIntegerField)
fmt.Println(p.UniqueIntegerField)
// 0
// 0
// 0
// 0

_ = db.Save(&p)

fmt.Println(p.Pk)
fmt.Println(p.IntegerField)
fmt.Println(p.IndexedIntegerField)
fmt.Println(p.UniqueIntegerField)
// 1
// 1
// 1
// 100

```

### Simple queries

Any object can be fetched, indexed or not. Rainstorm uses indexes when available, otherwise it uses the [query system](#advanced-queries).

#### Fetch one object

```go
var user User
err := db.One("Email", "john@provider.com", &user)
// err == nil

err = db.One("Name", "John", &user)
// err == nil

err = db.One("Name", "Jack", &user)
// err == rainstorm.ErrNotFound
```

#### Fetch multiple objects

```go
var users []User
err := db.Find("Group", "staff", &users)
```

#### Fetch all objects

```go
var users []User
err := db.All(&users)
```

#### Fetch all objects sorted by index

```go
var users []User
err := db.AllByIndex("CreatedAt", &users)
```

#### Fetch a range of objects

```go
var users []User
err := db.Range("Age", 10, 21, &users)
```

#### Fetch objects by prefix

```go
var users []User
err := db.Prefix("Name", "Jo", &users)
```

#### Skip, Limit and Reverse

```go
var users []User
err := db.Find("Group", "staff", &users, rainstorm.Skip(10))
err = db.Find("Group", "staff", &users, rainstorm.Limit(10))
err = db.Find("Group", "staff", &users, rainstorm.Reverse())
err = db.Find("Group", "staff", &users, rainstorm.Limit(10), rainstorm.Skip(10), rainstorm.Reverse())

err = db.All(&users, rainstorm.Limit(10), rainstorm.Skip(10), rainstorm.Reverse())
err = db.AllByIndex("CreatedAt", &users, rainstorm.Limit(10), rainstorm.Skip(10), rainstorm.Reverse())
err = db.Range("Age", 10, 21, &users, rainstorm.Limit(10), rainstorm.Skip(10), rainstorm.Reverse())
```

#### Delete an object

```go
err := db.DeleteStruct(&User{})
```

#### Update an object

```go
// Update multiple fields
// Only works for non zero-value fields (e.g. Name can not be "", Age can not be 0)
err := db.Update(&User{ID: 10, Name: "Jack", Age: 45})

// Update a single field
// Also works for zero-value fields (0, false, "", ...)
err := db.UpdateField(&User{ID: 10}, "Age", 0)
```

#### Count the objects
```go
number, err := db.Count(&User{})
```

#### Initialize buckets and indexes before saving an object

```go
err := db.Init(&User{})
```

Useful when starting your application

#### Drop a bucket

Using the struct

```go
err := db.Drop(&User)
```

Using the bucket name

```go
err := db.Drop("User")
```

#### Re-index a bucket

```go
err := db.ReIndex(&User{})
```

Useful when the structure has changed

### Advanced queries

For more complex queries, you can use the `Select` method.
`Select` takes any number of [`Matcher`](https://godoc.org/github.com/AndersonBargas/rainstorm/q#Matcher) from the [`q`](https://godoc.org/github.com/AndersonBargas/rainstorm/q) package.

Here are some common Matchers:

```go
// Equality
q.Eq("Name", John)

// Strictly greater than
q.Gt("Age", 7)

// Lesser than or equal to
q.Lte("Age", 77)

// Regex with name that starts with the letter D
q.Re("Name", "^D")

// In the given slice of values
q.In("Group", []string{"Staff", "Admin"})

// Comparing fields
q.EqF("FieldName", "SecondFieldName")
q.LtF("FieldName", "SecondFieldName")
q.GtF("FieldName", "SecondFieldName")
q.LteF("FieldName", "SecondFieldName")
q.GteF("FieldName", "SecondFieldName")
```

Matchers can also be combined with `And`, `Or` and `Not`:

```go

// Match if all match
q.And(
  q.Gt("Age", 7),
  q.Re("Name", "^D")
)

// Match if one matches
q.Or(
  q.Re("Name", "^A"),
  q.Not(
    q.Re("Name", "^B")
  ),
  q.Re("Name", "^C"),
  q.In("Group", []string{"Staff", "Admin"}),
  q.And(
    q.StrictEq("Password", []byte(password)),
    q.Eq("Registered", true)
  )
)
```

You can find the complete list in the [documentation](https://godoc.org/github.com/AndersonBargas/rainstorm/q#Matcher).

`Select` takes any number of matchers and wraps them into a `q.And()` so it's not necessary to specify it. It returns a [`Query`](https://godoc.org/github.com/AndersonBargas/rainstorm#Query) type.

```go
query := db.Select(q.Gte("Age", 7), q.Lte("Age", 77))
```

The `Query` type contains methods to filter and order the records.

```go
// Limit
query = query.Limit(10)

// Skip
query = query.Skip(20)

// Calls can also be chained
query = query.Limit(10).Skip(20).OrderBy("Age").Reverse()
```

But also to specify how to fetch them.

```go
var users []User
err = query.Find(&users)

var user User
err = query.First(&user)
```

Examples with `Select`:

```go
// Find all users with an ID between 10 and 100
err = db.Select(q.Gte("ID", 10), q.Lte("ID", 100)).Find(&users)

// Nested matchers
err = db.Select(q.Or(
  q.Gt("ID", 50),
  q.Lt("Age", 21),
  q.And(
    q.Eq("Group", "admin"),
    q.Gte("Age", 21),
  ),
)).Find(&users)

query := db.Select(q.Gte("ID", 10), q.Lte("ID", 100)).Limit(10).Skip(5).Reverse().OrderBy("Age", "Name")

// Find multiple records
err = query.Find(&users)
// or
err = db.Select(q.Gte("ID", 10), q.Lte("ID", 100)).Limit(10).Skip(5).Reverse().OrderBy("Age", "Name").Find(&users)

// Find first record
err = query.First(&user)
// or
err = db.Select(q.Gte("ID", 10), q.Lte("ID", 100)).Limit(10).Skip(5).Reverse().OrderBy("Age", "Name").First(&user)

// Delete all matching records
err = query.Delete(new(User))

// Fetching records one by one (useful when the bucket contains a lot of records)
query = db.Select(q.Gte("ID", 10),q.Lte("ID", 100)).OrderBy("Age", "Name")

err = query.Each(new(User), func(record interface{}) error {
  u := record.(*User)
  ...
  return nil
})
```

See the [documentation](https://godoc.org/github.com/AndersonBargas/rainstorm#Query) for a complete list of methods.

### Transactions

```go
tx, err := db.Begin(true)
if err != nil {
  return err
}
defer tx.Rollback()

accountA.Amount -= 100
accountB.Amount += 100

err = tx.Save(accountA)
if err != nil {
  return err
}

err = tx.Save(accountB)
if err != nil {
  return err
}

return tx.Commit()
```

### Options

Rainstorm options are functions that can be passed when constructing you Rainstorm instance. You can pass it any number of options.

#### BoltOptions

By default, Rainstorm opens a database with the mode `0600` and a timeout of one second.
You can change this behavior by using `BoltOptions`

```go
db, err := rainstorm.Open("my.db", rainstorm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}))
```

#### MarshalUnmarshaler

To store the data in BoltDB, Rainstorm marshals it in JSON by default. If you wish to change this behavior you can pass a codec that implements [`codec.MarshalUnmarshaler`](https://godoc.org/github.com/AndersonBargas/rainstorm/codec#MarshalUnmarshaler) via the [`rainstorm.Codec`](https://godoc.org/github.com/AndersonBargas/rainstorm#Codec) option:

```go
db := rainstorm.Open("my.db", rainstorm.Codec(myCodec))
```

##### Provided Codecs

You can easily implement your own `MarshalUnmarshaler`, but Rainstorm comes with built-in support for [JSON](https://godoc.org/github.com/AndersonBargas/rainstorm/codec/json) (default), [GOB](https://godoc.org/github.com/AndersonBargas/rainstorm/codec/gob),  [Sereal](https://godoc.org/github.com/AndersonBargas/rainstorm/codec/sereal), [Protocol Buffers](https://godoc.org/github.com/AndersonBargas/rainstorm/codec/protobuf) and [MessagePack](https://godoc.org/github.com/AndersonBargas/rainstorm/codec/msgpack).

These can be used by importing the relevant package and use that codec to configure Rainstorm. The example below shows all variants (without proper error handling):

```go
import (
  "github.com/AndersonBargas/rainstorm/v5"
  "github.com/AndersonBargas/rainstorm/v5/codec/gob"
  "github.com/AndersonBargas/rainstorm/v5/codec/json"
  "github.com/AndersonBargas/rainstorm/v5/codec/sereal"
  "github.com/AndersonBargas/rainstorm/v5/codec/protobuf"
  "github.com/AndersonBargas/rainstorm/v5/codec/msgpack"
)

var gobDb, _ = rainstorm.Open("gob.db", rainstorm.Codec(gob.Codec))
var jsonDb, _ = rainstorm.Open("json.db", rainstorm.Codec(json.Codec))
var serealDb, _ = rainstorm.Open("sereal.db", rainstorm.Codec(sereal.Codec))
var protobufDb, _ = rainstorm.Open("protobuf.db", rainstorm.Codec(protobuf.Codec))
var msgpackDb, _ = rainstorm.Open("msgpack.db", rainstorm.Codec(msgpack.Codec))
```

**Tip**: Adding Rainstorm tags to generated Protobuf files can be tricky. A good solution is to use [this tool](https://github.com/favadi/protoc-go-inject-tag) to inject the tags during the compilation.

#### Use existing Bolt connection

You can use an existing connection and pass it to Rainstorm

```go
bDB, _ := bolt.Open(filepath.Join(dir, "bolt.db"), 0600, &bolt.Options{Timeout: 10 * time.Second})
db := rainstorm.Open("my.db", rainstorm.UseDB(bDB))
```

#### Batch mode

Batch mode can be enabled to speed up concurrent writes (see [Batch read-write transactions](https://github.com/coreos/bbolt#batch-read-write-transactions))

```go
db := rainstorm.Open("my.db", rainstorm.Batch())
```

## Nodes and nested buckets

Rainstorm takes advantage of BoltDB nested buckets feature by using `rainstorm.Node`.
A `rainstorm.Node` is the underlying object used by `rainstorm.DB` to manipulate a bucket.
To create a nested bucket and use the same API as `rainstorm.DB`, you can use the `DB.From` method.

```go
repo := db.From("repo")

err := repo.Save(&Issue{
  Title: "I want more features",
  Author: user.ID,
})

err = repo.Save(newRelease("0.10"))

var issues []Issue
err = repo.Find("Author", user.ID, &issues)

var release Release
err = repo.One("Tag", "0.10", &release)
```

You can also chain the nodes to create a hierarchy

```go
chars := db.From("characters")
heroes := chars.From("heroes")
enemies := chars.From("enemies")

items := db.From("items")
potions := items.From("consumables").From("medicine").From("potions")
```

You can even pass the entire hierarchy as arguments to `From`:

```go
privateNotes := db.From("notes", "private")
workNotes :=  db.From("notes", "work")
```

### Node options

A Node can also be configured. Activating an option on a Node creates a copy, so a Node is always thread-safe.

```go
n := db.From("my-node")
```

Give a bolt.Tx transaction to the Node

```go
n = n.WithTransaction(tx)
```

Enable batch mode

```go
n = n.WithBatch(true)
```

Use a Codec

```go
n = n.WithCodec(gob.Codec)
```

## Simple Key/Value store

Rainstorm can be used as a simple, robust, key/value store that can store anything.
The key and the value can be of any type as long as the key is not a zero value.

Saving data :

```go
db.Set("logs", time.Now(), "I'm eating my breakfast man")
db.Set("sessions", bson.NewObjectId(), &someUser)
db.Set("weird storage", "754-3010", map[string]interface{}{
  "hair": "blonde",
  "likes": []string{"cheese", "star wars"},
})
```

Fetching data :

```go
user := User{}
db.Get("sessions", someObjectId, &user)

var details map[string]interface{}
db.Get("weird storage", "754-3010", &details)

db.Get("sessions", someObjectId, &details)
```

Deleting data :

```go
db.Delete("sessions", someObjectId)
db.Delete("weird storage", "754-3010")
```

You can find other useful methods in the [documentation](https://godoc.org/github.com/AndersonBargas/rainstorm#KeyValueStore).

## BoltDB

BoltDB is still easily accessible and can be used as usual

```go
db.Bolt.View(func(tx *bolt.Tx) error {
  bucket := tx.Bucket([]byte("my bucket"))
  val := bucket.Get([]byte("any id"))
  fmt.Println(string(val))
  return nil
})
```

A transaction can be also be passed to Rainstorm

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
