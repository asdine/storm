package storm

// Close the database
func (s *DB) Close() {
	s.Bolt.Close()
}
