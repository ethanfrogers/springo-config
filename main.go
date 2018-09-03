package main

import (
	"fmt"
	"os"
	"path"

	"github.com/ethanfrogers/springo-config/pkg"
	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

type DebugLogger struct {
	pkg.Logger
}

func (d DebugLogger) Debug(s string) {
	fmt.Println(s)
}

func (d DebugLogger) DebugF(s string, params ...interface{}) {
	fmt.Printf(s, params...)
}

func main() {
	flag.StringSlice("application", []string{"spinnaker"}, "spring application")
	flag.StringSlice("profiles", []string{"local"}, "spring profiles")
	flag.String("config-dir", path.Join(os.Getenv("HOME"), ".spinnaker"), "config home dir")
	flag.String("path", "", "desired property")
	flag.Bool("debug", false, "enable debug logging")
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	logger := DebugLogger{}

	cfg := pkg.NewConfig().
		WithApplications(viper.GetStringSlice("application")...).
		WithProfiles(viper.GetStringSlice("profiles")...).
		Debug(viper.GetBool("debug")).
		WithLogger(logger)

	err := cfg.Load(pkg.WithEnvironmentVariables())
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg.Get(viper.GetString("path")))
}
