package storm

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkFindWithIndex(b *testing.B) {
	db, cleanup := createDB(b)
	defer cleanup()

	var users []User
	for i := 0; i < 100; i++ {
		var w User

		if i%2 == 0 {
			w.Name = "John"
			w.Group = "Staff"
		} else {
			w.Name = "Jack"
			w.Group = "Admin"
		}
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.Find("Name", "John", &users)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkFindWithoutIndex(b *testing.B) {
	db, cleanup := createDB(b)
	defer cleanup()

	var users []User
	for i := 0; i < 100; i++ {
		var w User

		if i%2 == 0 {
			w.Name = "John"
			w.Group = "Staff"
		} else {
			w.Name = "Jack"
			w.Group = "Admin"
		}
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.Find("Group", "Staff", &users)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkOneWithIndex(b *testing.B) {
	db, cleanup := createDB(b)
	defer cleanup()

	var u User
	for i := 0; i < 100; i++ {
		w := User{Name: fmt.Sprintf("John%d", i), Group: fmt.Sprintf("Staff%d", i)}
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.One("Name", "John99", &u)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkOneByID(b *testing.B) {
	db, cleanup := createDB(b)
	defer cleanup()

	type User struct {
		ID          int
		Name        string `storm:"index"`
		age         int
		DateOfBirth time.Time `storm:"index"`
		Group       string
		Slug        string `storm:"unique"`
	}

	var u User
	for i := 0; i < 100; i++ {
		w := User{Name: fmt.Sprintf("John%d", i), Group: fmt.Sprintf("Staff%d", i)}
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.One("ID", 99, &u)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkOneWithoutIndex(b *testing.B) {
	db, cleanup := createDB(b)
	defer cleanup()

	var u User
	for i := 0; i < 100; i++ {
		w := User{Name: "John", Group: fmt.Sprintf("Staff%d", i)}
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.One("Group", "Staff99", &u)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkSave(b *testing.B) {
	db, cleanup := createDB(b)
	defer cleanup()

	w := User{Name: "John"}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}
}
