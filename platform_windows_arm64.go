//go:build windows && arm64

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_windows_arm64.a
import "C"
