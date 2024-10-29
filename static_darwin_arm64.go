//go:build darwin && arm64

package pcre

// #cgo !pcre_pkg_config LDFLAGS: ${SRCDIR}/libpcre_darwin_arm64.a
import "C"
