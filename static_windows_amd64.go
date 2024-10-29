//go:build windows && amd64

package pcre

// #cgo !pcre_pkg_config LDFLAGS: ${SRCDIR}/libpcre_windows_amd64.a
// #cgo !pcre_pkg_config CFLAGS: -DPCRE_STATIC
import "C"
