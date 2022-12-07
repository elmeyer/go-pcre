//go:build darwin && arm64

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_darwin_arm64.a
import "C"
