package cmp

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrLenSlicesNotEquals    = errors.New("Len slices not equals")
	ErrLenMapsNotEquals      = errors.New("Len maps not equals")
	ErrExpectNotFoundInArray = errors.New("Expect not found in array")
	ErrDifferentObjectTypes  = errors.New("Different object types")
	ErrUnknownType           = errors.New("Unknown type")
)

type ErrCmp struct {
	Field string
	Err   error
}

func NewErrCmpIndex(index int, err error) *ErrCmp {
	return &ErrCmp{
		Field: fmt.Sprintf("[%d]", index),
		Err:   err,
	}
}

func NewErrCmpField(field string, err error) *ErrCmp {
	return &ErrCmp{
		Field: field,
		Err:   err,
	}
}

func (e *ErrCmp) Error() string {
	var bs bytes.Buffer
	bs.WriteString(e.Field)

	nextErr := e.Err
	for {
		err, ok := nextErr.(*ErrCmp)
		if !ok {
			break
		}
		bs.WriteRune('.')
		bs.WriteString(err.Field)
		nextErr = err.Err
	}

	return fmt.Sprintf("not equals %q: %s", bs.String(), nextErr.Error())
}

type ErrDifferentTypes struct {
	ResType, ExpType reflect.Type
}

func NewErrDifferentTypes(res, exp interface{}) *ErrDifferentTypes {
	return &ErrDifferentTypes{
		ResType: reflect.TypeOf(res),
		ExpType: reflect.TypeOf(exp),
	}
}

func (e *ErrDifferentTypes) Error() string {
	return fmt.Sprintf("different types: %s (result), %s (expect)", e.ResType.String(), e.ExpType.String())
}

type ErrNotEqual struct {
	R, E interface{}
}

func NewErrNotEqual(result, expect interface{}) *ErrNotEqual {
	return &ErrNotEqual{result, expect}
}

func (e *ErrNotEqual) Error() string {
	return fmt.Sprintf("values not equal: %v, %v", e.R, e.E)
}

type ErrFieldNotFound struct {
	Field string
}

func NewErrFieldNotFound(f string) *ErrFieldNotFound {
	return &ErrFieldNotFound{f}
}

func (e *ErrFieldNotFound) Error() string {
	return fmt.Sprintf("field (%s) not found", e.Field)
}
