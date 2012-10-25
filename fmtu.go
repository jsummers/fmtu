// ◄◄◄ fmtu.go ►►►
// 
// By Jason Summers, 2012.

package fmtu

import "fmt"
import "strings"
import "io"
import "os"
import "time"
import "errors"
import "reflect"

type segmentInfoType struct {
	start        int // Start of this segment (index into the original format string)
	end          int // 1 + the index of the last byte in this segment
	isFmtSegment bool
	verb         byte   // Valid if isFmtSegment = true
	formatted    string // This segment’s contents, after formatting
}

type ctxType struct {
	format string
	seg    []segmentInfoType
	args   []interface{}

	segCount    int
	curSegStart int
	inFmt       bool
}

func isVerb(x byte) bool {
	switch x {
	case 'b', 'c', 'd', 'e', 'E', 'f', 'g', 'G', 'o', 'p',
		'q', 's', 't', 'T', 'U', 'v', 'x', 'X':
		return true
	}
	return false
}

// Record the location of a format segment.
// i = position of the first byte after this segment.
func (ctx *ctxType) endFmtSegment(i int) {
	if ctx.segCount >= len(ctx.seg) {
		panic("too many format specifiers")
	}

	ctx.seg[ctx.segCount].verb = ctx.format[i-1]
	ctx.seg[ctx.segCount].start = ctx.curSegStart
	ctx.seg[ctx.segCount].end = i
	ctx.seg[ctx.segCount].isFmtSegment = true
	ctx.segCount++
}

// Record the location of a non-format segment.
// i = position of the first byte after this segment.
func (ctx *ctxType) endNonfmtSegment(i int) {
	if i <= ctx.curSegStart {
		return // Ignore zero-length segments
	}
	if ctx.inFmt {
		return
	}

	if ctx.segCount >= len(ctx.seg) {
		panic("too many format specifiers")
	}
	ctx.seg[ctx.segCount].start = ctx.curSegStart
	ctx.seg[ctx.segCount].end = i
	ctx.seg[ctx.segCount].isFmtSegment = false
	ctx.segCount++
}

// Split the string into segments.
// Each segment is either a format specifier, or other string data.
// A “format” segment consists of a format specifier, and uses one printf
// argument.
// A “nonformat” segment uses no format variables, though it may contain
// escaped percent signs (“%%”).
// The most segments there can be is 2×(number_of_args)+1.
func (ctx *ctxType) parseFormatString() {
	var i int

	ctx.seg = make([]segmentInfoType, 2*len(ctx.args)+1)

	ctx.segCount = 0
	ctx.curSegStart = 0
	ctx.inFmt = false

	// Scan through the format string, byte by byte.
	for i = 0; i < len(ctx.format); i++ {
		if ctx.inFmt && isVerb(ctx.format[i]) {
			// Found the end of a format specifier
			ctx.endFmtSegment(i + 1)
			// Record the start position of the next segment.
			ctx.inFmt = false
			ctx.curSegStart = i + 1
		} else if !ctx.inFmt && ctx.format[i] == '%' {
			if i+1 < len(ctx.format) && ctx.format[i+1] == '%' {
				// An escaped %
				i++ // Skip the second %
				continue
			} else {
				// Found the start of a format specifier
				ctx.endNonfmtSegment(i)
				ctx.inFmt = true
				ctx.curSegStart = i
			}
		}
	}

	// Handle anything after the last format specifier.
	ctx.endNonfmtSegment(len(ctx.format))

	ctx.seg = ctx.seg[:ctx.segCount] // Re-slice to set len(seg)
}

func fixupNumber(s string) string {
	if strings.Index(s, "-") < 0 {
		return s
	}
	return strings.Replace(s, "\u002d", "\u2212", -1)
}

func fixupQuoted(s string) string {
	var n int = len(s)
	if n < 2 {
		return s
	}
	if s[0] == '\'' && s[n-1] == '\'' {
		return "\u2018" + s[1:n-1] + "\u2019"
	}
	if s[0] == '"' && s[n-1] == '"' {
		return "\u201c" + s[1:n-1] + "\u201d"
	}
	return s
}

func fixupDuration(s string) string {
	if len(s) >= 2 && s[len(s)-2] == 'u' && s[len(s)-1] == 's' {
		s = strings.Replace(s, "us", "μs", 1)
	}
	// Durations can be negative, so fix any minus sign also.
	return fixupNumber(s)
}

func (ctx *ctxType) customFormat(segNum int, argNum int) string {
	// The format specifier:
	ufmt := ctx.format[ctx.seg[segNum].start:ctx.seg[segNum].end]
	// Format it using the standard Sprintf().
	formatted := fmt.Sprintf(ufmt, ctx.args[argNum:argNum+1]...)

	// Handle special types:
	switch ctx.args[argNum].(type) {
	case time.Duration:
		return fixupDuration(formatted)
	}

	// Handle %q:
	if ctx.seg[segNum].verb == 'q' {
		return fixupQuoted(formatted)
	}

	// Other signed numbers:
	switch reflect.TypeOf(ctx.args[argNum]).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return fixupNumber(formatted)
	}

	// If we’re not sure what to do, better not do anything.
	return formatted
}

// Format each segment individually.
// Put the formatted text in the ctx.seg[x].formatted field.
func (ctx *ctxType) applyFormats() {
	curArgNum := 0
	for i := range ctx.seg {
		if ctx.seg[i].isFmtSegment {
			if curArgNum >= len(ctx.args) {
				panic("too many format specifiers")
			}
			ctx.seg[i].formatted = ctx.customFormat(i, curArgNum)
			curArgNum++
		} else {
			// No true format specifiers, but use Sprintf because we may still
			// need to process “%%”.
			ctx.seg[i].formatted = fmt.Sprintf(ctx.format[ctx.seg[i].start:ctx.seg[i].end])
		}
	}

	if curArgNum < len(ctx.args) {
		panic("not enough format specifiers")
	}
}

func (ctx *ctxType) combineFormattedStrings() string {
	s := ""
	for i := range ctx.seg {
		s += ctx.seg[i].formatted
	}
	return s
}

// Fprintf is a Unicodized version of fmt.Fprintf.
func Fprintf(w io.Writer, format string, args ...interface{}) (n int, err error) {
	n, err = w.Write([]byte(Sprintf(format, args...)))
	return
}

// Printf is a Unicodized version of fmt.Printf.
func Printf(format string, args ...interface{}) (n int, err error) {
	n, err = Fprintf(os.Stdout, format, args...)
	return
}

// Errorf is a Unicodized version of fmt.Errorf.
func Errorf(format string, args ...interface{}) error {
	return errors.New(Sprintf(format, args...))
}

// Sprintf is a Unicodized version of fmt.Sprintf.
func Sprintf(format string, args ...interface{}) string {
	ctx := new(ctxType)
	ctx.format = format
	ctx.args = args
	ctx.parseFormatString()
	ctx.applyFormats()
	return ctx.combineFormattedStrings()
}
