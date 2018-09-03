package pkg

import (
	"testing"
)

var baseCase = `
foo:
  bar: foobar
foo1:
  bar1: ${foo.bar}
`

var baseResult = `
foo:
  bar: foobar
foo1:
  bar1: foobar
`

var withEnvironmentVariablesCase = `
foo:
  bar: ${TEST_CASE:tacobell}
`

var withEnvironmentVariablesResult = `
foo:
  bar: HELLOWORLD
`

var selfReferentalWithDefaultBase = `
services:
  taco: bell
foo:
  bar: ${services.test:false}
`

var selfReferentalWithDefaultResult = `
services:
  taco: bell
foo:
  bar: false
`

var environmentVariableWithUrlBase = `
foo:
  bar: ${TEST_ME:http://test.io}
`

var environmentVariableWithUrlResult = `
foo:
  bar: http://test.io
`

var onlyEnvironmentVariableBase = `
foo:
  bar: ${TEST_CASE}
`

var onlyEnvionmentVariableResult = `
foo:
  bar: HELLOWORLD
`

func TestParseAndEvaluateYAML(t *testing.T) {
	cases := []struct {
		test       string
		expected   string
		withFunc   []WithFunc
		desciption string
	}{
		{
			test:       baseCase,
			expected:   baseResult,
			desciption: "base case",
		},
		{
			desciption: "case with environment variables",
			test:       withEnvironmentVariablesCase,
			expected:   withEnvironmentVariablesResult,
			withFunc: []WithFunc{func() (string, interface{}) {
				return "Env", map[string]string{"TEST_CASE": "HELLOWORLD"}
			}},
		},
		{
			desciption: "self referential with default",
			test:       selfReferentalWithDefaultBase,
			expected:   selfReferentalWithDefaultResult,
		},
		{
			desciption: "environment variable with url default",
			test:       environmentVariableWithUrlBase,
			expected:   environmentVariableWithUrlResult,
		},
		{
			desciption: "only environment variable",
			test:       onlyEnvironmentVariableBase,
			expected:   onlyEnvionmentVariableResult,
			withFunc: []WithFunc{func() (string, interface{}) {
				return "Env", map[string]string{"TEST_CASE": "HELLOWORLD"}
			}},
		},
	}

	for _, c := range cases {
		r, err := ParseAndEvaluateYAML([]byte(c.test), c.withFunc...)
		if err != nil {
			t.Fatalf("case: %s, err: %s", c.desciption, err.Error())
		}

		if string(r) != c.expected {
			t.Fatalf("expected: %s\ngot: %s\n", c.expected, r)
		}
	}
}
