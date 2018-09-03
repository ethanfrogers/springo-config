package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/imdario/mergo"

	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

type Logger interface {
	Info(s string)
	InfoF(s string, params ...interface{})
	Debug(s string)
	DebugF(s string, params ...interface{})
}

type Config struct {
	applications []string
	profiles     []string
	baseDir      string
	configs      map[string][]byte
	priority     []string
	debug        bool
	logger       Logger
}

func NewConfig() *Config {
	return &Config{
		applications: []string{"spinnaker"},
		profiles:     []string{"local"},
		baseDir:      path.Join(os.Getenv("HOME"), ".spinnaker"),
		configs:      map[string][]byte{},
	}
}

func (c *Config) WithApplications(applications ...string) *Config {
	c.applications = applications
	return c
}

func (c *Config) WithProfiles(profiles ...string) *Config {
	c.profiles = profiles
	return c
}

func (c *Config) WithBaseDir(baseDir string) *Config {
	c.baseDir = baseDir
	return c
}

func (c *Config) Debug(enabled bool) *Config {
	c.debug = enabled
	return c
}

func (c *Config) WithLogger(l Logger) *Config {
	c.logger = l
	return c
}

func (c *Config) Load(withFuncs ...WithFunc) error {

	// keeping 2 copies of the slice is a silly workaround because
	// when we reverse the priority files gets reversed as well
	// because memory
	// TODO(ethanfrogers): rename this to something not dumb
	var files = []string{}
	var priority = []string{}
	for _, application := range c.applications {
		files = append(files, application)
		priority = append(priority, application)
		for _, profile := range c.profiles {
			files = append(files, application+"-"+profile)
			priority = append(priority, application+"-"+profile)
		}
	}

	c.priority = priority

	// reverses the list of strings
	for i := len(c.priority)/2 - 1; i >= 0; i-- {
		opp := len(c.priority) - 1 - i
		c.priority[i], c.priority[opp] = c.priority[opp], c.priority[i]
	}

	var previous map[string]interface{}
	for _, f := range files {
		if c.debug {
			c.logger.DebugF("evaluating %s.yml\n", f)
		}
		d, err := readFileIfExists(filepath.Join(c.baseDir, f+".yml"))
		if err != nil {
			// return err
			// fmt.Println(err.Error())
			if c.debug {
				c.logger.Debug(err.Error())
			}
			continue
		}

		parsed, err := ParseAndEvaluateYAML(d, previous, withFuncs...)
		if err != nil {
			return err
		}
		c.configs[f] = parsed

		var decoded map[string]interface{}
		err = yaml.Unmarshal(parsed, &decoded)
		if err != nil {
			return err
		}

		if err := mergo.Merge(&previous, decoded, mergo.WithOverride); err != nil {
			return err
		}

	}
	return nil
}

func (c *Config) Get(property string) interface{} {
	var target interface{}
	for _, candidate := range c.priority {
		v := viper.New()
		v.SetConfigType("yaml")
		v.ReadConfig(bytes.NewReader(c.configs[candidate]))
		i := v.Get(property)
		if i != nil {
			target = i
			break
		}
	}
	return target
}

func readFileIfExists(pth string) ([]byte, error) {
	if _, err := os.Stat(pth); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("file %s not found.", pth))
	}

	d, err := ioutil.ReadFile(pth)
	if err != nil {
		return nil, err
	}

	return d, err
}
