package ops

type OpResult struct {
	Args        []string
	Suffix      string
	TargetExt   string
	OutputPaths []string
	Notes       []string
	Cleanup     func()
}
