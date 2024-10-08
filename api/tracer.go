package api

import (
	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/logging"
	"github.com/pkg/errors"
)

// TraceError records the provided error to the given seaEntry.
func TraceError(entry *base.SeaEntry, err error) {
	defer func() {
		if e := recover(); e != nil {
			logging.Error(errors.Errorf("%+v", e), "Failed to api.TraceError()")
			return
		}
	}()

	if entry == nil || err == nil {
		return
	}

	entry.SetError(err)
}
