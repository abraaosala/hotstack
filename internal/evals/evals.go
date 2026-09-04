package evals

type Case struct {
	Base     string
	BaseDir  string
	Intent   string
	Graders  []Grader
	Context  string
}

type Grader struct {
	Type       string `toml:"type"`
	Tool       string `toml:"tool"`
	InputMatch string `toml:"input_match"`
	Path       string `toml:"path"`
	Pattern    string `toml:"pattern"`
	Match      string `toml:"match"`
	Min        int    `toml:"min"`
	Max        int    `toml:"max"`
	Code       int    `toml:"code"`
	Command    string `toml:"command"`
	Output     string `toml:"output"`
	Snapshot   string `toml:"snapshot"`
}

type Result struct {
	CaseName string
	Passed   []string
	Failed   []string
}

func (r Result) PassedCount() int { return len(r.Passed) }
func (r Result) FailedCount() int { return len(r.Failed) }
func (r Result) OK() bool         { return len(r.Failed) == 0 }
