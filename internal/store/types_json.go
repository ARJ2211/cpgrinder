package store

/*
SampleJson represents the unmarshalled structure of how a sample
would be stored in the database under a problem.
*/
type SampleJson struct {
	Name string `json:"name"`
	In   string `json:"in"`
	Out  string `json:"out"`
}

/*
ProblemJson represents the unmarshalled structure of how a problem
would be stored in the database.
*/
type ProblemJson struct {
	ID          string       `json:"id"`
	Source      string       `json:"source"`
	SourceID    string       `json:"source_id"`
	Title       string       `json:"title"`
	URL         string       `json:"url"`
	Difficulty  string       `json:"difficulty"`
	Rating      int          `json:"rating"`
	Topics      []string     `json:"topics"`
	Tags        []string     `json:"tags"`
	StatementMd string       `json:"statement_md"`
	Samples     []SampleJson `json:"samples"`
}
