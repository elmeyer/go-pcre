// Copyright (c) 2011 Florian Weimer. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// * Redistributions of source code must retain the above copyright
//   notice, this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright
//   notice, this list of conditions and the following disclaimer in the
//   documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package pcre provides access to the Perl Compatible Regular
// Expresion library, PCRE.
//
// It implements two main types, Regexp and Matcher.  Regexp objects
// store a compiled regular expression. They consist of two immutable
// parts: pcre and pcre_extra. Compile()/MustCompile() initialize pcre.
// Calling Study() on a compiled Regexp initializes pcre_extra.
// Compilation of regular expressions using Compile or MustCompile is
// slightly expensive, so these objects should be kept and reused,
// instead of compiling them from scratch for each matching attempt.
// CompileJIT and MustCompileJIT are way more expensive, because they
// run Study() after compiling a Regexp, but they tend to give
// much better perfomance:
// http://sljit.sourceforge.net/regex_perf.html
//
// Matcher objects keeps the results of a match against a []byte or
// string subject.  The Group and GroupString functions provide access
// to capture groups; both versions work no matter if the subject was a
// []byte or string, but the version with the matching type is slightly
// more efficient.
//
// Matcher objects contain some temporary space and refer the original
// subject.  They are mutable and can be reused (using Match,
// MatchString, Reset or ResetString).
//
// For details on the regular expression language implemented by this
// package and the flags defined below, see the PCRE documentation.
// http://www.pcre.org/pcre.txt
package pcre

// #include <string.h>
// #include "./pcre.h"
// #include "./pcre_fallback.h"
// static inline void pcre_free_stub(void *re) {
//     pcre_free(re);
// }
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"unsafe"
)

// Flags for Compile and Match functions.
const (
	ANCHORED          = C.PCRE_ANCHORED
	BSR_ANYCRLF       = C.PCRE_BSR_ANYCRLF
	BSR_UNICODE       = C.PCRE_BSR_UNICODE
	NEWLINE_ANY       = C.PCRE_NEWLINE_ANY
	NEWLINE_ANYCRLF   = C.PCRE_NEWLINE_ANYCRLF
	NEWLINE_CR        = C.PCRE_NEWLINE_CR
	NEWLINE_CRLF      = C.PCRE_NEWLINE_CRLF
	NEWLINE_LF        = C.PCRE_NEWLINE_LF
	NO_START_OPTIMIZE = C.PCRE_NO_START_OPTIMIZE
	NO_UTF8_CHECK     = C.PCRE_NO_UTF8_CHECK
)

// Flags for Compile functions
const (
	CASELESS          = C.PCRE_CASELESS
	DOLLAR_ENDONLY    = C.PCRE_DOLLAR_ENDONLY
	DOTALL            = C.PCRE_DOTALL
	DUPNAMES          = C.PCRE_DUPNAMES
	EXTENDED          = C.PCRE_EXTENDED
	EXTRA             = C.PCRE_EXTRA
	FIRSTLINE         = C.PCRE_FIRSTLINE
	JAVASCRIPT_COMPAT = C.PCRE_JAVASCRIPT_COMPAT
	MULTILINE         = C.PCRE_MULTILINE
	NEVER_UTF         = C.PCRE_NEVER_UTF
	NO_AUTO_CAPTURE   = C.PCRE_NO_AUTO_CAPTURE
	UNGREEDY          = C.PCRE_UNGREEDY
	UTF8              = C.PCRE_UTF8
	UCP               = C.PCRE_UCP
)

// Flags for Match functions
const (
	NOTBOL           = C.PCRE_NOTBOL
	NOTEOL           = C.PCRE_NOTEOL
	NOTEMPTY         = C.PCRE_NOTEMPTY
	NOTEMPTY_ATSTART = C.PCRE_NOTEMPTY_ATSTART
	PARTIAL_HARD     = C.PCRE_PARTIAL_HARD
	PARTIAL_SOFT     = C.PCRE_PARTIAL_SOFT
)

// Flags for Study function
const (
	STUDY_JIT_COMPILE              = C.PCRE_STUDY_JIT_COMPILE
	STUDY_JIT_PARTIAL_SOFT_COMPILE = C.PCRE_STUDY_JIT_PARTIAL_SOFT_COMPILE
	STUDY_JIT_PARTIAL_HARD_COMPILE = C.PCRE_STUDY_JIT_PARTIAL_HARD_COMPILE
)

// Exec-time and get/set-time error codes
const (
	ERROR_NOMATCH        = C.PCRE_ERROR_NOMATCH
	ERROR_NULL           = C.PCRE_ERROR_NULL
	ERROR_BADOPTION      = C.PCRE_ERROR_BADOPTION
	ERROR_BADMAGIC       = C.PCRE_ERROR_BADMAGIC
	ERROR_UNKNOWN_OPCODE = C.PCRE_ERROR_UNKNOWN_OPCODE
	ERROR_UNKNOWN_NODE   = C.PCRE_ERROR_UNKNOWN_NODE
	ERROR_NOMEMORY       = C.PCRE_ERROR_NOMEMORY
	ERROR_NOSUBSTRING    = C.PCRE_ERROR_NOSUBSTRING
	ERROR_MATCHLIMIT     = C.PCRE_ERROR_MATCHLIMIT
	ERROR_CALLOUT        = C.PCRE_ERROR_CALLOUT
	ERROR_BADUTF8        = C.PCRE_ERROR_BADUTF8
	ERROR_BADUTF8_OFFSET = C.PCRE_ERROR_BADUTF8_OFFSET
	ERROR_PARTIAL        = C.PCRE_ERROR_PARTIAL
	ERROR_BADPARTIAL     = C.PCRE_ERROR_BADPARTIAL
	ERROR_RECURSIONLIMIT = C.PCRE_ERROR_RECURSIONLIMIT
	ERROR_INTERNAL       = C.PCRE_ERROR_INTERNAL
	ERROR_BADCOUNT       = C.PCRE_ERROR_BADCOUNT
	ERROR_JIT_STACKLIMIT = C.PCRE_ERROR_JIT_STACKLIMIT
)

// Regexp holds a reference to a compiled regular expression.
// Use Compile or MustCompile to create such objects.
// Use FreeRegexp to free memory when done with the struct.
type Regexp struct {
	ptr   *C.pcre
	extra *C.pcre_extra
}

// Number of bytes in the compiled pattern
func pcreSize(ptr *C.pcre) (size C.size_t) {
	C.pcre_fullinfo(ptr, nil, C.PCRE_INFO_SIZE, unsafe.Pointer(&size))
	return
}

// Number of capture groups
func pcreGroups(ptr *C.pcre) (count C.int) {
	C.pcre_fullinfo(ptr, nil,
		C.PCRE_INFO_CAPTURECOUNT, unsafe.Pointer(&count))
	return
}

// Free c allocated memory related to regexp.
func (re *Regexp) FreeRegexp() {
	// pcre_free is a function pointer, call a stub that calls it.
	if re.ptr != nil {
		C.pcre_free_stub(unsafe.Pointer(re.ptr))
		re.ptr = nil
	}
	if re.extra != nil {
		C.pcre_free_study(re.extra)
		re.extra = nil
	}
	runtime.SetFinalizer(re, nil)
}

// Compile the pattern and return a compiled regexp.
// If compilation fails, the second return value holds a *CompileError.
func Compile(pattern string, flags int) (re *Regexp, err error) {
	pattern1 := C.CString(pattern)
	defer C.free(unsafe.Pointer(pattern1))
	if clen := int(C.strlen(pattern1)); clen != len(pattern) {
		err = &CompileError{
			Pattern: pattern,
			Message: "NUL byte in pattern",
			Offset:  clen,
		}
		return
	}
	var errptr *C.char
	var erroffset C.int
	re = &Regexp{}
	re.ptr = C.pcre_compile(pattern1, C.int(flags), &errptr, &erroffset, nil)
	if re.ptr == nil {
		err = &CompileError{
			Pattern: pattern,
			Message: C.GoString(errptr),
			Offset:  int(erroffset),
		}
		return
	}
	runtime.SetFinalizer(re, (*Regexp).FreeRegexp)
	return
}

// CompileJIT is a combination of Compile and Study. It first compiles
// the pattern and if this succeeds calls Study on the compiled pattern.
// comFlags are Compile flags, jitFlags are study flags.
// If compilation fails, the second return value holds a *CompileError.
func CompileJIT(pattern string, comFlags, jitFlags int) (*Regexp, error) {
	re, err := Compile(pattern, comFlags)
	if err == nil {
		err = re.Study(jitFlags)
	}
	return re, err
}

// MustCompile compiles the pattern.  If compilation fails, panic.
func MustCompile(pattern string, flags int) (re *Regexp) {
	re, err := Compile(pattern, flags)
	if err != nil {
		panic(err)
	}
	return
}

// MustCompileJIT compiles and studies the pattern.  On failure it panics.
func MustCompileJIT(pattern string, comFlags, jitFlags int) (re *Regexp) {
	re, err := CompileJIT(pattern, comFlags, jitFlags)
	if err != nil {
		panic(err)
	}
	return
}

// Study adds Just-In-Time compilation to a Regexp. This may give a huge
// speed boost when matching. If an error occurs, return value is non-nil.
// Flags optionally specifies JIT compilation options for partial matches.
func (re *Regexp) Study(flags int) error {
	if re.extra != nil {
		return fmt.Errorf("Study: Regexp has already been optimized")
	}
	if flags == 0 {
		flags = STUDY_JIT_COMPILE
	}

	var err *C.char
	re.extra = C.pcre_study(re.ptr, C.int(flags), &err)
	if err != nil {
		return fmt.Errorf("%s", C.GoString(err))
	}
	if re.extra == nil {
		// Studying the pattern may not produce useful information.
		return nil
	}
	return nil
}

// Groups returns the number of capture groups in the compiled pattern.
func (re *Regexp) Groups() int {
	if re.ptr == nil {
		panic("Regexp.Groups: uninitialized")
	}
	out := int(pcreGroups(re.ptr))
	return out
}

// Matcher objects provide a place for storing match results.
// They can be created by the Matcher and MatcherString functions,
// or they can be initialized with Reset or ResetString.
type Matcher struct {
	re       *Regexp
	groups   int
	ovector  []C.int // scratch space for capture offsets
	matches  bool    // last match was successful
	partial  bool    // was the last match a partial match?
	subjects string  // one of these fields is set to record the subject,
	subjectb []byte  // so that Group/GroupString can return slices
	err      error
}

// NewMatcher creates a new matcher object for the given Regexp.
func (re *Regexp) NewMatcher() (m *Matcher) {
	m = new(Matcher)
	m.Init(re)
	return
}

// Matcher creates a new matcher object, with the byte slice as subject.
// It also starts a first match on subject. Test for success with Matches().
func (re *Regexp) Matcher(subject []byte, flags int) (m *Matcher) {
	m = re.NewMatcher()
	m.Match(subject, flags)
	return
}

// MatcherString creates a new matcher, with the specified subject string.
// It also starts a first match on subject. Test for success with Matches().
func (re *Regexp) MatcherString(subject string, flags int) (m *Matcher) {
	m = re.NewMatcher()
	m.MatchString(subject, flags)
	return
}

// Reset switches the matcher object to the specified regexp and subject.
// It also starts a first match on subject.
func (m *Matcher) Reset(re *Regexp, subject []byte, flags int) bool {
	m.Init(re)
	out := m.Match(subject, flags)
	return out
}

// ResetString switches the matcher object to the given regexp and subject.
// It also starts a first match on subject.
func (m *Matcher) ResetString(re *Regexp, subject string, flags int) bool {
	m.Init(re)
	out := m.MatchString(subject, flags)
	return out
}

// Init binds an existing Matcher object to the given Regexp.
func (m *Matcher) Init(re *Regexp) {
	if re.ptr == nil {
		panic("Matcher.Init: uninitialized")
	}
	m.matches = false
	m.err = nil
	if m.re != nil && m.re.ptr != nil && m.re.ptr == re.ptr {
		// Skip group count extraction if the matcher has
		// already been initialized with the same regular
		// expression.
		return
	}
	m.re = re
	m.groups = re.Groups()
	if ovectorlen := 3 * (1 + m.groups); len(m.ovector) < ovectorlen {
		m.ovector = make([]C.int, ovectorlen)
	}
}

// Err returns first error encountered by Matcher.
func (m *Matcher) Err() error {
	return m.err
}

var nullbyte = []byte{0}

// Match tries to match the specified byte slice to
// the current pattern by calling Exec and collects the result.
// Returns true if the match succeeds.
// Match is a no-op if err is not nil.
func (m *Matcher) Match(subject []byte, flags int) bool {
	if m.err != nil {
		return false
	}
	if m.re == nil || m.re.ptr == nil {
		panic("Matcher.Match: uninitialized")
	}
	rc := m.Exec(subject, flags)
	m.matches, m.err = matched(rc)
	m.partial = (rc == ERROR_PARTIAL)
	return m.matches
}

// MatchString tries to match the specified subject string to
// the current pattern by calling ExecString and collects the result.
// Returns true if the match succeeds.
func (m *Matcher) MatchString(subject string, flags int) bool {
	if m.err != nil {
		return false
	}
	if m.re == nil || m.re.ptr == nil {
		panic("Matcher.MatchString: uninitialized")
	}
	rc := m.ExecString(subject, flags)
	m.matches, m.err = matched(rc)
	m.partial = (rc == ERROR_PARTIAL)
	return m.matches
}

// Exec tries to match the specified byte slice to
// the current pattern. Returns the raw pcre_exec error code.
func (m *Matcher) Exec(subject []byte, flags int) int {
	if m.re == nil || m.re.ptr == nil {
		panic("Matcher.Exec: uninitialized")
	}
	length := len(subject)
	m.subjects = ""
	m.subjectb = subject
	if length == 0 {
		subject = nullbyte // make first character adressable
	}
	subjectptr := (*C.char)(unsafe.Pointer(&subject[0]))
	return m.exec(subjectptr, length, flags)
}

// ExecString tries to match the specified subject string to
// the current pattern. It returns the raw pcre_exec error code.
func (m *Matcher) ExecString(subject string, flags int) int {
	if m.re == nil || m.re.ptr == nil {
		panic("Matcher.ExecString: uninitialized")
	}
	length := len(subject)
	m.subjects = subject
	m.subjectb = nil
	if length == 0 {
		subject = "\000" // make first character addressable
	}
	// The following is a non-portable kludge to avoid a copy
	subjectptr := *(**C.char)(unsafe.Pointer(&subject))
	return m.exec(subjectptr, length, flags)
}

func (m *Matcher) exec(subjectptr *C.char, length, flags int) int {
	rc := C.pcre_exec(m.re.ptr, m.re.extra,
		subjectptr, C.int(length),
		0, C.int(flags), &m.ovector[0], C.int(len(m.ovector)))
	return int(rc)
}

// matched checks the return code of a pattern match for success.
func matched(rc int) (bool, error) {
	switch {
	case rc >= 0 || rc == C.PCRE_ERROR_PARTIAL:
		return true, nil
	case rc == C.PCRE_ERROR_NOMATCH:
		return false, nil
	case rc == C.PCRE_ERROR_BADOPTION:
		return false, errors.New("PCRE.Match: invalid option flag")
	}
	err := errors.New(
		"unexpected return code from pcre_exec: " + strconv.Itoa(rc),
	)
	return false, err
}

// Matches returns true if a previous call to Matcher, MatcherString, Reset,
// ResetString, Match or MatchString succeeded.
func (m *Matcher) Matches() bool {
	return m.matches
}

// Partial returns true if a previous call to Matcher, MatcherString, Reset,
// ResetString, Match or MatchString found a partial match.
func (m *Matcher) Partial() bool {
	return m.partial
}

// Groups returns the number of groups in the current pattern.
func (m *Matcher) Groups() int {
	return m.groups
}

// Present returns true if the numbered capture group is present in the last
// match (performed by Matcher, MatcherString, Reset, ResetString,
// Match, or MatchString).  Group numbers start at 1.  A capture group
// can be present and match the empty string.
func (m *Matcher) Present(group int) bool {
	return m.ovector[2*group] >= 0
}

// Group returns the numbered capture group of the last match (performed by
// Matcher, MatcherString, Reset, ResetString, Match, or MatchString).
// Group 0 is the part of the subject which matches the whole pattern;
// the first actual capture group is numbered 1.  Capture groups which
// are not present return a nil slice.
func (m *Matcher) Group(group int) []byte {
	start := m.ovector[2*group]
	end := m.ovector[2*group+1]
	if start >= 0 {
		if m.subjectb != nil {
			return m.subjectb[start:end]
		}
		return []byte(m.subjects[start:end])
	}
	return nil
}

// Extract returns a slice of byte slices for a single match.
// The first byte slice contains the complete match.
// Subsequent byte slices contain the captured groups.
// If there was no match then nil is returned.
func (m *Matcher) Extract() [][]byte {
	if !m.matches {
		return nil
	}
	extract := make([][]byte, m.groups+1)
	extract[0] = m.subjectb
	for i := 1; i <= m.groups; i++ {
		x0 := m.ovector[2*i]
		x1 := m.ovector[2*i+1]
		extract[i] = m.subjectb[x0:x1]
	}
	return extract
}

// ExtractString returns a slice of strings for a single match.
// The first string contains the complete match.
// Subsequent strings in the slice contain the captured groups.
// If there was no match then nil is returned.
func (m *Matcher) ExtractString() []string {
	if !m.matches {
		return nil
	}
	extract := make([]string, m.groups+1)
	extract[0] = m.subjects
	for i := 1; i <= m.groups; i++ {
		x0 := m.ovector[2*i]
		x1 := m.ovector[2*i+1]
		extract[i] = m.subjects[x0:x1]
	}
	return extract
}

// GroupIndices returns the numbered capture group positions of the last
// match (performed by Matcher, MatcherString, Reset, ResetString, Match,
// or MatchString). Group 0 is the part of the subject which matches
// the whole pattern; the first actual capture group is numbered 1.
// Capture groups which are not present return a nil slice.
func (m *Matcher) GroupIndices(group int) []int {
	start := m.ovector[2*group]
	end := m.ovector[2*group+1]
	if start >= 0 {
		return []int{int(start), int(end)}
	}
	return nil
}

// GroupString returns the numbered capture group as a string.  Group 0
// is the part of the subject which matches the whole pattern; the first
// actual capture group is numbered 1.  Capture groups which are not
// present return an empty string.
func (m *Matcher) GroupString(group int) string {
	start := m.ovector[2*group]
	end := m.ovector[2*group+1]
	if start >= 0 {
		if m.subjectb != nil {
			return string(m.subjectb[start:end])
		}
		return m.subjects[start:end]
	}
	return ""
}

// Index returns the start and end of the first match, if a previous
// call to Matcher, MatcherString, Reset, ResetString, Match or
// MatchString succeeded. loc[0] is the start and loc[1] is the end.
func (m *Matcher) Index() (loc []int) {
	if !m.matches {
		return nil
	}
	loc = []int{int(m.ovector[0]), int(m.ovector[1])}
	return
}

// name2index converts a group name to its group index number.
func (m *Matcher) name2index(name string) (int, error) {
	if m.re == nil || m.re.ptr == nil {
		return 0, fmt.Errorf("Matcher.Named: uninitialized")
	}
	name1 := C.CString(name)
	defer C.free(unsafe.Pointer(name1))
	group := int(C.pcre_get_stringnumber(m.re.ptr, name1))
	if group < 0 {
		return group, fmt.Errorf("Matcher.Named: unknown name: " + name)
	}
	return group, nil
}

// Named returns the value of the named capture group.
// This is a nil slice if the capture group is not present.
// If the name does not refer to a group then error is non-nil.
func (m *Matcher) Named(group string) ([]byte, error) {
	groupNum, err := m.name2index(group)
	if err != nil {
		return []byte{}, err
	}
	return m.Group(groupNum), nil
}

// NamedString returns the value of the named capture group,
// or an empty string if the capture group is not present.
// If the name does not refer to a group then error is non-nil.
func (m *Matcher) NamedString(group string) (string, error) {
	groupNum, err := m.name2index(group)
	if err != nil {
		return "", err
	}
	return m.GroupString(groupNum), nil
}

// NamedPresent returns true if the named capture group is present.
// If the name does not refer to a group then error is non-nil.
func (m *Matcher) NamedPresent(group string) (bool, error) {
	groupNum, err := m.name2index(group)
	if err != nil {
		return false, err
	}
	return m.Present(groupNum), nil
}

// FindIndex returns the start and end of the first match,
// or nil if no match.  loc[0] is the start and loc[1] is the end.
func (re *Regexp) FindIndex(bytes []byte, flags int) (loc []int) {
	m := re.Matcher(bytes, flags)
	if m.Matches() {
		loc = []int{int(m.ovector[0]), int(m.ovector[1])}
		return
	}
	return nil
}

// ReplaceAll returns a copy of a byte slice
// where all pattern matches are replaced by repl.
func (re *Regexp) ReplaceAll(bytes, repl []byte, flags int) ([]byte, error) {
	m := re.Matcher(bytes, flags)
	r := []byte{}
	for m.matches {
		r = append(append(r, bytes[:m.ovector[0]]...), repl...)
		bytes = bytes[m.ovector[1]:]
		m.Match(bytes, flags)
	}
	return append(r, bytes...), m.err
}

// ReplaceAllString is equivalent to ReplaceAll with string return type.
func (re *Regexp) ReplaceAllString(in, repl string, flags int) (string, error) {
	str, err := re.ReplaceAll([]byte(in), []byte(repl), flags)
	return string(str), err
}

// Match holds details about a single successful regex match.
type Match struct {
	Finding string // Text that was found.
	Loc     []int  // Index bounds for location of finding.
}

// FindAll finds all instances that match the regex.
func (re *Regexp) FindAll(subject string, flags int) ([]Match, error) {
	matches := make([]Match, 0)
	m := re.MatcherString(subject, flags)
	offset := 0
	for m.Matches() {
		leftIdx := int(m.ovector[0]) + offset
		rightIdx := int(m.ovector[1]) + offset
		matches = append(
			matches,
			Match{
				subject[leftIdx:rightIdx],
				[]int{leftIdx, rightIdx},
			},
		)
		offset += maxInt(1, int(m.ovector[1]))
		if offset < len(subject) {
			m.MatchString(subject[offset:], flags)
		} else {
			break
		}
	}
	return matches, m.err
}

// CompileError holds details about a compilation error,
// as returned by the Compile function.  The offset is
// the byte position in the pattern string at which the
// error was detected.
type CompileError struct {
	Pattern string // The failed pattern
	Message string // The error message
	Offset  int    // Byte position of error
}

// Error converts a compile error to a string
func (e *CompileError) Error() string {
	return e.Pattern + " (" + strconv.Itoa(e.Offset) + "): " + e.Message
}

func maxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
