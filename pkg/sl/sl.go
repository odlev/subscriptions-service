// Package sl was created to simplify and speed up certain actions
package sl

import "log/slog"

// Err is automates the key error and value, in slog.Attr (for slog.Logger.Error) "error”, “value” quick record ‘error’ “specific error”
func Err(err error) slog.Attr {
	return slog.Any("error", err)
}
