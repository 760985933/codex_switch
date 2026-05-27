//go:build darwin

package main

/*
*/
import "C"

//export goTrayOnClick
func goTrayOnClick(tag C.int) {
	handleTrayClick(int(tag))
}
