package storm_test

import (
	"fmt"
	"io/ioutil"
	"log"
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

	if err != nil {
		log.Fatal(err)
	}

	user2 := user
	user2.ID = 0

	// Save will fail because of the unique constraint on Email
	err = db.Save(&user2)
	fmt.Println(err)

	// Output:
	// already exists
}

func ExampleDB_One() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var user User

	err := db.One("Email", "john@provider.com", &user)

	if err != nil {
		log.Fatal(err)
	}

	// One only works for indexed fields.
	err = db.One("Name", "John", &user)
	fmt.Println(err)

	// Output:
	// not found
}

func ExampleDB_Find() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.Find("Group", "staff", &users)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Found", len(users))

	// Output:
	// Found 3
}

func ExampleDB_All() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.All(&users)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Found", len(users))

	// Output:
	// Found 3
}

func ExampleDB_AllByIndex() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.AllByIndex("CreatedAt", &users)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Found", len(users))

	// Output:
	// Found 3
}

func ExampleDB_Range() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.Range("Age", 21, 22, &users)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Found", len(users))

	// Output:
	// Found 2
}

func ExampleLimit() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.All(&users, storm.Limit(2))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Found", len(users))

	// Output:
	// Found 2
}

func ExampleSkip() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User
	err := db.All(&users, storm.Skip(1))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Found", len(users))

	// Output:
	// Found 2
}

func ExampleDB_Remove() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var user User

	err := db.One("ID", 1, &user)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Remove(user)
	fmt.Println(err)

	// Output:
	// <nil>
}

func ExampleDB_Begin() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	// both start out with a balance of 10000 cents
	var account1, account2 Account

	tx, err := db.Begin(true)

	if err != nil {
		log.Fatal(err)
	}

	err = tx.One("ID", 1, &account1)

	if err != nil {
		log.Fatal(err)
	}

	err = tx.One("ID", 2, &account2)

	if err != nil {
		log.Fatal(err)
	}

	account1.Amount -= 1000
	account2.Amount += 1000

	err = tx.Save(account1)

	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	err = tx.Save(account2)

	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()

	var account1Reloaded, account2Reloaded Account

	err = db.One("ID", 1, &account1Reloaded)

	if err != nil {
		log.Fatal(err)
	}

	err = db.One("ID", 2, &account2Reloaded)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Amount in account 1:", account1Reloaded.Amount)
	fmt.Println("Amount in account 2:", account2Reloaded.Amount)

	// Output:
	// Amount in account 1: 9000
	// Amount in account 2: 11000
}

type User struct {
	ID        int    `storm:"id"`
	Group     string `storm:"index"`
	Email     string `storm:"unique"`
	Name      string
	Age       int       `storm:"index"`
	CreatedAt time.Time `storm:"index"`
}

type Account struct {
	ID     int   `storm:"id"`
	Amount int64 // amount in cents
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
			log.Fatal(err)
		}
	}

	for i := int64(0); i < 10; i++ {
		account := Account{Amount: 10000}

		err := db.Save(&account)

		if err != nil {
			log.Fatal(err)
		}
	}

	return dir, db

}
