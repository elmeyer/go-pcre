//go:build windows && 386

package pcre

// #cgo !pcre_pkg_config LDFLAGS: ${SRCDIR}/libpcre_windows_386.a
// #cgo !pcre_pkg_config CFLAGS: -DPCRE_STATIC
import "C"
