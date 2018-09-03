package pkg

import (
	"bytes"
	"html/template"
	"os"
	"regexp"
	"strings"

	"github.com/imdario/mergo"

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
func ParseAndEvaluateYAML(b []byte, context map[string]interface{}, withFuncs ...WithFunc) ([]byte, error) {
	var decoded map[string]interface{}
	err := yaml.Unmarshal(b, &decoded)
	if err != nil {
		return nil, err
	}

	replaced := string(b)
	// ORDER MATTERS HERE!!!!
	replaced = selfReferentalRegex.ReplaceAllString(replaced, "${ .context.$1 }")
	replaced = environmentVariableRegex.ReplaceAllString(replaced, `${ .Env.$1 }`)
	replaced = selfReferentalWithDefaultRegex.ReplaceAllString(replaced, `${ or .context.$1 "$2"}`)
	replaced = environmentVariableWithDefaultRegex.ReplaceAllString(replaced, `${ or .Env.$1 "$2" }`)

	// if we pass some context along, merge it with the decoded file
	// this handles parent evaluations
	if context != nil {
		if err := mergo.Merge(&decoded, context, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	parsed, err := template.New("config").Delims("${", "}").Parse(replaced)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{"context": decoded}

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
		return ParseAndEvaluateYAML([]byte(bufToString), decoded, withFuncs...)
	}

	return []byte(buf.String()), nil
}
