package storm

import (
	"github.com/genjidb/genji"
)

// DB is the main database handle. It wraps the underlying Genji database and provides
// useful methods to manipulate data.
type DB struct {
	// The root node that points to the root bucket.
	db *genji.DB
}

// Open a database at the given path. If path is a valid file path, it will open a BoltDB database at that path.
// If path is the string ":memory:", it will open an in-memory database.
func Open(path string) (*DB, error) {
	var err error

	db, err := genji.Open(path)
	if err != nil {
		return nil, err
	}

	s := DB{
		db: db,
	}

	return &s, nil
}

// Close the database.
func (db *DB) Close() error {
	return db.db.Close()
}

// Begin a read-only or read/write transaction.
func (db *DB) Begin(writable bool) (*Tx, error) {
	tx, err := db.db.Begin(writable)
	if err != nil {
		return nil, err
	}

	return &Tx{
		db: db.db,
		tx: tx,
	}, nil
}

// CreateStore creates a store in the underlying engine and returns it.
func (db *DB) CreateStore(storeName string) (*Store, error) {
	err := db.db.Update(func(tx *genji.Tx) error {
		return tx.CreateTable(storeName, nil)
	})
	if err != nil {
		return nil, err
	}

	return db.Store(storeName), nil
}

// GetStore returns a store if it exists.
func (db *DB) GetStore(storeName string) (*Store, error) {
	err := db.db.View(func(tx *genji.Tx) error {
		_, err := tx.GetTable(storeName)
		return err
	})
	if err != nil {
		return nil, err
	}

	return db.Store(storeName), nil
}

// Store returns a store. If the store doesn't exists, calls to the store methods will fail.
func (db *DB) Store(storeName string) *Store {
	return &Store{name: storeName, db: db.db}
}

// Tx represents a database transaction. It provides methods for managing records and stores.
// Tx is either read-only or read/write. Read-only can be used to read stores and read/write can be used to read, create, delete and modify stores.
type Tx struct {
	db *genji.DB
	tx *genji.Tx
}

// Rollback the transaction. Can be used safely after commit.
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// Commit the transaction. Calling this method on read-only transactions
// will return an error.
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

// CreateStore creates a store in the underlying engine and returns it.
func (tx *Tx) CreateStore(storeName string) (*Store, error) {
	err := tx.tx.CreateTable(storeName, nil)
	if err != nil {
		return nil, err
	}

	return &Store{
		name: storeName,
	}, nil
}

// GetStore returns a store if it exists.
func (tx *Tx) GetStore(storeName string) (*Store, error) {
	_, err := tx.tx.GetTable(storeName)
	if err != nil {
		return nil, err
	}

	return tx.Store(storeName), nil
}

// Store returns a store. If the store doesn't exists, calls to the store methods will fail.
func (tx *Tx) Store(storeName string) *Store {
	return &Store{
		name: storeName,
		db:   tx.db,
		tx:   tx.tx,
	}
}
