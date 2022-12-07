//go:build linux && arm64

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_linux_arm64.a
import "C"
