// Package interfaces contains interfaces for wpool package.
package interfaces

// Task is a task interface.
type Task interface {
	GetID() string
	Do()
}
