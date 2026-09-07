package ops

type OpResult struct {
	Args       []string
	Suffix     string
	TargetExt  string
	OutputPath string
	Notes      []string
	Cleanup    func()
}
