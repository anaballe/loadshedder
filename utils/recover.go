package utils

import (
	"go.uber.org/zap"
)

// RunSafe runs the given fn, recovers if fn panics.
func RunSafe(fn func()) {
	defer Recover()

	fn()
}

// Recover is used with defer to do cleanup on panics.
// Use it like:
//
//	defer Recover(func() {})
func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		zap.S().Error(p)
	}
}
