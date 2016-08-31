package storm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	var u User

	err := db.Save(&User{ID: 10, Name: "John", Group: "Staff", Slug: "john"})
	assert.NoError(t, err)

	// nil
	err = db.Update(nil)
	assert.Equal(t, ErrStructPtrNeeded, err)

	// no id
	err = db.Update(&User{Name: "Jack"})
	assert.Equal(t, ErrNoID, err)

	// Unknown user
	err = db.Update(&User{ID: 11, Name: "Jack"})
	assert.Equal(t, ErrNotFound, err)

	// actual user
	err = db.Update(&User{ID: 10, Name: "Jack"})
	assert.NoError(t, err)

	err = db.One("Name", "John", &u)
	assert.Equal(t, ErrNotFound, err)

	err = db.One("Name", "Jack", &u)
	assert.NoError(t, err)
	assert.Equal(t, "Jack", u.Name)
}

func TestUpdateField(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	var u User

	err := db.Save(&User{ID: 10, Name: "John", Group: "Staff", Slug: "john"})
	assert.NoError(t, err)

	// nil
	err = db.UpdateField(nil, "", nil)
	assert.Equal(t, ErrStructPtrNeeded, err)

	// no id
	err = db.UpdateField(&User{}, "Name", "Jack")
	assert.Equal(t, ErrNoID, err)

	// Unknown user
	err = db.UpdateField(&User{ID: 11}, "Name", "Jack")
	assert.Equal(t, ErrNotFound, err)

	// Unknown field
	err = db.UpdateField(&User{ID: 11}, "Address", "Jack")
	assert.Equal(t, ErrNotFound, err)

	// Uncompatible value
	err = db.UpdateField(&User{ID: 10}, "Name", 50)
	assert.Equal(t, ErrIncompatibleValue, err)

	// actual user
	err = db.UpdateField(&User{ID: 10}, "Name", "Jack")
	assert.NoError(t, err)

	err = db.One("Name", "John", &u)
	assert.Equal(t, ErrNotFound, err)

	err = db.One("Name", "Jack", &u)
	assert.NoError(t, err)
	assert.Equal(t, "Jack", u.Name)

	// zero value
	err = db.UpdateField(&User{ID: 10}, "Name", "")
	assert.NoError(t, err)

	err = db.One("Name", "Jack", &u)
	assert.Equal(t, ErrNotFound, err)

	err = db.One("ID", 10, &u)
	assert.NoError(t, err)
	assert.Equal(t, "", u.Name)
}
