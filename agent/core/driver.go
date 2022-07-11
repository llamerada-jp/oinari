package core

type GlobalCommandDriver interface {
	EncouragePod(nid, uuid string) error
}

type LocalCommandDriver interface {
}
