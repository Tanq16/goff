package ops

type OpCategory string

const (
	CategoryVideo OpCategory = "video"
	CategoryAudio OpCategory = "audio"
	CategoryMulti OpCategory = "multi"
)

type OpResult struct {
	Args      []string
	Suffix    string
	TargetExt string
}
