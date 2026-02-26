package store

/*
Unmarshalled structure of how a sample would
be stored in the database under a problem
*/
type Sample struct {
	Name string
	In   string
	Out  string
}

/*
Unmarshalled structure of how a problem would
be stored in the database
*/
type Problem struct {
	Id         string
	Source     string
	Title      string
	Url        string
	Difficulty string
	Rating     int
	Topics     []string
	Tags       []string
	CreatedAt  int
}

/*
Different filters that can be applied to
the ListProblems method.
*/
type UserFilters struct {
	Title      string
	Source     string
	Difficulty string
	Topic      string // Single topic filter
	Tag        string // Single tag filter
	Limit      int
}
