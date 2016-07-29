package q

type Criteria interface{}

type gt struct{}
type lt struct{}
type eq struct{}

type or struct {
	Children []Criteria
}

type and struct {
	Children []Criteria
}

func Gt(v interface{}) Criteria { return &gt{} }
func Lt(v interface{}) Criteria { return &lt{} }
func Eq(v interface{}) Criteria { return &eq{} }

func Or(criterias ...Criteria) Criteria  { return &or{Children: criterias} }
func And(criterias ...Criteria) Criteria { return &and{Children: criterias} }
