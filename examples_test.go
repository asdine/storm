package storm_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/asdine/storm"
)

func ExampleDB_Save() {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)

	type User struct {
		ID        int    `storm:"id"`
		Group     string `storm:"index"`
		Email     string `storm:"unique"`
		Name      string
		Age       int       `storm:"index"`
		CreatedAt time.Time `storm:"index"`
	}

	// Open takes an optional list of options as the last argument.
	// AutoIncrement will auto-increment integer IDs without existing values.
	db, _ := storm.Open(filepath.Join(dir, "storm.db"), storm.AutoIncrement())
	defer db.Close()

	user := User{
		Group:     "staff",
		Email:     "john@provider.com",
		Name:      "John",
		Age:       21,
		CreatedAt: time.Now(),
	}

	err := db.Save(&user)
	fmt.Println(err)

	user2 := user
	user2.ID = 0

	// Save will fail because of the unique constraint on Email
	err = db.Save(&user2)
	fmt.Println(err)

	// Output:
	// <nil>
	// already exists
}

func ExampleDB_One() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var user User

	err := db.One("Email", "john@provider.com", &user)
	fmt.Println(err)

	// One only works for indexed fields.
	err = db.One("Name", "John", &user)
	fmt.Println(err)

	// Output:
	// <nil>
	// not found
}

func ExampleDB_Find() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.Find("Group", "staff", &users)

	fmt.Println(err)
	fmt.Println("Found", len(users))

	// Output:
	// <nil>
	// Found 3
}

func ExampleDB_All() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.All(&users)

	fmt.Println(err)
	fmt.Println("Found", len(users))

	// Output:
	// <nil>
	// Found 3
}

func ExampleDB_AllByIndex() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.AllByIndex("CreatedAt", &users)

	fmt.Println(err)
	fmt.Println("Found", len(users))

	// Output:
	// <nil>
	// Found 3
}

func ExampleDB_Range() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.Range("Age", 21, 22, &users)

	fmt.Println(err)
	fmt.Println("Found", len(users))

	// Output:
	// <nil>
	// Found 2
}

func ExampleLimit() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.All(&users, storm.Limit(2))

	fmt.Println(err)
	fmt.Println("Found", len(users))

	// Output:
	// <nil>
	// Found 2
}

func ExampleSkip() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.All(&users, storm.Skip(1))

	fmt.Println(err)
	fmt.Println("Found", len(users))

	// Output:
	// <nil>
	// Found 2
}

func ExampleDB_Remove() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var user User

	err := db.One("ID", 1, &user)
	fmt.Println(err)
	err = db.Remove(user)
	fmt.Println(err)

	// Output:
	// <nil>
	// <nil>
}

type User struct {
	ID        int    `storm:"id"`
	Group     string `storm:"index"`
	Email     string `storm:"unique"`
	Name      string
	Age       int       `storm:"index"`
	CreatedAt time.Time `storm:"index"`
}

func prepareDB() (string, *storm.DB) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	db, _ := storm.Open(filepath.Join(dir, "storm.db"), storm.AutoIncrement())

	for i, name := range []string{"John", "Eric", "Dilbert"} {
		email := strings.ToLower(name + "@provider.com")
		user := User{
			Group:     "staff",
			Email:     email,
			Name:      name,
			Age:       21 + i,
			CreatedAt: time.Now(),
		}
		err := db.Save(&user)

		if err != nil {
			panic(err)
		}
	}

	return dir, db

}
