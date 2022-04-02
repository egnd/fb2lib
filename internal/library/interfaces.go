package library

type IUnmarshaler func(data []byte, v any) error
