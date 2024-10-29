//go:build linux && arm64

package pcre

// #cgo !pcre_pkg_config LDFLAGS: ${SRCDIR}/libpcre_linux_arm64.a
import "C"
