package main

import "errors"

var (
	ErrWrongStatus        = errors.New("Wrong status")
	ErrRequestResultIsNil = errors.New("Request result is nil")
)
