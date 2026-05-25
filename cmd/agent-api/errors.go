package main

import "errors"

var (
	errUnauthorized     = errors.New("unauthorized")
	errMissingInstance  = errors.New("missing X-Instance-ID or instance_id")
	errInvalidID    = errors.New("invalid id")
	errNotFound     = errors.New("not found")
	errEmptyBody    = errors.New("empty body")
	errBodyTooLarge  = errors.New("body exceeds max size")
	errInvalidMethod = errors.New("method not allowed")
)
