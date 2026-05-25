package main

import "errors"

var (
	errUnauthorized = errors.New("unauthorized")
	errInvalidID    = errors.New("invalid id")
	errNotFound     = errors.New("not found")
	errEmptyBody    = errors.New("empty body")
	errBodyTooLarge  = errors.New("body exceeds max size")
	errInvalidMethod = errors.New("method not allowed")
)
