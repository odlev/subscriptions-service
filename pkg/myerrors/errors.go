// Package myerrors is a nice package
package myerrors

import "errors"

var ( 
	ErrNotFound = errors.New("subscription not found")
	ErrInvalidDateRange = errors.New("end_date can not be earlier than start_date")
)
