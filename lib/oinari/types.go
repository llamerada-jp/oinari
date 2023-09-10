package oinari

const (
	ApplicationCrosslinkPath = "application/api/core"
	NodeCrosslinkPath        = "node/api/core"
)

type Application interface {
	Setup(firstInPod bool) error
	Dump() ([]byte, error)
	Restore(data []byte) error
	Teardown(lastInPod bool) error
}
