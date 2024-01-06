package q_test

import (
	"fmt"
	"log"

	"time"

	"os"

	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/AndersonBargas/rainstorm/v5"
	"github.com/AndersonBargas/rainstorm/v5/q"
)

func ExampleRe() {
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	var users []User

	// Find all users with name that starts with the letter D.
	if err := db.Select(q.Re("Name", "^D")).Find(&users); err != nil {
		log.Println("error: Select failed:", err)
		return
	}

	// Donald and Dilbert
	fmt.Println("Found", len(users), "users.")

	// Output:
	// Found 2 users.
}

type User struct {
	ID        int    `rainstorm:"id,increment"`
	Group     string `rainstorm:"index"`
	Email     string `rainstorm:"unique"`
	Name      string
	Age       int       `rainstorm:"index"`
	CreatedAt time.Time `rainstorm:"index"`
}

func prepareDB() (string, *rainstorm.DB) {
	dir, _ := ioutil.TempDir(os.TempDir(), "rainstorm")
	db, _ := rainstorm.Open(filepath.Join(dir, "rainstorm.db"))

	for i, name := range []string{"John", "Norm", "Donald", "Eric", "Dilbert"} {
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

	return dir, db
}
