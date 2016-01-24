# Storm

Storm is a wrapper and simple ORM around BoltDB. The goal of this project is to provide a simple way to save any object in BoltDB and to easily retrieve it.

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

By default, Storm open a database with the mode `0600` and a timeout of one second.
You can change this behavior by using `OpenWithOptions`
```go
db, err := storm.OpenWithOptions("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
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
```

## Simple ORM

### Declare your structures

```go
type User struct {
  ID int // primary key
  Group string `storm:"index"` // this field will be indexed
  Email string `storm:"unique"` // this field will be indexed and marked as unique
  Name string // this field will not be indexed
}
```

The primary key can be of any type as long as it is not a zero value. Storm will search for the tag `id`, if not present storm will search for a field named `ID`.

```go
type User struct {
  ThePrimaryKey string `storm:"id"`// primary key
  Group string `storm:"index"` // this field will be indexed
  Email string `storm:"unique"` // this field will be indexed with a unique constraint
  Name string // this field will not be indexed
}
```

Storm handles tags in nested structures

```go
type Base struct {
  Ident bson.ObjectId `storm:"id"`
}

type User struct {
  Base `storm:"inline"`
  Group string `storm:"index"`
  Email string `storm:"unique"`
  Name string
}
```

### Save your object

```go
user := User{
  ID: 10,
  Group: "staff",
  Email: "john@provider.com",
  Name: "John",
}

err := db.Save(&user)
// err == nil

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

### Remove an object
```go
err := db.Remove(&user)
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

## TODO

- Improve documentation / comments
- Btree indexes and more
- Better errors
- Order
- Limit
- Offset
- Create time tags
- Update time tags

## License

MIT

## Author

**Asdine El Hrychy**

- [Twitter](https://twitter.com/asdine_)
- [Github](https://github.com/asdine)
