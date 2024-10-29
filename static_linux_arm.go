//go:build linux && arm

package pcre

// #cgo !pcre_pkg_config LDFLAGS: ${SRCDIR}/libpcre_linux_arm.a
import "C"
