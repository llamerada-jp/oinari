package oinari

const (
	ApplicationCrosslinkPath = "application/api/core"
	NodeCrosslinkPath        = "node/api/core"
)

type Application interface {
	Setup(isInitialize bool, record []byte) error
	Marshal() ([]byte, error)
	Teardown(isFinalize bool) ([]byte, error)
}
