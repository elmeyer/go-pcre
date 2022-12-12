//go:build windows && amd64

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_windows_amd64.a
// #cgo CFLAGS: -DPCRE_STATIC
import "C"
