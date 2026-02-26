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
	Id          string
	Source      string
	SourceId    string
	Title       string
	Url         string
	Difficulty  string
	Rating      int
	Topics      []string
	Tags        []string
	StatementMd string
	Samples     []Sample
}

/*
Unmarshalled structure of how a problem summary would
be stored. This is used by ListProblems() function
*/
type ProblemSummary struct {
	Id         string
	Title      string
	Difficulty string
	Rating     string
	Source     string
	SourceId   string
}
