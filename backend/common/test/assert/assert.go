package assert

import (
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var cmpOpts sync.Map
var errorCompareFn = func(e1, e2 error) bool {
	if e1 == nil || e2 == nil {
		return e1 == nil && e2 == nil
	}

	return e1.Error() == e2.Error()
}
var errorCompareOpts = cmp.Options{cmp.Comparer(errorCompareFn)}

func RegisterOpts(t reflect.Type, opts ...cmp.Option) {
	cmpOpts.Store(t, cmp.Options(opts))
}

func Equal(t *testing.T, actual, expected interface{}, opts ...cmp.Option) {
	t.Helper()
	Equalf(t,
		actual, expected,
		opts,
		"expected args to equal (expected, actual)\n  %T{%+v}\n  %T{%+v}",
		expected, expected,
		actual, actual,
	)
}

func Equalf(t *testing.T, actual, expected interface{}, opts []cmp.Option, msg string, args ...interface{}) {
	t.Helper()
	test(t, cmp.Equal(expected, actual, getCmpOpts(expected, opts)...), msg, args...)
}

func NotEqual(t *testing.T, actual, expected interface{}, opts ...cmp.Option) {
	t.Helper()
	NotEqualf(t,
		actual, expected,
		opts,
		"expected args to not equal\n  %T{%+v}",
		expected, expected,
	)
}

func NotEqualf(t *testing.T, actual, expected interface{}, opts []cmp.Option, msg string, args ...interface{}) {
	t.Helper()
	test(t, !cmp.Equal(expected, actual, getCmpOpts(expected, opts)...), msg, args...)
}

func Nil(t *testing.T, val interface{}) {
	t.Helper()
	Nilf(t, val, "expected <nil> value, instead got: %+v", val)
}

func Nilf(t *testing.T, val interface{}, msg string, args ...interface{}) {
	t.Helper()
	test(t, val == nil, msg, args...)
}

func NotNil(t *testing.T, val interface{}) {
	t.Helper()
	NotNilf(t, val, "expected non-nil value, instead got: <nil>")
}

func NotNilf(t *testing.T, val interface{}, msg string, args ...interface{}) {
	t.Helper()
	test(t, val != nil, msg, args...)
}

func False(t *testing.T, val bool) {
	t.Helper()
	Falsef(t, val, "expected <false> value, instead got: <true>")
}

func Falsef(t *testing.T, val bool, msg string, args ...interface{}) {
	t.Helper()
	test(t, !val, msg, args...)
}

func True(t *testing.T, val bool) {
	t.Helper()
	Truef(t, val, "expected <true> value, instead got: <false>")
}

func Truef(t *testing.T, val bool, msg string, args ...interface{}) {
	t.Helper()
	test(t, val, msg, args...)
}

func Fail(t *testing.T) {
	t.Helper()
	Failf(t, "asserted test failure here")
}

func Failf(t *testing.T, msg string, args ...interface{}) {
	t.Helper()
	test(t, false, msg, args...)
}

func test(t *testing.T, ok bool, msg string, args ...interface{}) {
	t.Helper()
	if !ok {
		t.Fatalf("failed test:\n"+msg, args...)
	}
}

func getCmpOpts(o interface{}, opts []cmp.Option) cmp.Options {
	var typeOpts cmp.Options
	if tos, ok := cmpOpts.Load(reflect.TypeOf(o)); ok {
		typeOpts = tos.(cmp.Options)
	} else if _, ok := o.(error); ok {
		typeOpts = append(typeOpts, errorCompareOpts...)
	}

	out := make(cmp.Options, 0, len(typeOpts)+len(opts))
	out = append(out, typeOpts...)
	out = append(out, opts...)

	return out
}
