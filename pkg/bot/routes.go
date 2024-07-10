package bot

// Вдохновился echo фреймворком ребята...

type HandlerFunc func(c Context) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type Group struct {
	middleware []MiddlewareFunc
	routes     map[string]HandlerFunc
}

func (b *Bot) NewGroup(name string) *Group {
	g := &Group{routes: make(map[string]HandlerFunc)}
	b.routers[name] = g
	return g
}

func (b *Bot) Use(middleware ...MiddlewareFunc) {
	b.middleware = append(b.middleware, middleware...)
}

func (g *Group) AddRoute(name string, handler HandlerFunc) {
	g.routes[name] = handler
}

func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
