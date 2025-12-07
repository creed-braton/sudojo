package lobby

type Artifact interface {
}

type artifact struct {
	actor     string
	timestamp int64
	row       int
	column    int
	value     int
}

var _ Artifact = &artifact{}

func NewArtifact(actor string, ts int64, row, col, val int) *artifact {
	return &artifact{}
}
