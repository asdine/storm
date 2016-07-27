package index

// NewOptions creates initialized Options
func NewOptions() *Options {
	return &Options{
		Limit:   -1,
		Skip:    0,
		Reverse: false,
	}
}

// Options are used to customize queries
type Options struct {
	Limit   int
	Skip    int
	Reverse bool
}
