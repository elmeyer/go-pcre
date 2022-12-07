//go:build windows && 386

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_windows_386.a
import "C"
