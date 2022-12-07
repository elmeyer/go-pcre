//go:build darwin && amd64

package pcre

// #cgo LDFLAGS: ${SRCDIR}/libpcre_darwin_amd64.a
import "C"
