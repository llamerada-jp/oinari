package crosslink

const (
	TAG_PATH            = "path"
	TAG_LEAF            = "leaf"
	TAG_PATH_MATCH_KIND = "match_kind"

	PATH_MATCH_KIND_EXACT = "E"
	PATH_MATCH_KIND_HEAD  = "H"
)

type ResponseWriter interface {
	ReplySuccess(response any)
	ReplyError(message string)
}

type Handler interface {
	Serve(dataRaw []byte, tags map[string]string, writer ResponseWriter)
}

type Crosslink interface {
	Call(path string, obj any, tags map[string]string, cb func([]byte, error))
}

type MultiPlexer interface {
	Handler
	SetHandler(pattern string, handler Handler)
	SetDefaultHandler(handler Handler)
}
