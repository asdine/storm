package storm

import (
	"io"
	"time"
)

type ClassicNoTags struct {
	PublicField  int
	privateField string
	Date         time.Time
	InlineStruct struct {
		a float32
		B float64
	}
	Interf io.Writer
}

type ClassicBadTags struct {
	ID           string
	PublicField  int `storm:"mrots"`
	privateField string
	Date         time.Time
	InlineStruct struct {
		a float32
		B float64
	}
	Interf io.Writer
}

type ClassicUnique struct {
	ID            string
	PublicField   int       `storm:"unique"`
	privateField  string    `storm:"unique"`
	privateField2 string    `storm:"unique"`
	Date          time.Time `storm:"unique"`
	InlineStruct  struct {
		A float32
		B float64
	} `storm:"unique"`
	Interf io.Writer `storm:"unique"`
}

type ClassicIndex struct {
	ID           string
	PublicField  int       `storm:"index"`
	privateField string    `storm:"index"`
	Date         time.Time `storm:"index"`
	InlineStruct struct {
		a float32
		B float64
	} `storm:"index"`
	InlineStructPtr *UserWithNoID `storm:"index"`
	Interf          io.Writer     `storm:"index"`
}

type ClassicInline struct {
	PublicField  int `storm:"unique"`
	ClassicIndex `storm:"inline"`
	*ToEmbed     `storm:"inline"`
	Date         time.Time `storm:"unique"`
}

type User struct {
	ID              int       `storm:"id"`
	Name            string    `storm:"index"`
	Age             int       `storm:"index,increment"`
	DateOfBirth     time.Time `storm:"index"`
	Group           string
	unexportedField int
	Slug            string `storm:"unique"`
}

type ToEmbed struct {
	ID string
}

type NestedID struct {
	ToEmbed `storm:"inline"`
	Name    string
}

type SimpleUser struct {
	ID   int `storm:"id"`
	Name string
	age  int
}

type UserWithNoID struct {
	Name string
}

type UserWithIDField struct {
	ID   int
	Name string
}

type UserWithUint64IDField struct {
	ID   uint64
	Name string
}

type UserWithStringIDField struct {
	ID   string
	Name string
}

type UserWithEmbeddedIDField struct {
	UserWithIDField `storm:"inline"`
	Age             int
}

type UserWithEmbeddedField struct {
	UserWithNoID `storm:"inline"`
	ID           uint64
}

type UserWithIncrementField struct {
	ID   int
	Name string
	Age  int `storm:"unique,increment"`
}

type IndexedNameUser struct {
	ID          int    `storm:"id"`
	Name        string `storm:"index"`
	Score       int    `storm:"index,increment"`
	age         int
	DateOfBirth time.Time `storm:"index"`
	Group       string
}

type UniqueNameUser struct {
	ID   int    `storm:"id"`
	Name string `storm:"unique"`
	Age  int    `storm:"index,increment"`
}
