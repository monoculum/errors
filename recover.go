package errors

import "runtime"

func Recover() []byte {
	stack := make([]byte, 1<<16)
	length := runtime.Stack(stack, false)
	return stack[:length]
}
