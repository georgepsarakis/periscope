package newcontext

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ctxKey struct{}

var ctxKeyLogger = ctxKey{}
var ctxKeyDBTransaction = ctxKey{}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	v := ctx.Value(ctxKeyLogger)
	if v == nil {
		return nil
	}
	logger, ok := v.(*zap.Logger)
	if ok {
		return logger
	}
	return nil
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, logger)
}

func WithDBTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKeyDBTransaction, tx)
}

func DBTransactionFromContext(ctx context.Context) *gorm.DB {
	v, ok := ctx.Value(ctxKeyDBTransaction).(*gorm.DB)
	if !ok {
		return nil
	}
	return v
}
