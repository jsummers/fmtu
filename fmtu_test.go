// ◄◄◄ fmtu_test.go ►►►
//
// By Jason Summers, 2012.
//
// Regression tests for package fmtu.

package fmtu

import "testing"
import "time"

func check(t *testing.T, name, e, g string) {
	if g != e {
		t.Logf("%s: expected ‘%s’ got ‘%s’\n", name, e, g)
		t.Fail()
	}
}

type myInt int

func TestOne(t *testing.T) {
	var g string

	g = Sprintf("%% %d %%", 1)
	check(t, "escape", "% 1 %", g)

	g = Sprintf("aa-a%06dbbb%dc", -23, 24)
	check(t, "int1", "aa-a−00023bbb24c", g)

	g = Sprintf("%.3f", -6.7)
	check(t, "float1", "−6.700", g)

	g = Sprintf(" %v ", -4.5e-12)
	check(t, "float2", " −4.5e−12 ", g)

	g = Sprintf("%.3f", complex128(-3.14-4i))
	check(t, "cmplx", "(−3.140−4.000i)", g)

	g = Sprintf("%q", "Test\n")
	check(t, "q1", "“Test\\n”", g)

	g = Sprintf("%q", "Te'\"st\n")
	check(t, "q2", "“Te'\\\"st\\n”", g)

	g = Sprintf(" %q ", 0x263a)
	check(t, "q3", " ‘☺’ ", g)

	g = Sprintf("%+q", 0x263a)
	check(t, "q4", "‘\\u263a’", g)

	g = Sprintf("%10q", "x")
	check(t, "q5", "       “x”", g)

	g = Sprintf("%-10q", "x")
	check(t, "q6", "“x”       ", g)

	g = Sprintf("%v", time.Duration(-25)*time.Microsecond)
	check(t, "duration", "−25μs", g)

	g = Sprintf("%s", "'-'")
	check(t, "s1", "'-'", g)

	g = Sprintf("%v", myInt(-17))
	check(t, "customtype", "−17", g)

	g = Sprintf("%d %*d %*.*f %d", 1, 3, -2, 10, 4, 1.234567, 5)
	check(t, "*", "1  −2     1.2346 5", g)
}
