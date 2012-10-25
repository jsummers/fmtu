// ◄◄◄ doc.go ►►►
// 
// By Jason Summers, 2012.

/*
Package fmtu performs string formatting that’s not afraid of Unicode.

This package is only half serious.

Consider this line of code:

    fmt.Printf("%d\n", 3-4)

Which prints:

    -1

Look at the “-” symbol in front of the 1. Technically, it’s not a true minus
sign; it’s a U+002D HYPHEN-MINUS character.

In proportional fonts, like on most web pages, HYPHEN-MINUS is usually very
short, like a hyphen. That’s pretty bad.

In monospace fonts, its length is usually between that of a hyphen and a minus
sign. That’s not as bad, but we can do better.

If you instead use fmtu to do your formatting:

    fmtu.Printf("%d\n", 3-4)

You get:

    −1

Yay, it’s a genuine U+2212 MINUS SIGN, not a plastic ASCII imitation.

Granted, you could have just done something like this instead:

    fmt.Printf("%s\n", strings.Replace(fmt.Sprintf("%d", 3-4),
        "-", "−", -1))

(And that’s more or less what fmtu does behind the scenes. It’s pretty
low-tech.)

Another problem is the %q format. It generates quotation marks, but it uses
the ASCII-compatible U+0022 QUOTATION MARK or U+0027 APOSTROPHE character.
fmtu will improve upon that, because it knows about U+201C LEFT DOUBLE
QUOTATION MARK, U+201D RIGHT DOUBLE QUOTATION MARK, U+2018 LEFT SINGLE
QUOTATION MARK, and U+2019 RIGHT SINGLE QUOTATION MARK.

fmtu also knows about the time.Duration type, and will fix its “us”
suffix to be “μs”.

Note: If your format string is invalid or not compatible with the arguments,
the behavior is undefined. It may cause a run-time panic, or it may not do
what you expect. This is different from the fmt package.

Warning for Windows users: As of Go version 1.0.3, it is difficult for a Go
program to print Unicode characters to a Windows console. You may get garbage
characters instead. That isn’t fmtu’s fault (it’s a limitation of Go’s io
library), and it will probably be fixed in a future version of Go.
*/
package fmtu
