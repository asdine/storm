package storm

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/codec/json"
	"github.com/asdine/storm/q"
	"github.com/coreos/bbolt"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	var u IndexedNameUser
	err := db.One("Name", "John", &u)
	require.Equal(t, ErrNotFound, err)

	err = db.Init(&u)
	require.NoError(t, err)

	err = db.One("Name", "John", &u)
	require.Error(t, err)
	require.Equal(t, ErrNotFound, err)

	err = db.Init(&ClassicBadTags{})
	require.Error(t, err)
	require.Equal(t, ErrUnknownTag, err)

	err = db.Init(10)
	require.Error(t, err)
	require.Equal(t, ErrBadType, err)

	err = db.Init(&ClassicNoTags{})
	require.Error(t, err)
	require.Equal(t, ErrNoID, err)

	err = db.Init(&struct{ ID string }{})
	require.Error(t, err)
	require.Equal(t, ErrNoName, err)
}

func TestInitMetadata(t *testing.T) {
	db, cleanup := createDB(t, Batch())
	defer cleanup()

	err := db.Init(new(User))
	require.NoError(t, err)
	n := db.WithCodec(gob.Codec)
	err = n.Init(new(User))
	require.Equal(t, ErrDifferentCodec, err)
}

func TestReIndex(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	for i := 1; i < 10; i++ {
		type User struct {
			ID   int
			Age  int    `storm:"index"`
			Name string `storm:"unique"`
		}

		u := User{
			ID:   i,
			Age:  i % 2,
			Name: fmt.Sprintf("John%d", i),
		}
		err := db.Save(&u)
		require.NoError(t, err)
	}

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("User"))
		require.NotNil(t, bucket)

		require.NotNil(t, bucket.Bucket([]byte(indexPrefix+"Name")))
		require.NotNil(t, bucket.Bucket([]byte(indexPrefix+"Age")))
		return nil
	})

	type User struct {
		ID    int
		Age   int
		Name  string `storm:"index"`
		Group string `storm:"unique"`
	}

	require.NoError(t, db.ReIndex(new(User)))

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("User"))
		require.NotNil(t, bucket)

		require.NotNil(t, bucket.Bucket([]byte(indexPrefix+"Name")))
		require.Nil(t, bucket.Bucket([]byte(indexPrefix+"Age")))
		require.NotNil(t, bucket.Bucket([]byte(indexPrefix+"Group")))
		return nil
	})
}

func TestSave(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	err := db.Save(&SimpleUser{ID: 10, Name: "John"})
	require.NoError(t, err)

	err = db.Save(&SimpleUser{Name: "John"})
	require.Error(t, err)
	require.Equal(t, ErrZeroID, err)

	err = db.Save(&ClassicBadTags{ID: "id", PublicField: 100})
	require.Error(t, err)
	require.Equal(t, ErrUnknownTag, err)

	err = db.Save(&UserWithNoID{Name: "John"})
	require.Error(t, err)
	require.Equal(t, ErrNoID, err)

	err = db.Save(&UserWithIDField{ID: 10, Name: "John"})
	require.NoError(t, err)

	u := UserWithEmbeddedIDField{}
	u.ID = 150
	u.Name = "Pete"
	u.Age = 10
	err = db.Save(&u)
	require.NoError(t, err)

	v := UserWithIDField{ID: 10, Name: "John"}
	err = db.Save(&v)
	require.NoError(t, err)

	w := UserWithEmbeddedField{}
	w.ID = 150
	w.Name = "John"
	err = db.Save(&w)
	require.NoError(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("UserWithIDField"))
		require.NotNil(t, bucket)

		i, err := toBytes(10, json.Codec)
		require.NoError(t, err)

		val := bucket.Get(i)
		require.NotNil(t, val)

		content, err := db.Codec().Marshal(&v)
		require.NoError(t, err)
		require.Equal(t, content, val)
		return nil
	})
}

func TestSaveUnique(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	u1 := UniqueNameUser{ID: 10, Name: "John", Age: 10}
	err := db.Save(&u1)
	require.NoError(t, err)

	u2 := UniqueNameUser{ID: 11, Name: "John", Age: 100}
	err = db.Save(&u2)
	require.Error(t, err)
	require.True(t, ErrAlreadyExists == err)

	// same id
	u3 := UniqueNameUser{ID: 10, Name: "Jake", Age: 100}
	err = db.Save(&u3)
	require.NoError(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("UniqueNameUser"))

		uniqueBucket := bucket.Bucket([]byte(indexPrefix + "Name"))
		require.NotNil(t, uniqueBucket)

		id := uniqueBucket.Get([]byte("Jake"))
		i, err := toBytes(10, json.Codec)
		require.NoError(t, err)
		require.Equal(t, i, id)

		id = uniqueBucket.Get([]byte("John"))
		require.Nil(t, id)
		return nil
	})
}

func TestSaveUniqueStruct(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	a := ClassicUnique{ID: "id1"}
	a.InlineStruct.A = 10.0
	a.InlineStruct.B = 12.0

	err := db.Save(&a)
	require.NoError(t, err)

	b := ClassicUnique{ID: "id2"}
	b.InlineStruct.A = 10.0
	b.InlineStruct.B = 12.0

	err = db.Save(&b)
	require.Equal(t, ErrAlreadyExists, err)

	err = db.One("InlineStruct", struct {
		A float32
		B float64
	}{A: 10.0, B: 12.0}, &b)
	require.NoError(t, err)
	require.Equal(t, a.ID, b.ID)
}

func TestSaveIndex(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	u1 := IndexedNameUser{ID: 10, Name: "John", age: 10}
	err := db.Save(&u1)
	require.NoError(t, err)

	u1 = IndexedNameUser{ID: 10, Name: "John", age: 10}
	err = db.Save(&u1)
	require.NoError(t, err)

	u2 := IndexedNameUser{ID: 11, Name: "John", age: 100}
	err = db.Save(&u2)
	require.NoError(t, err)

	name1 := "Jake"
	name2 := "Jane"
	name3 := "James"

	for i := 0; i < 100; i++ {
		u := IndexedNameUser{ID: i + 1}

		if i%2 == 0 {
			u.Name = name1
		} else {
			u.Name = name2
		}

		db.Save(&u)
	}

	var users []IndexedNameUser
	err = db.Find("Name", name1, &users)
	require.NoError(t, err)
	require.Len(t, users, 50)

	err = db.Find("Name", name2, &users)
	require.NoError(t, err)
	require.Len(t, users, 50)

	err = db.Find("Name", name3, &users)
	require.Error(t, err)
	require.Equal(t, ErrNotFound, err)

	err = db.Save(nil)
	require.Error(t, err)
	require.Equal(t, ErrStructPtrNeeded, err)
}

func TestSaveEmptyValues(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	u := User{
		ID: 10,
	}
	err := db.Save(&u)
	require.NoError(t, err)

	var v User
	err = db.One("ID", 10, &v)
	require.NoError(t, err)
	require.Equal(t, 10, v.ID)

	u.Name = "John"
	u.Slug = "john"
	err = db.Save(&u)
	require.NoError(t, err)

	err = db.One("Name", "John", &v)
	require.NoError(t, err)
	require.Equal(t, "John", v.Name)
	require.Equal(t, "john", v.Slug)
	err = db.One("Slug", "john", &v)
	require.NoError(t, err)
	require.Equal(t, "John", v.Name)
	require.Equal(t, "john", v.Slug)

	u.Name = ""
	u.Slug = ""
	err = db.Save(&u)
	require.NoError(t, err)

	err = db.One("Name", "John", &v)
	require.Error(t, err)
	err = db.One("Slug", "john", &v)
	require.Error(t, err)
}

func TestSaveIncrement(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	type User struct {
		Identifier int    `storm:"id,increment"`
		Name       string `storm:"index,increment"`
		Age        int    `storm:"unique,increment=18"`
	}

	for i := 1; i < 10; i++ {
		s1 := User{Name: fmt.Sprintf("John%d", i)}
		err := db.Save(&s1)
		require.NoError(t, err)
		require.Equal(t, i, s1.Identifier)
		require.Equal(t, i-1+18, s1.Age)
		require.Equal(t, fmt.Sprintf("John%d", i), s1.Name)

		var s2 User
		err = db.One("Identifier", i, &s2)
		require.NoError(t, err)
		require.Equal(t, s1, s2)

		var list []User
		err = db.Find("Age", i-1+18, &list)
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, s1, list[0])
	}
}

func TestSaveDifferentBucketRoot(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	require.Len(t, db.Node.(*node).rootBucket, 0)

	dbSub := db.From("sub").(*node)

	require.NotEqual(t, dbSub, db)
	require.Len(t, dbSub.rootBucket, 1)

	err := db.Save(&User{ID: 10, Name: "John"})
	require.NoError(t, err)
	err = dbSub.Save(&User{ID: 11, Name: "Paul"})
	require.NoError(t, err)

	var (
		john User
		paul User
	)

	err = db.One("Name", "John", &john)
	require.NoError(t, err)
	err = db.One("Name", "Paul", &paul)
	require.Error(t, err)

	err = dbSub.One("Name", "Paul", &paul)
	require.NoError(t, err)
	err = dbSub.One("Name", "John", &john)
	require.Error(t, err)
}

func TestSaveEmbedded(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	type Base struct {
		ID int `storm:"id,increment"`
	}

	type User struct {
		Base      `storm:"inline"`
		Group     string `storm:"index"`
		Email     string `storm:"unique"`
		Name      string
		Age       int
		CreatedAt time.Time `storm:"index"`
	}

	user := User{
		Group:     "staff",
		Email:     "john@provider.com",
		Name:      "John",
		Age:       21,
		CreatedAt: time.Now(),
	}

	err := db.Save(&user)
	require.NoError(t, err)
	require.Equal(t, 1, user.ID)
}

func TestSaveByValue(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	w := User{Name: "John"}
	err := db.Save(w)
	require.Error(t, err)
	require.Equal(t, ErrStructPtrNeeded, err)
}

func TestSaveWithBatch(t *testing.T) {
	db, cleanup := createDB(t, Batch())
	defer cleanup()

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := db.Save(&User{ID: i + 1, Name: "John"})
			require.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestSaveMetadata(t *testing.T) {
	db, cleanup := createDB(t, Batch())
	defer cleanup()

	w := User{ID: 10, Name: "John"}
	err := db.Save(&w)
	require.NoError(t, err)
	n := db.WithCodec(gob.Codec)
	err = n.Save(&w)
	require.Equal(t, ErrDifferentCodec, err)
}

func TestUpdate(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	type User struct {
		ID          int       `storm:"id,increment"`
		Name        string    `storm:"index"`
		Age         uint64    `storm:"index,increment"`
		DateOfBirth time.Time `storm:"index"`
		Group       string
		Slug        string `storm:"unique"`
	}

	var u User

	err := db.Save(&User{ID: 10, Name: "John", Age: 5, Group: "Staff", Slug: "john"})
	require.NoError(t, err)

	// nil
	err = db.Update(nil)
	require.Equal(t, ErrStructPtrNeeded, err)

	// no id
	err = db.Update(&User{Name: "Jack"})
	require.Equal(t, ErrNoID, err)

	// Unknown user
	err = db.Update(&User{ID: 11, Name: "Jack"})
	require.Equal(t, ErrNotFound, err)

	// actual user
	err = db.Update(&User{ID: 10, Name: "Jack"})
	require.NoError(t, err)

	err = db.One("Name", "John", &u)
	require.Equal(t, ErrNotFound, err)

	err = db.One("Name", "Jack", &u)
	require.NoError(t, err)
	require.Equal(t, "Jack", u.Name)
	require.Equal(t, uint64(5), u.Age)

	// indexed field with zero value #170
	err = db.Update(&User{ID: 10, Group: "Staff"})
	require.NoError(t, err)

	err = db.One("Name", "Jack", &u)
	require.NoError(t, err)
	require.Equal(t, "Jack", u.Name)
	require.Equal(t, uint64(5), u.Age)
	require.Equal(t, "Staff", u.Group)
}

func TestUpdateField(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	type User struct {
		ID          int       `storm:"id,increment"`
		Name        string    `storm:"index"`
		Age         uint64    `storm:"index,increment"`
		DateOfBirth time.Time `storm:"index"`
		Group       string
		Slug        string `storm:"unique"`
	}

	var u User

	err := db.Save(&User{ID: 10, Name: "John", Age: 5, Group: "Staff", Slug: "john"})
	require.NoError(t, err)

	// nil
	err = db.UpdateField(nil, "", nil)
	require.Equal(t, ErrStructPtrNeeded, err)

	// no id
	err = db.UpdateField(&User{}, "Name", "Jack")
	require.Equal(t, ErrNoID, err)

	// Unknown user
	err = db.UpdateField(&User{ID: 11}, "Name", "Jack")
	require.Equal(t, ErrNotFound, err)

	// Unknown field
	err = db.UpdateField(&User{ID: 11}, "Address", "Jack")
	require.Equal(t, ErrNotFound, err)

	// Incompatible value
	err = db.UpdateField(&User{ID: 10}, "Name", 50)
	require.Equal(t, ErrIncompatibleValue, err)

	// actual user
	err = db.UpdateField(&User{ID: 10}, "Name", "Jack")
	require.NoError(t, err)

	err = db.One("Name", "John", &u)
	require.Equal(t, ErrNotFound, err)

	err = db.One("Name", "Jack", &u)
	require.NoError(t, err)
	require.Equal(t, "Jack", u.Name)

	// zero value
	err = db.UpdateField(&User{ID: 10}, "Name", "")
	require.NoError(t, err)

	err = db.One("Name", "Jack", &u)
	require.Equal(t, ErrNotFound, err)

	err = db.One("ID", 10, &u)
	require.NoError(t, err)
	require.Equal(t, "", u.Name)

	// zero value with int and increment
	err = db.UpdateField(&User{ID: 10}, "Age", uint64(0))
	require.NoError(t, err)

	err = db.Select(q.Eq("Age", uint64(5))).First(&u)
	require.Equal(t, ErrNotFound, err)

	err = db.Select(q.Eq("Age", uint64(0))).First(&u)
	require.NoError(t, err)
}

func TestDropByString(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	n := db.From("b1", "b2", "b3")
	err := n.Save(&SimpleUser{ID: 10, Name: "John"})
	require.NoError(t, err)

	err = db.From("b1").Drop("b2")
	require.NoError(t, err)

	err = db.From("b1").Drop("b2")
	require.Error(t, err)

	n.From("b4").Drop("b5")
	require.Error(t, err)

	err = db.Drop("b1")
	require.NoError(t, err)

	db.Bolt.Update(func(tx *bolt.Tx) error {
		require.Nil(t, db.From().GetBucket(tx, "b1"))
		d := db.WithTransaction(tx)
		n := d.From("a1")
		err = n.Save(&SimpleUser{ID: 10, Name: "John"})
		require.NoError(t, err)

		err = d.Drop("a1")
		require.NoError(t, err)

		return nil
	})
}

func TestDropByStruct(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	n := db.From("b1", "b2", "b3")
	err := n.Save(&SimpleUser{ID: 10, Name: "John"})
	require.NoError(t, err)

	err = n.Drop(&SimpleUser{})
	require.NoError(t, err)

	db.Bolt.Update(func(tx *bolt.Tx) error {
		require.Nil(t, n.GetBucket(tx, "SimpleUser"))
		d := db.WithTransaction(tx)
		n := d.From("a1")
		err = n.Save(&SimpleUser{ID: 10, Name: "John"})
		require.NoError(t, err)

		err = n.Drop(&SimpleUser{})
		require.NoError(t, err)

		require.Nil(t, n.GetBucket(tx, "SimpleUser"))
		return nil
	})
}

func TestDeleteStruct(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	u1 := IndexedNameUser{ID: 10, Name: "John", age: 10}
	err := db.Save(&u1)
	require.NoError(t, err)

	err = db.DeleteStruct(u1)
	require.Equal(t, ErrStructPtrNeeded, err)

	err = db.DeleteStruct(&u1)
	require.NoError(t, err)

	err = db.DeleteStruct(&u1)
	require.Equal(t, ErrNotFound, err)

	u2 := IndexedNameUser{}
	err = db.Get("IndexedNameUser", 10, &u2)
	require.True(t, ErrNotFound == err)

	err = db.DeleteStruct(nil)
	require.Equal(t, ErrStructPtrNeeded, err)

	var users []User
	for i := 0; i < 10; i++ {
		user := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1), DateOfBirth: time.Now().Add(-time.Duration(i*10) * time.Minute)}
		err = db.Save(&user)
		require.NoError(t, err)
		users = append(users, user)
	}

	err = db.DeleteStruct(&users[0])
	require.NoError(t, err)
	err = db.DeleteStruct(&users[1])
	require.NoError(t, err)

	users = nil
	err = db.All(&users)
	require.NoError(t, err)
	require.Len(t, users, 8)
	require.Equal(t, 3, users[0].ID)
}
