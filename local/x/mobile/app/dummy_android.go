//go:build android

package app

/*
void nm_check_marker() {}
*/
import "C"

func nmCheckMarker() {
	C.nm_check_marker()
}
