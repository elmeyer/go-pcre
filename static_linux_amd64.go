//go:build linux && amd64

package pcre

// #cgo !pcre_pkg_config LDFLAGS: ${SRCDIR}/libpcre_linux_amd64.a
import "C"
