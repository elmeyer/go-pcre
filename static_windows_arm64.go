//go:build windows && arm64

package pcre

// #cgo !pcre_pkg_config LDFLAGS: ${SRCDIR}/libpcre_windows_arm64.a
// #cgo !pcre_pkg_config CFLAGS: -DPCRE_STATIC
import "C"
