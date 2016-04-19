package storm

// Close the database
func (s *DB) Close() error {
	return s.Bolt.Close()
}
