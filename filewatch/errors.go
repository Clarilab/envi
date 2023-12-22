package filewatch

import "errors"

// ErrNoTriggers is returned when no trigger channels are provided.
var ErrNoTriggers = errors.New("no trigger channels provided")

// ErrLoaderNotSet is returned when no loader is specified.
var ErrLoaderNotSet = errors.New("no loader is specified")

// ErrNoPath is returned when no path is specified.
var ErrNoPath = errors.New("no path is specified")

// ErrEmptyPrefix is returned when the prefix is empty.
var ErrEmptyPrefix = errors.New("prefix cannot be empty")
