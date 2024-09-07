package bot

type HandlerFunc func(c Context) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type router struct {
	command  map[string]HandlerFunc
	callback map[string]HandlerFunc
	state    map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		command:  make(map[string]HandlerFunc),
		callback: make(map[string]HandlerFunc),
		state:    make(map[string]HandlerFunc),
	}
}

func (b *Bot) Command(name string, h HandlerFunc) {
	b.routers.command[name] = h
}

func (b *Bot) Callback(name string, h HandlerFunc) {
	b.routers.callback[name] = h
}

func (b *Bot) State(name string, h HandlerFunc) {
	b.routers.state[name] = h
}

func (b *Bot) Use(middleware ...MiddlewareFunc) {
	b.middleware = append(b.middleware, middleware...)
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
