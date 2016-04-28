# Storm

[![Build Status](https://travis-ci.org/asdine/storm.svg)](https://travis-ci.org/asdine/storm)
[![GoDoc](https://godoc.org/github.com/asdine/storm?status.svg)](https://godoc.org/github.com/asdine/storm)
[![Go Report Card](https://goreportcard.com/badge/github.com/asdine/storm)](https://goreportcard.com/report/github.com/asdine/storm)
[![Coverage](http://gocover.io/_badge/github.com/asdine/storm)](http://gocover.io/github.com/asdine/storm)

Storm is simple and powerful ORM for [BoltDB](https://github.com/boltdb/bolt). The goal of this project is to provide a simple way to save any object in BoltDB and to easily retrieve it.

## Getting Started

```bash
go get -u github.com/asdine/storm
```

## Import Storm

```go
import "github.com/asdine/storm"
```

## Open a database

Quick way of opening a database
```go
db, err := storm.Open("my.db")

defer db.Close()
```

`Open` can receive multiple options to customize the way it behaves. See [Options](#options) below

## Simple ORM

### Declare your structures

```go
type User struct {
  ID int // primary key
  Group string `storm:"index"` // this field will be indexed
  Email string `storm:"unique"` // this field will be indexed with a unique constraint
  Name string // this field will not be indexed
  Age int `storm:"index"`
}
```

The primary key can be of any type as long as it is not a zero value. Storm will search for the tag `id`, if not present Storm will search for a field named `ID`.

```go
type User struct {
  ThePrimaryKey string `storm:"id"`// primary key
  Group string `storm:"index"` // this field will be indexed
  Email string `storm:"unique"` // this field will be indexed with a unique constraint
  Name string // this field will not be indexed
}
```

Storm handles tags in nested structures with the `inline` tag

```go
type Base struct {
  Ident bson.ObjectId `storm:"id"`
}

type User struct {
	Base      `storm:"inline"`
	Group     string `storm:"index"`
	Email     string `storm:"unique"`
	Name      string
	CreatedAt time.Time `storm:"index"`
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
// err == "already exists"
```

That's it.

`Save` creates or updates all the required indexes and buckets, checks the unique constraints and saves the object to the store.

### Fetch your object

Only indexed fields can be used to find a record

```go
var user User
err := db.One("Email", "john@provider.com", &user)
// err == nil

err = db.One("Name", "John", &user)
// err == "not found"
```

### Fetch multiple objects

```go
var users []User
err := db.Find("Group", "staff", &users)
```

### Fetch all objects

```go
var users []User
err := db.All(&users)
```

### Fetch all objects sorted by index

```go
var users []User
err := db.AllByIndex("CreatedAt", &users)
```

### Fetch a range of objects

```go
var users []User
err := db.Range("Age", 10, 21, &users)
```

### Skip and Limit

```go
var users []User
err := db.Find("Group", "staff", &users, storm.Skip(10))
err = db.Find("Group", "staff", &users, storm.Limit(10))
err = db.Find("Group", "staff", &users, storm.Limit(10), storm.Skip(10))

err = db.All(&users, storm.Limit(10), storm.Skip(10))
err = db.AllByIndex("CreatedAt", &users, storm.Limit(10), storm.Skip(10))
err = db.Range("Age", 10, 21, &users, storm.Limit(10), storm.Skip(10))
```

### Remove an object

```go
err := db.Remove(&user)
```

### Initialize buckets and indexes before saving an object

```go
err := db.Init(&User{})
```

Useful when starting your application

### Transactions

```go
tx, err := db.Begin(true)

accountA.Amount -= 100
accountB.Amount += 100

err = tx.Save(accountA)
if err != nil {
  tx.Rollback()
  return err
}

err = tx.Save(accountB)
if err != nil {
  tx.Rollback()
  return err
}

tx.Commit()
```
### Options

Storm options are functions that can be passed when constructing you Storm instance. You can pass it any number of options.

#### BoltOptions

By default, Storm opens a database with the mode `0600` and a timeout of one second.
You can change this behavior by using `BoltOptions`

```go
db, err := storm.Open("my.db", storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}))
```

#### EncodeDecoder

To store the data in BoltDB, Storm encodes it in GOB by default. If you wish to change this behavior you can pass a codec that implements [`codec.EncodeDecoder`](https://godoc.org/github.com/asdine/storm/codec#EncodeDecoder) via the [`storm.Codec`](https://godoc.org/github.com/asdine/storm#Codec) option:

```go
db := storm.Open("my.db", storm.Codec(myCodec))
```

##### Provided Codecs

You can easily implement your own `EncodeDecoder`, but Storm comes with built-in support for [GOB](https://godoc.org/github.com/asdine/storm/codec/gob) (default), [JSON](https://godoc.org/github.com/asdine/storm/codec/json), and [Sereal](https://godoc.org/github.com/asdine/storm/codec/sereal).

These can be used by importing the relevant package and use that codec to configure Storm. The example below shows all three (without proper error handling):

```go
import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/codec/json"
	"github.com/asdine/storm/codec/sereal"
)

var gobDb, _ = storm.Open("gob.db", storm.Codec(gob.Codec))
var jsonDb, _ = storm.Open("json.db", storm.Codec(json.Codec))
var serealDb, _ = storm.Open("sereal.db", storm.Codec(sereal.Codec))
```

#### Auto Increment

Storm can auto increment integer IDs so you don't have to worry about that when saving your objects.

```go
db := storm.Open("my.db", storm.AutoIncrement())
```

## Nodes and nested buckets

Storm takes advantage of BoltDB nested buckets feature by using `storm.Node`.
A `storm.Node` is the underlying object used by `storm.DB` to manipulate a bucket.
To create a nested bucket and use the same API as `storm.DB`, you can use the `DB.From` method.

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

## Simple Key/Value store

Storm can be used as a simple, robust, key/value store that can store anything.
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

## TODO

- Search
- Reverse order
- More indexes

## License

MIT

## Author

**Asdine El Hrychy**

- [Twitter](https://twitter.com/asdine_)
- [Github](https://github.com/asdine)
