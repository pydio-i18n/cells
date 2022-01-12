package middleware

import (
	"context"

	servercontext "github.com/pydio/cells/v4/common/server/context"

	clientcontext "github.com/pydio/cells/v4/common/client/context"
)

func ClientConnIncomingContext(serverRuntimeContext context.Context) func(ctx context.Context) (context.Context, bool, error) {
	clientConn := clientcontext.GetClientConn(serverRuntimeContext)
	return func(ctx context.Context) (context.Context, bool, error) {
		return clientcontext.WithClientConn(ctx, clientConn), true, nil
	}
}

func RegistryIncomingContext(serverRuntimeContext context.Context) func(ctx context.Context) (context.Context, bool, error) {
	registry := servercontext.GetRegistry(serverRuntimeContext)
	return func(ctx context.Context) (context.Context, bool, error) {
		return servercontext.WithRegistry(ctx, registry), true, nil
	}
}
