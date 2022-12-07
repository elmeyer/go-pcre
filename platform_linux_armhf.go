//go:build linux && arm

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_linux_armhf.a
import "C"
