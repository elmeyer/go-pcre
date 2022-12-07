//go:build linux && amd64

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_linux_amd64.a
import "C"
