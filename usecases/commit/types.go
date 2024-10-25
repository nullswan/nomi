package commit

type fileCommitPlan struct {
	CommitPlan []fileAction `json:"commitPlan"`
}

type fileAction struct {
	Files         []string `json:"files"`
	CommitMessage string   `json:"commitMessage"`
}
