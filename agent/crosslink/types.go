package crosslink

const (
	TAG_PATH            = "path"
	TAG_LEAF            = "leaf"
	TAG_PATH_MATCH_KIND = "match_kind"

	PATH_MATCH_KIND_EXACT = "E"
	PATH_MATCH_KIND_HEAD  = "H"
)

type ResponseWriter interface {
	ReplySuccess(result string)
	ReplyError(message string)
}

type ResponseObjWriter interface {
	ReplySuccess(result any)
	ReplyError(message string)
}

type Handler interface {
	Serve(data string, tags map[string]string, writer ResponseWriter)
}

type Crosslink interface {
	Call(path string, data []byte, tags map[string]string, cb func(result []byte, err error))
}

type MultiPlexer interface {
	Handler
	SetHandler(pattern string, handler Handler)
	SetDefaultHandler(handler Handler)
}
