package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"

	internal "github.com/BeardedWonderDev/DIS-Reader/internal"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Minimal bridge-server for hosting AgentService and health/metrics.
// Expects DISREADER_BRIDGE_MODE=remote and bridge credentials configured.
func main() {
	cfg := loadConfig()

	svc, err := internal.NewDISReaderService(cfg.DIS, nil)
	if err != nil {
		log.Fatalf("init service: %v", err)
	}
	defer svc.Shutdown()

	if svc.BridgeServer() == nil {
		log.Fatalf("bridge mode disabled; set bridge.mode=remote")
	}

	grpcPort := envDefault("DISREADER_BRIDGE_PORT", "8443")
	httpPort := envDefault("DISREADER_BRIDGE_HTTP_PORT", "8080")

	// gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("listen gRPC: %v", err)
		}
		srv := grpc.NewServer()
		svc.RegisterBridge(srv)
		log.Printf("bridge gRPC listening on :%s", grpcPort)
		log.Fatal(srv.Serve(lis))
	}()

	// HTTP health/metrics
	mux := http.NewServeMux()
	svc.RegisterHealth(mux)
	log.Printf("bridge HTTP health/metrics on :%s", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, mux))
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Minimal config loader (subset of main NewConfig) to avoid importing package main.
func loadConfig() *types.DISUIConfig {
	viper.SetConfigFile("disreader.yaml")
	viper.SetDefault("appName", "DIS Reader")
	viper.SetDefault("appVersion", "0.1.0")
	viper.SetDefault("theme", types.DefaultThemeName)
	viper.SetDefault("disConfig.logLevel", 0)
	viper.SetDefault("disConfig.jdbcConfig.javaPath", "java")
	viper.SetDefault("disConfig.jdbcConfig.jdbcPort", "8888")
	viper.SetDefault("bridge.mode", "embedded")
	viper.SetDefault("bridge.tls.insecureSkipVerify", false)

	_, err := os.ReadFile("disreader.yaml")
	if err == nil {
		_ = viper.ReadInConfig()
	}

	var config types.DISUIConfig
	for _, fieldName := range getFlattenedStructFields(reflect.TypeOf(config)) {
		envKey := strings.ToUpper("DISREADER_" + strings.ReplaceAll(fieldName, ".", "_"))
		if val := os.Getenv(envKey); val != "" {
			viper.Set(fieldName, val)
		}
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("config error: %v", err)
	}
	if config.Bridge == nil {
		config.Bridge = &types.BridgeConfig{Mode: "embedded"}
	}
	if config.DIS != nil {
		config.DIS.Bridge = config.Bridge
	}
	return &config
}

func getFlattenedStructFields(t reflect.Type) []string {
	return getFlattenedStructFieldsHelper(t, []string{})
}

func getFlattenedStructFieldsHelper(t reflect.Type, prefixes []string) []string {
	unwrapped := t
	if t.Kind() == reflect.Pointer {
		unwrapped = t.Elem()
	}
	fields := []string{}
	for i := 0; i < unwrapped.NumField(); i++ {
		f := unwrapped.Field(i)
		name := f.Tag.Get("mapstructure")
		switch f.Type.Kind() {
		case reflect.Struct, reflect.Pointer:
			fields = append(fields, getFlattenedStructFieldsHelper(f.Type, append(prefixes, name))...)
		default:
			flatten := name
			if len(prefixes) > 0 {
				flatten = strings.Join(append(prefixes, name), ".")
			}
			fields = append(fields, flatten)
		}
	}
	return fields
}
