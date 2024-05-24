package logger

import (
	"context"
	"fmt"
)

// TODO: https://www.youtube.com/watch?v=p9XiOOU52Qw&t=1591s - подумать о добавлении инфо в ошибку на уровне
// Структура данных в middleware адаптируется под конкретный проект.

type logCtx struct {
	requestID int
	userID    int
}

type keyCtx int

const keyLog = keyCtx(0)

func WithLogRequestID(ctx context.Context, requestID int) context.Context {
	if logRec, ok := ctx.Value(keyLog).(logCtx); ok {
		logRec.requestID = requestID

		return context.WithValue(ctx, keyLog, logRec)
	}

	return context.WithValue(ctx, keyLog, logCtx{requestID: requestID}) //nolint:exhaustruct // only one
}

func WithLogUserID(ctx context.Context, userID int) context.Context {
	if logRec, ok := ctx.Value(keyLog).(logCtx); ok {
		logRec.userID = userID

		return context.WithValue(ctx, keyLog, logRec)
	}

	return context.WithValue(ctx, keyLog, logCtx{userID: userID}) //nolint:exhaustruct // only one
}

type HandlerMiddleware struct {
	next Handler
}

func NewHandlerMiddleware(next Handler) *HandlerMiddleware {
	return &HandlerMiddleware{next: next}
}

func (h *HandlerMiddleware) Enabled(ctx context.Context, rec Level) bool {
	return h.next.Enabled(ctx, rec)
}

func (h *HandlerMiddleware) Handle(ctx context.Context, rec Record) error { // TODO: заменить реализацию на мапу, вместо структуры
	if logRec, ok := ctx.Value(keyLog).(logCtx); ok {
		if logRec.requestID != 0 {
			rec.Add("request-id", logRec.requestID)
		}
		if logRec.userID != 0 {
			rec.Add("user-id", logRec.userID)
		}
	}

	return fmt.Errorf("next handle: %w", h.next.Handle(ctx, rec))
}

func (h *HandlerMiddleware) WithAttrs(attrs []Attr) Handler { //nolint:ireturn,nolintlint // implementation Handler interface
	return &HandlerMiddleware{next: h.next.WithAttrs(attrs)}
}

func (h *HandlerMiddleware) WithGroup(name string) Handler { //nolint:ireturn,nolintlint // implementation Handler interface
	return &HandlerMiddleware{next: h.next.WithGroup(name)}
}
