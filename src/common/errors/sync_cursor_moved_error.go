package errors

import stderrors "errors"

// ErrSyncCursorMoved is returned when another replica advanced the cursor first.
var ErrSyncCursorMoved = stderrors.New("sync cursor was moved by another replica")
