package types

import "log/slog"

func LogError(logger *slog.Logger, msg string, err error, attrs ...any) {
	logger.Error(msg, append(attrs, slog.Any("err", err))...)
}
