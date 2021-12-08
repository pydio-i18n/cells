/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package servicecontext

import (
	"context"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	metadata2 "github.com/pydio/cells/v4/common/service/context/metadata"
)

func MetaUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			localMd := make(map[string]string, len(md))
			for k, v := range md {
				localMd[k] = strings.Join(v, "")
			}
			ctx = metadata2.NewContext(ctx, localMd)
		}
		return handler(ctx, req)
	}
}

func MetaStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			localMd := make(map[string]string, len(md))
			for k, v := range md {
				localMd[k] = strings.Join(v, "")
			}
			ctx = metadata2.NewContext(ctx, localMd)
			wrapped := grpc_middleware.WrapServerStream(stream)
			wrapped.WrappedContext = ctx
			return handler(srv, wrapped)
		}
		return handler(srv, stream)
	}
}
