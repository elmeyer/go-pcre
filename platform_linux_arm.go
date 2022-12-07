//go:build linux && arm

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_linux_arm.a
import "C"
