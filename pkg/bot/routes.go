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

func (b *Bot) Command(name string, h HandlerFunc) {
	b.Add(OnCommand, name, h)
}

func (b *Bot) Message(name string, h HandlerFunc) {
	b.Add(OnMessage, name, h)
}

func (b *Bot) Callback(name string, h HandlerFunc) {
	b.Add(OnCallback, name, h)
}

func (b *Bot) State(name string, h HandlerFunc) {
	b.Add(OnState, name, h)
}

func (b *Bot) Add(m method, name string, h HandlerFunc) {
	b.routers[m].static[name] = h
}

func (b *Bot) AddTree(m method, path string, h HandlerFunc) {
	b.routers[m].addTree(path, h)
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

// Опять вдохновился echo
type node struct {
	kind    kind
	path    string
	child   []*node
	handler HandlerFunc
}

func (r *route) addTree(path string, h HandlerFunc) {
	segments := strings.Split(path, "/")
	currNode := r.tree

	for _, seg := range segments {
		if seg == "" {
			continue
		}

		c := r.findChild(currNode, seg)
		if c == nil {
			c = &node{
				path: seg,
				kind: staticKind,
			}
			if strings.HasPrefix(seg, ":") {
				c.kind = paramKind
			}
			currNode.child = append(currNode.child, c)
		}
		currNode = c
	}
	currNode.handler = h
}

func (r *route) findTree(c Context, path string) (HandlerFunc, bool) {
	ctx := c.(*nativeContext)
	segments := strings.Split(path, "/")
	currNode := r.tree

	for _, seg := range segments {
		if seg == "" {
			continue
		}

		child := r.findChild(currNode, seg)
		if child == nil {
			return nil, false
		}

		if child.kind == paramKind {
			p := strings.TrimPrefix(child.path, ":")
			ctx.setParam(p, seg)
		}
		currNode = child
	}
	return currNode.handler, currNode.handler != nil
}

func (r *route) findChild(n *node, seg string) *node {
	for _, c := range n.child {
		if c.path == seg || c.kind == paramKind {
			return c
		}
	}
	return nil
}

func (r *route) find(c Context, path string) (HandlerFunc, bool) {
	if f, ok := r.static[path]; ok {
		return f, ok
	}
	return r.findTree(c, path)
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
