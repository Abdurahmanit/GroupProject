package repository

import "errors"

var (
	ErrNotFound         = errors.New("entity not found")
	ErrUpdateFailed     = errors.New("update failed")
	ErrDeleteFailed     = errors.New("delete failed")
	ErrAlreadyExists    = errors.New("entity already exists")
	ErrOptimisticLock   = errors.New("optimistic lock conflict: data was modified by another process")
	ErrConnectionFailed = errors.New("database connection failed")
	ErrQueryFailed      = errors.New("database query failed")
)
