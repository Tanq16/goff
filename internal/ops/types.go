package ops

type OpResult struct {
	Args       []string
	Suffix     string
	TargetExt  string
	OutputPath string
	Cleanup    func()
}
