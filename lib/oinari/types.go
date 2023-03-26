package oinari

const (
	CrosslinkPath = "application"
)

type Application interface {
	Setup(firstInPod bool) error
	Dump() ([]byte, error)
	Restore(data []byte) error
	Teardown(lastInPod bool) error
}
