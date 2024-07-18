package status

type Status string

const (
	InProcessing = Status("InProcessing")
	Processed    = Status("Processed")
)
