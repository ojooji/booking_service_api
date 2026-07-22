package repository

import "errors"

// ErrNotFound is returned by repositories when the requested row does not exist.
var ErrNotFound = errors.New("record not found")
