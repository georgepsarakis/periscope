package http

import (
	"context"
	"encoding/json"
	"errors"

	"go.uber.org/zap"

	"github.com/georgepsarakis/periscope/newcontext"
)

const ErrorCodeServerError = 500
const ErrorCodeJSONDecodingFailed = 1001
const ErrorCodeValidationFailed = 1002

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewJSONError(message string, code int) []byte {
	b, err := json.Marshal(Error{Message: message, Code: code})
	if err != nil {
		panic(err)
	}
	return b
}

func NewServerError(ctx context.Context, err error) []byte {
	logger := newcontext.LoggerFromContext(ctx)
	logCtx := []zap.Field{zap.Error(err)}
	if f := ZapErrorLogContext(err); len(f) > 0 {
		logCtx = append(logCtx, f...)
	}
	if logger != nil {
		logger.Error("unexpected error", logCtx...)
	}
	return NewJSONError("Server Error", ErrorCodeServerError)
}

type ZapError struct {
	originalErr error
	logCtx      []zap.Field
}

func NewZapError(err error, fields ...zap.Field) ZapError {
	return ZapError{originalErr: err, logCtx: fields}
}

func (e ZapError) Error() string {
	return e.originalErr.Error()
}

func (e ZapError) Unwrap() error {
	return e.originalErr
}

func (e ZapError) LogContext() []zap.Field {
	return e.logCtx
}

func ZapErrorLogContext(err error) []zap.Field {
	if err == nil {
		return nil
	}
	var zapErr *ZapError
	if errors.As(err, &zapErr) {
		return zapErr.LogContext()
	}
	return nil
}
