package boondoggle

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

type TestSet struct {
	TestName          string
	Environment       string
	SetStateAll       string
	ServiceState      []string
	ExtraEnv          map[string]string
	Namespace         string
	Release           string
	UseSecrets        bool
	TLS               bool
	TillerNamespace   string
	ExpectInResult    []string
	NotExpectInResult []string
}

var tests = []TestSet{
	{
		TestName:    "Test Secrets",
		Environment: "dev",
		Namespace:   "mynamespace",
		Release:     "testrelease",
		UseSecrets:  true,
		ExpectInResult: []string{
			"helm secrets",
		},
	},
	{
		TestName:    "Test No Secrets",
		Environment: "dev",
		Namespace:   "mynamespace",
		Release:     "testrelease",
		UseSecrets:  false,
		NotExpectInResult: []string{
			"secrets",
		},
	},
	{
		TestName:    "Test Service State",
		Environment: "dev",
		ServiceState: []string{
			"service2=local",
		},
		Namespace:  "mynamespace",
		Release:    "testrelease",
		UseSecrets: false,
		ExpectInResult: []string{
			"--set-string alias-service2.localdev=true",
		},
	},
	{
		TestName:    "Test Set State all",
		Environment: "dev",
		Namespace:   "mynamespace",
		Release:     "testrelease",
		UseSecrets:  false,
		ExpectInResult: []string{
			"--set service1-chart.boondoggleCacheBust",
			"--set alias-service2.boondoggleCacheBust",
		},
		SetStateAll: "local",
	},
	{
		TestName:  "Test Extra Env",
		Namespace: "mynamespace",
		Release:   "prodrelease",
		ExtraEnv: map[string]string{
			"FOO":        "bar",
			"WITHCOMMAS": "stuff/, stuff2",
		},
		UseSecrets: false,
		ExpectInResult: []string{
			"--set-string global.myglobalvalue=bar",
			"--set-string global.withCommas=stuff/, stuff2",
		},
		NotExpectInResult: []string{
			"--set alias-service2.boondoggleCacheBust",
		},
	},
	{
		TestName:  "Test TLS",
		Namespace: "mynamespace",
		Release:   "prodrelease",
		TLS:       true,
		ExpectInResult: []string{
			"--tls",
		},
	},
	{
		TestName:        "Test Tiller Namespace",
		Namespace:       "mynamespace",
		Release:         "prodrelease",
		TillerNamespace: "tiller-namespace",
		ExpectInResult: []string{
			"--tiller-namespace tiller-namespace",
		},
		NotExpectInResult: []string{
			"--tls",
		},
	},
	{
		TestName:        "Test No Tiller Namespace",
		Namespace:       "mynamespace",
		Release:         "prodrelease",
		TillerNamespace: "kube-system",
		NotExpectInResult: []string{
			"--tiller-namespace",
		},
	},
}

func TestUpgradeCommandBuilder(t *testing.T) {
	viper.SetConfigFile("../example/boondoggle.yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}
	var config RawBoondoggle
	viper.Unmarshal(&config)
	for _, value := range tests {
		b := NewBoondoggle(config, value.Environment, value.SetStateAll, value.ServiceState, value.ExtraEnv, log.New(os.Stdout, "", 0))
		out, _ := b.DoUpgrade(value.Namespace, value.Release, true, value.UseSecrets, value.TLS, value.TillerNamespace, false)
		for _, expected := range value.ExpectInResult {
			if strings.Contains(string(out), expected) == false {
				t.Error(
					"\n For the test:", value.TestName, "\n",
					"Expected to find:", expected, "\n",
					"Got:", string(out), "\n",
				)
			}
		}

		for _, notExpected := range value.NotExpectInResult {
			if strings.Contains(string(out), notExpected) {
				t.Error(
					"\n For the test:", value.TestName, "\n",
					"Did not expect to find:", notExpected, "\n",
					"Got:", string(out), "\n",
				)
			}
		}
	}
}
