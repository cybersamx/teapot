package httpx

type HandlerWrapperFactory struct {
	Server         *Server
	preMiddleware  []HandlerFunc
	postMiddleware []HandlerFunc
}

func NewHandlerWrapperFactory(server *Server) *HandlerWrapperFactory {
	return &HandlerWrapperFactory{
		Server: server,
	}
}

func (hwf *HandlerWrapperFactory) Handler(handlerFn HandlerFunc) *HandlerWrapper {
	hw := HandlerWrapper{
		Factory:     hwf,
		HandlerFunc: handlerFn,
	}

	return &hw
}

func (hwf *HandlerWrapperFactory) AddPreMiddleware(hf ...HandlerFunc) {
	hwf.preMiddleware = append(hwf.preMiddleware, hf...)
}

func (hwf *HandlerWrapperFactory) AddPostMiddleware(hf ...HandlerFunc) {
	hwf.postMiddleware = append(hwf.postMiddleware, hf...)
}
