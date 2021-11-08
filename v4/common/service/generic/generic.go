package generic

import (
	"context"
)

type Server interface {
	Handle(Handler)
	Serve() error
}

type Handler func(Server) error

type genericServer struct {
	ctx      context.Context
	handlers []Handler
}

func NewGenericServer(ctx context.Context) Server {
	return &genericServer{ctx: ctx}
}

func (g *genericServer) Handle(h Handler) {
	g.handlers = append(g.handlers, h)
}

func (g *genericServer) Serve() error {
	for _, h := range g.handlers {
		go h(g)
	}

	return nil
}
