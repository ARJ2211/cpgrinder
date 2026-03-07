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

/*
Json Format for the attemt so we can
marshal / unmarshal it
*/
type Attempt struct {
	ID               string  `json:"id"`
	ProblemID        string  `json:"problem_id"`
	DailySetID       *string `json:"daily_set_id,omitempty"`
	StartedAt        *int64  `json:"started_at,omitempty"`
	FinishedAt       *int64  `json:"finished_at,omitempty"`
	Status           string  `json:"status"`
	Verdict          string  `json:"verdict"`
	Notes            string  `json:"notes"`
	Language         string  `json:"language"`
	TimeSpentSeconds int     `json:"time_spent_seconds"`
	CreatedAt        int64   `json:"created_at"`
}
