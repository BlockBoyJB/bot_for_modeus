package bot

import (
	"strings"
)

type HandlerFunc func(c Context) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type method uint8

// Типы входящих запросов
const (
	OnCommand method = iota
	OnMessage
	OnCallback
	OnState
)

type router = map[method]*route

func newRouter() router {
	return router{
		OnCommand:  newRoute(),
		OnMessage:  newRoute(),
		OnCallback: newRoute(),
		OnState:    newRoute(),
	}
}

// Router defines all router handle interfaces, including bot and group router.
type Router interface {
	Add(m method, p string, h HandlerFunc, middleware ...MiddlewareFunc)
	Command(name string, h HandlerFunc, m ...MiddlewareFunc)
	Message(name string, h HandlerFunc, m ...MiddlewareFunc)
	Callback(name string, h HandlerFunc, m ...MiddlewareFunc)
	State(name string, h HandlerFunc, m ...MiddlewareFunc)

	// AddTree регистрирует обработчик в дерево с возможностью задавать сегменты пути в виде параметрических переменных
	AddTree(m method, path string, h HandlerFunc, middleware ...MiddlewareFunc)

	// Use добавляет новый мидлварь в конец стека
	Use(middleware ...MiddlewareFunc)

	// PreUse добавляет мидлварь в стек предобработчиков.
	// Это означает, что мидлварь из этого стека будут вызваны сразу после основного хэндлера, но до основного стека мидлварей.
	// Удобно, если нужна миддлварь, которая будет вызываться перед всеми остальными вне зависимости от их порядка добавления
	PreUse(middleware ...MiddlewareFunc)

	// Group создает новую группу обработчиков, в которой удобно работать со своим пространством мидлварей.
	// Все мидлвари, добавленные в группу распространяются только на хэндлеры группы.
	Group(m ...MiddlewareFunc) Router
}

func (b *Bot) Add(m method, p string, h HandlerFunc, middleware ...MiddlewareFunc) {
	stack := append(b.middleware, middleware...)
	b.routers[m].static[p] = applyMiddleware(h, append(stack, b.premiddleware...)...)
}

func (b *Bot) Command(name string, h HandlerFunc, m ...MiddlewareFunc) {
	b.Add(OnCommand, name, h, m...)
}

func (b *Bot) Message(name string, h HandlerFunc, m ...MiddlewareFunc) {
	b.Add(OnMessage, name, h, m...)
}

func (b *Bot) Callback(name string, h HandlerFunc, m ...MiddlewareFunc) {
	b.Add(OnCallback, name, h, m...)
}

func (b *Bot) State(name string, h HandlerFunc, m ...MiddlewareFunc) {
	b.Add(OnState, name, h, m...)
}

func (b *Bot) AddTree(m method, path string, h HandlerFunc, middleware ...MiddlewareFunc) {
	stack := append(b.middleware, middleware...)
	b.routers[m].addTree(path, applyMiddleware(h, append(stack, b.premiddleware...)...))
}

func (b *Bot) Use(middleware ...MiddlewareFunc) {
	b.middleware = append(b.middleware, middleware...)
}

func (b *Bot) PreUse(middleware ...MiddlewareFunc) {
	b.premiddleware = append(b.premiddleware, middleware...)
}

type group struct {
	parent        Router
	middleware    []MiddlewareFunc
	premiddleware []MiddlewareFunc
}

func (b *Bot) Group(m ...MiddlewareFunc) Router {
	return &group{
		parent:     b,
		middleware: m,
	}
}

func (g *group) Add(m method, name string, h HandlerFunc, middleware ...MiddlewareFunc) {
	stack := append(g.middleware, middleware...)
	g.parent.Add(m, name, h, append(stack, g.premiddleware...)...)
}

func (g *group) Command(name string, h HandlerFunc, m ...MiddlewareFunc) {
	g.Add(OnCommand, name, h, m...)
}

func (g *group) Message(name string, h HandlerFunc, m ...MiddlewareFunc) {
	g.Add(OnMessage, name, h, m...)
}

func (g *group) Callback(name string, h HandlerFunc, m ...MiddlewareFunc) {
	g.Add(OnCallback, name, h, m...)
}

func (g *group) State(name string, h HandlerFunc, m ...MiddlewareFunc) {
	g.Add(OnState, name, h, m...)
}

func (g *group) AddTree(m method, path string, h HandlerFunc, middleware ...MiddlewareFunc) {
	stack := append(g.middleware, middleware...)
	g.parent.AddTree(m, path, h, append(stack, g.premiddleware...)...)
}

func (g *group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

func (g *group) PreUse(m ...MiddlewareFunc) {
	g.premiddleware = append(g.premiddleware, m...)
}

func (g *group) Group(m ...MiddlewareFunc) Router {
	return &group{
		parent:     g,
		middleware: m,
	}
}

// Структура одного направления поддерживает как статические ручки (для O(1)), так и в виде дерева с возможностью более гибкой маршрутизации
type route struct {
	tree   *node
	static map[string]HandlerFunc
}

func newRoute() *route {
	return &route{
		tree:   &node{},
		static: make(map[string]HandlerFunc),
	}
}

// Представляет собой тип узла (статический, переменная и тд)
type kind uint8

const (
	staticKind kind = iota
	paramKind
)

// Узел для динамической маршрутизации.
// Изначально была потребность в гибких коллбэках с параметрами.
// Это был переход от fsm и состояний к гибкости и скорости (в кэш смотреть, очевидно, дольше, чем гибкий коллбэк)
// Текущий вариант поддерживает как статический сегмент пути, так и параметрический (который может меняться)
type node struct {
	kind    kind
	handler HandlerFunc
	static  map[string]*node
	param   *node
	path    string
}

func (r *route) addTree(path string, h HandlerFunc) {
	currNode := r.tree

	for len(path) > 0 {
		seg := path
		if idx := strings.IndexByte(path, '/'); idx != -1 {
			seg, path = path[:idx], path[idx+1:]
		} else {
			path = ""
		}

		if seg == "" {
			continue
		}

		var c *node
		if seg[0] == ':' {
			if currNode.param == nil {
				currNode.param = &node{
					path: seg,
					kind: paramKind,
				}
			}
			c = currNode.param
		} else { // Это статический узел
			if currNode.static == nil {
				currNode.static = make(map[string]*node)
			}
			c = currNode.static[seg]
			if c == nil {
				c = &node{
					path: seg,
					kind: staticKind,
				}
				currNode.static[seg] = c
			}
		}
		currNode = c
	}
	currNode.handler = h
}

func (r *route) findTree(c Context, path string) (HandlerFunc, bool) {
	ctx := c.(*nativeContext)
	currNode := r.tree

	for len(path) > 0 {
		seg := path
		if index := strings.IndexByte(path, '/'); index != -1 {
			seg, path = path[:index], path[index+1:]
		} else {
			path = ""
		}

		if seg == "" {
			continue
		}

		child := r.findChild(currNode, seg)
		if child == nil {
			return nil, false
		}

		if child.kind == paramKind {
			ctx.setParam(child.path[1:], seg)
		}
		currNode = child
	}
	return currNode.handler, currNode.handler != nil
}

func (r *route) findChild(n *node, seg string) *node {
	if n.static != nil {
		if c, ok := n.static[seg]; ok {
			return c
		}
	}
	if n.param != nil {
		return n.param
	}
	return nil
}

func (r *route) find(c Context, path string) (HandlerFunc, bool) {
	if f, ok := r.static[path]; ok {
		return f, ok
	}
	return r.findTree(c, path)
}

// applyMiddleware функция, которая "навешивает" на хэндлер мидлвари.
// Важно понимать, что добавляет она их по принципу стека,
// т.е самый первый из слайса будет добавлен последним (станет внешним), а последний - первым (внутренним)
func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
