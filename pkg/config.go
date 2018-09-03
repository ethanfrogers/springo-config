package pkg

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	applications []string
	profiles     []string
	baseDir      string
	configs      map[string][]byte
	priority     []string
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

func (c *Config) Load(withFuncs ...WithFunc) error {

	// TODO(ethanfrogers): rename this to something not dumb
	var files []string
	for _, application := range c.applications {
		files = append(files, application)
		for _, profile := range c.profiles {
			files = append(files, application+"-"+profile)
		}
	}

	c.priority = files
	// reverses the list of strings
	for i := len(c.priority)/2 - 1; i >= 0; i-- {
		opp := len(c.priority) - 1 - i
		c.priority[i], c.priority[opp] = c.priority[opp], c.priority[i]
	}

	for _, f := range files {
		d, err := readFileIfExists(filepath.Join(c.baseDir, f+".yml"))
		if err != nil {
			// return err
			// fmt.Println(err.Error())
			continue
		}
		parsed, err := ParseAndEvaluateYAML(d, withFuncs...)
		if err != nil {
			return err
		}
		c.configs[f] = parsed
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
		return nil, errors.New("config file not found")
	}

	d, err := ioutil.ReadFile(pth)
	if err != nil {
		return nil, err
	}

	return d, err
}
