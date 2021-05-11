package storm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
)

// A Store provides methods to manipulate a single bucket of records.
type Store struct {
	name string
	db   *genji.DB
	tx   *genji.Tx
}

func (s *Store) viewTx(fn func(tx *genji.Tx) error) (err error) {
	tx := s.tx

	if tx == nil {
		tx, err = s.db.Begin(false)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	return fn(tx)
}

func (s *Store) updateTx(fn func(tx *genji.Tx) error) (err error) {
	tx := s.tx

	if tx == nil {
		tx, err = s.db.Begin(true)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}()
	}

	return fn(tx)
}

// Insert data into the store. Data can be one of the following:
// - struct or struct pointer
// - map or map pointer with string key and any other type as value
// - byte slice with valid json object
// If no primary key was specified upon store creation, it will automatically
// generate a docid and return it, otherwise it returns the encoded primary key.
func (s *Store) Insert(data interface{}) ([]byte, error) {
	param := data
	if jsonBytes, ok := data.([]byte); ok {
		param = document.NewFromJSON(jsonBytes)
	}

	var pk []byte
	err := s.updateTx(func(tx *genji.Tx) error {
		res, err := tx.Query(fmt.Sprintf("INSERT INTO %s VALUES ? RETURNING pk()", s.name), param)
		if err != nil {
			return err
		}
		defer res.Close()

		return res.Iterate(func(d document.Document) error {
			return document.Scan(d, &pk)
		})
	})
	return pk, err
}

// All selects all documents and scans them into dest. dest must be a pointer
// to a valid slice or array.
//
// It dest is a slice pointer and its capacity is too low, a new slice will be allocated.
// Otherwise, its length is set to 0 so that its content is overwritten.
//
// If dest is an array pointer, its capacity must be bigger than the length of a, otherwise an error is
// returned.
func (s *Store) All(dest interface{}) error {
	return s.Query().Find(dest)
}

// Query creates a query builder linked to the current store.
func (s *Store) Query() *Query {
	return &Query{store: s}
}

// Query is a DSL used to build SQL queries to run against Genji.
type Query struct {
	store        *Store
	whereClauses []string
	orderBy      string
	limit        int
	offset       int
	params       []interface{}
}

func (q *Query) buildClauses() string {
	var b strings.Builder

	if len(q.whereClauses) > 0 {
		b.WriteString(" WHERE ")
		for i, clause := range q.whereClauses {
			if i > 0 {
				b.WriteString(" AND ")
			}
			b.WriteString(clause)
		}
	}

	if q.orderBy != "" {
		b.WriteString(" ORDER BY ")
		b.WriteString(q.orderBy)
	}

	if q.limit != 0 {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.Itoa(q.limit))
	}

	if q.offset != 0 {
		b.WriteString(" OFFSET ")
		b.WriteString(strconv.Itoa(q.offset))
	}

	return b.String()
}

// First runs a query and scans the first result.
// The result is scanned into dest which must be either a struct pointer, a map or a map pointer.
func (q *Query) First(dest interface{}) error {
	query := "SELECT * FROM " + q.store.name + q.buildClauses()

	return q.store.viewTx(func(tx *genji.Tx) error {
		d, err := tx.QueryDocument(query, q.params...)
		if err != nil {
			return err
		}

		return document.ScanDocument(d, dest)
	})
}

// Find runs a query and scans all the results into dest. dest must be a pointer
// to a valid slice or array.
//
// It dest is a slice pointer and its capacity is too low, a new slice will be allocated.
// Otherwise, its length is set to 0 so that its content is overwritten.
//
// If dest is an array pointer, its capacity must be bigger than the length of a, otherwise an error is
// returned.
func (q *Query) Find(dest interface{}) error {
	query := "SELECT * FROM " + q.store.name + q.buildClauses()

	return q.store.viewTx(func(tx *genji.Tx) error {
		res, err := tx.Query(query, q.params...)
		if err != nil {
			return err
		}
		defer res.Close()

		return document.ScanIterator(res.Iterator, dest)
	})
}

// Offset skips n documents.
// SQL: OFFSET n
func (q *Query) Offset(n int) *Query {
	q.offset = n
	return q
}

// Limit the number of documents returned by the query.
// SQL: LIMIT n
func (q *Query) Limit(n int) *Query {
	q.limit = n
	return q
}

// OrderBy returns the documents ordered by the given expression.
// To control the order direction, append ASC or DESC.
// SQL: ORDER BY expr
func (q *Query) OrderBy(expr string) *Query {
	q.orderBy = expr
	return q
}

// Where filters out documents that don't match the predicate.
func (q *Query) Where(expr, operator string, v interface{}) *Query {
	q.params = append(q.params, v)
	q.whereClauses = append(q.whereClauses, fmt.Sprintf("%s %s $%d", expr, operator, len(q.params)))
	return q
}

// Eq adds a WHERE clause with the operator '='.
func (q *Query) Eq(expr string, v interface{}) *Query {
	return q.Where(expr, "=", v)
}

// Neq adds a WHERE clause with the operator '!='.
func (q *Query) Neq(expr string, v interface{}) *Query {
	return q.Where(expr, "!=", v)
}

// Gt adds a WHERE clause with the operator '>'.
func (q *Query) Gt(expr string, v interface{}) *Query {
	return q.Where(expr, ">", v)
}

// Gte adds a WHERE clause with the operator '>='.
func (q *Query) Gte(expr string, v interface{}) *Query {
	return q.Where(expr, ">=", v)
}

// Lt adds a WHERE clause with the operator '<'.
func (q *Query) Lt(expr string, v interface{}) *Query {
	return q.Where(expr, "<", v)
}

// Lte adds a WHERE clause with the operator '<='.
func (q *Query) Lte(expr string, v interface{}) *Query {
	return q.Where(expr, "<=", v)
}

// In adds a WHERE clause with the operator 'IN'.
func (q *Query) In(expr string, v ...interface{}) *Query {
	return q.Where(expr, "IN", v)
}

// NotIn adds a WHERE clause with the operator 'NOT IN'.
func (q *Query) NotIn(expr string, v ...interface{}) *Query {
	return q.Where(expr, "NOT IN", v)
}

// Like adds a WHERE clause with the operator 'LIKE'.
func (q *Query) Like(expr string, v interface{}) *Query {
	return q.Where(expr, "LIKE", v)
}

// NotLike adds a WHERE clause with the operator 'NOT LIKE'.
func (q *Query) NotLike(expr string, v interface{}) *Query {
	return q.Where(expr, "NOT LIKE", v)
}

// Is adds a WHERE clause with the operator 'IS'.
func (q *Query) Is(expr string, v interface{}) *Query {
	return q.Where(expr, "IS", v)
}

// IsNot adds a WHERE clause with the operator 'IS NOT'.
func (q *Query) IsNot(expr string, v interface{}) *Query {
	return q.Where(expr, "IS NOT", v)
}
