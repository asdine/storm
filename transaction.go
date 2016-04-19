package storm

// Begin starts a new transaction.
func (n Node) Begin(writable bool) (*Node, error) {
	var err error

	n.tx, err = n.s.Bolt.Begin(writable)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

// Rollback closes the transaction and ignores all previous updates.
func (n *Node) Rollback() error {
	if n.tx == nil {
		return ErrNotInTransaction
	}

	err := n.tx.Rollback()
	n.tx = nil

	return err
}

// Commit writes all changes to disk.
func (n *Node) Commit() error {
	if n.tx == nil {
		return ErrNotInTransaction
	}

	err := n.tx.Commit()
	n.tx = nil

	return err
}

// Begin starts a new transaction.
func (s *DB) Begin(writable bool) (*Node, error) {
	return s.root.Begin(writable)
}

// Rollback closes the transaction and ignores all previous updates.
func (s *DB) Rollback() error {
	return s.root.Rollback()
}

// Commit writes all changes to disk.
func (s *DB) Commit() error {
	return s.root.Rollback()
}
