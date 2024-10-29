//go:build darwin && amd64

package pcre

// #cgo !pcre_pkg_config LDFLAGS: ${SRCDIR}/libpcre_darwin_amd64.a
import "C"
