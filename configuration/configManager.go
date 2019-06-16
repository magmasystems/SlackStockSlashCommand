package configuration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// ConfigManagerOps - the operations that the COnfigManager can perform
type ConfigManagerOps interface {
	Config() *AppSettings
}

// ConfigManager - handles configuration
type ConfigManager struct {
}

// AppSettings - config settings for the bot
type AppSettings struct {
	Driver      string
	APIKeys     map[string]string
	SlackSecret string
	Webhook     string
	DMWebhook   string
	Port        int
	Database    struct {
		User     string
		Password string
		Host     string
		Port     int
		DbName   string
		SSL      bool
	}
	QuoteCheckInterval int
}

// Config - get the config settings from appSettings.json
func (mgr *ConfigManager) Config() *AppSettings {
	appSettingsFileName := "appSettings.json"

	// See if there is an "env" argument on the command line. If so, it can point to another config file.
	env := getArg("env", "")
	if env != "" {
		s := fmt.Sprintf("appSettings.%s.json", env)
		// see if this config file exists
		if _, err := os.Stat(s); err == nil {
			appSettingsFileName = s
		}
	}

	bytes, err := ioutil.ReadFile(appSettingsFileName)
	if err != nil {
		log.Fatalf("cannot find the %s file", appSettingsFileName)
	}

	settings := new(AppSettings)
	json.Unmarshal(bytes, &settings)

	// In case it's omitted, we need a sensible value on the Quote Checking Interval
	if settings.QuoteCheckInterval == 0 {
		settings.QuoteCheckInterval = 5
	}

	return settings
}

func getArg(key string, defautVal string) string {
	var val string

	for idx, arg := range os.Args {
		switch strings.ToLower(arg) {
		case key:
			val = os.Args[idx+1]
			break
		}
	}

	if len(val) == 0 {
		val = defautVal
	}

	return val
}
