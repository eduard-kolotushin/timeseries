package timeseries

import "errors"

var (
	// ErrLengthMismatch is returned when times and values have different lengths.
	ErrLengthMismatch = errors.New("timeseries: times and values length mismatch")

	// ErrUnsorted is returned when timestamps are not strictly ascending.
	ErrUnsorted = errors.New("timeseries: timestamps must be strictly ascending")

	// ErrDuplicateTime is returned when duplicate timestamps are present.
	ErrDuplicateTime = errors.New("timeseries: duplicate timestamps are not allowed")

	// ErrEmpty is returned when an operation requires a non-empty series.
	ErrEmpty = errors.New("timeseries: series is empty")

	// ErrInvalidWindow is returned when a window size is not positive.
	ErrInvalidWindow = errors.New("timeseries: window must be positive")

	// ErrInvalidDuration is returned when a duration is not positive.
	ErrInvalidDuration = errors.New("timeseries: duration must be positive")

	// ErrIndexOutOfRange is returned for invalid index access.
	ErrIndexOutOfRange = errors.New("timeseries: index out of range")

	// ErrConflict is returned when merge encounters conflicting values at the same time.
	ErrConflict = errors.New("timeseries: conflicting values at the same timestamp")
)
