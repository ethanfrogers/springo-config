package pkg

import (
	"bytes"
	"html/template"
	"os"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var selfReferentalRegex = regexp.MustCompile(`\${([a-z][^:]*)}`)
var environmentVariableRegex = regexp.MustCompile(`\${([A-Z][^:]*)}`)
var environmentVariableWithDefaultRegex = regexp.MustCompile(`\${([A-Z][^:]*):(.+)?}`)
var selfReferentalWithDefaultRegex = regexp.MustCompile(`\${([a-z].*):([a-z].*)}`)

type WithFunc func() (string, interface{})

func WithEnvironmentVariables() func() (string, interface{}) {
	return func() (string, interface{}) {
		environments := map[string]string{}
		for _, pair := range os.Environ() {
			s := strings.Split(pair, "=")
			environments[s[0]] = s[1]
		}
		return "Env", environments
	}
}

// ParseAndEvaluateYAML accepts a byte array that looks like YAML and evaluates
// it against itself
func ParseAndEvaluateYAML(b []byte, withFuncs ...WithFunc) ([]byte, error) {
	var decoded interface{}
	err := yaml.Unmarshal(b, &decoded)
	if err != nil {
		return nil, err
	}

	replaced := string(b)
	// ORDER MATTERS HERE!!!!
	replaced = selfReferentalRegex.ReplaceAllString(replaced, "${ .data.$1 }")
	replaced = environmentVariableRegex.ReplaceAllString(replaced, `${ .Env.$1 }`)
	replaced = selfReferentalWithDefaultRegex.ReplaceAllString(replaced, `${ or .data.$1 "$2"}`)
	replaced = environmentVariableWithDefaultRegex.ReplaceAllString(replaced, `${ or .Env.$1 "$2" }`)

	parsed, err := template.New("config").Delims("${", "}").Parse(replaced)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{"data": decoded}

	for _, f := range withFuncs {
		k, v := f()
		data[k] = v
	}

	var buf bytes.Buffer
	err = parsed.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	bufToString := buf.String()

	if strings.Contains(bufToString, "${") {
		return ParseAndEvaluateYAML([]byte(bufToString))
	}

	return []byte(buf.String()), nil
}
