package service

import (
	"context"
	"go.uber.org/zap"

	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

// WithLogger adds a logger to the service context
func WithLogger(logger *zap.Logger) ServiceOption {
	return func(o *ServiceOptions) {
		o.BeforeInit = append(o.BeforeInit, func(ctx context.Context) error {
			ctx = servicecontext.WithLogger(ctx, logger.Named(o.Name))

			o.Context = ctx

			return nil
		})
	}
}