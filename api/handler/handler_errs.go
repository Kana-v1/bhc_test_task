package handler

import "errors"

var ErrFailedToReadBody = errors.New("failed to read or parse request body")
