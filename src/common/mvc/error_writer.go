package mvc

import (
	"net/http"
	"strings"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ErrorWriter writes and logs every error response; a 500 or 502 logs its cause and never writes it to the body.
type ErrorWriter struct {
	logger *zap.Logger
}

func NewErrorWriter(logger *zap.Logger) *ErrorWriter {
	return &ErrorWriter{logger: logger}
}

func (w *ErrorWriter) Write(ctx fiber.Ctx, err error) error {
	if fiberError, ok := err.(*fiber.Error); ok {
		code := w.fiberErrorCode(fiberError.Code)
		return w.write(ctx, &errors.ClassifiedError{
			Status: fiberError.Code,
			Code:   code,
			Body:   fiber.Map{"code": code, "message": fiberError.Message},
		})
	}
	return w.write(ctx, errors.NewClassifiedError(err))
}

func (w *ErrorWriter) write(ctx fiber.Ctx, classified *errors.ClassifiedError) error {
	if err := ctx.Status(classified.Status).JSON(classified.Body); err != nil {
		return err
	}
	// 404s are routine probes and would drown the log.
	if classified.Status == fiber.StatusNotFound {
		return nil
	}
	fields := []zap.Field{
		zap.String("method", ctx.Method()),
		zap.String("path", ctx.OriginalURL()),
		zap.Int("status", classified.Status),
		zap.String("code", classified.Code),
		zap.ByteString("response", ctx.Response().Body()),
		w.causeField(classified),
	}
	if classified.Failed() {
		w.logger.Error("request failed", fields...)
	} else {
		w.logger.Warn("request rejected", fields...)
	}
	return nil
}

func (w *ErrorWriter) causeField(classified *errors.ClassifiedError) zap.Field {
	if classified.Cause == nil {
		return zap.Skip()
	}
	return zap.String("cause", classified.CauseMessage())
}

// fiberErrorCode turns a status into a snake_case code ("Request Entity Too Large" → request_entity_too_large).
func (w *ErrorWriter) fiberErrorCode(status int) string {
	text := http.StatusText(status)
	if text == "" {
		return "http_error"
	}
	return strings.ReplaceAll(strings.ToLower(text), " ", "_")
}
