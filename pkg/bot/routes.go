package bot

// Типы входящих сообщений. Можно было сделать их публичными и передавать в одну функцию обработчик в качестве аргумента, но не хочу =)
const (
	onCommand = iota
	onMessage
	onCallback
	onState
)

type HandlerFunc func(c Context) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type (
	// алиасы для удобства
	route  = map[string]HandlerFunc
	router = map[int]route
)

func newRouter() router {
	return router{
		onCommand:  route{},
		onMessage:  route{},
		onCallback: route{},
		onState:    route{},
	}
}

func (b *Bot) Command(name string, h HandlerFunc) {
	b.routers[onCommand][name] = h
}

func (b *Bot) Message(name string, h HandlerFunc) {
	b.routers[onMessage][name] = h
}

func (b *Bot) Callback(name string, h HandlerFunc) {
	b.routers[onCallback][name] = h
}

func (b *Bot) State(name string, h HandlerFunc) {
	b.routers[onState][name] = h
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
