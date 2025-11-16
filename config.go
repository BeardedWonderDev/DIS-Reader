package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/spf13/viper"
)

const (
	DefaultAppName    = "DIS Reader"
	DefaultAppVersion = "0.0.5"
	DefaultJavaPath   = "java"
	DefaultJDBCPort   = "8888"
	DefualtLogLevel   = slog.LevelInfo
	Prefix            = "disreader"
	ConfigFileName    = "disreader.yaml"
)

func NewConfig() *types.DISUIConfig {
	viper.SetConfigFile(ConfigFileName)
	viper.SetDefault("appName", DefaultAppName)
	viper.SetDefault("appVersion", DefaultAppVersion)
	viper.SetDefault("theme", types.DefaultThemeName)
	viper.SetDefault("disConfig.logLevel", DefualtLogLevel)
	viper.SetDefault("disConfig.jdbcConfig.javaPath", DefaultJavaPath)
	viper.SetDefault("disConfig.jdbcConfig.jdbcPort", DefaultJDBCPort)
	viper.SetDefault("debugSearch.defaultOutputMode", string(types.DebugSearchOutputSQLite))
	viper.SetDefault("debugSearch.defaultOutputPath", "")

	// If config file exists, use it
	_, err := os.ReadFile(ConfigFileName)
	if err == nil {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error while reading %s. Shutting down: %s", ConfigFileName, err)
		}
	} else {
		if os.IsNotExist(err) {
			log.Printf("Could not find %s. Attempting to use environment variables.\n", ConfigFileName)
		} else {
			log.Fatalf("Error while reading %s. Shutting down.", ConfigFileName)
		}
	}

	var config types.DISUIConfig
	// If available, use env vars for config
	for _, fieldName := range getFlattenedStructFields(reflect.TypeOf(config)) {
		envKey := strings.ToUpper(fmt.Sprintf("%s_%s", Prefix, strings.ReplaceAll(fieldName, ".", "_")))
		envVar := os.Getenv(envKey)
		if envVar != "" {
			viper.Set(fieldName, envVar)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalln("Error while creating config. Shutting down.")
	}

	return &config
}

func getFlattenedStructFields(t reflect.Type) []string {
	return getFlattenedStructFieldsHelper(t, []string{})
}

func getFlattenedStructFieldsHelper(t reflect.Type, prefixes []string) []string {
	unwrappedT := t
	if t.Kind() == reflect.Pointer {
		unwrappedT = t.Elem()
	}

	flattenedFields := make([]string, 0)
	for i := 0; i < unwrappedT.NumField(); i++ {
		field := unwrappedT.Field(i)
		fieldName := field.Tag.Get("mapstructure")
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Pointer:
			flattenedFields = append(flattenedFields, getFlattenedStructFieldsHelper(field.Type, append(prefixes, fieldName))...)
		default:
			flattenedField := fieldName
			if len(prefixes) > 0 {
				flattenedField = fmt.Sprintf("%s.%s", strings.Join(prefixes, "."), fieldName)
			}
			flattenedFields = append(flattenedFields, flattenedField)
		}
	}

	return flattenedFields
}
