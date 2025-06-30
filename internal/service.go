package internal

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

//go:embed JDBCRunner.class
var jdbcRunnerClass []byte

//go:embed jt400-21.0.4.jar
var jt400Jar []byte

type DISReaderService struct {
	config *types.Config
	db     *database.JDBCRunnerService
	logger *slog.Logger
}

func NewDISReaderService(config *types.Config, logger *slog.Logger) *DISReaderService {
	if logger == nil {
		logger = slog.Default()
	}

	// Write embedded assets to temp directory
	tmp, err := os.MkdirTemp("", "disreader-jdbc-*")
	if err != nil {
		logger.Error("Failed to create temp directory", "error", err)
	} else {
		// write JAR
		jarPath := filepath.Join(tmp, "jt400-21.0.4.jar")
		if err := os.WriteFile(jarPath, jt400Jar, 0644); err != nil {
			logger.Error("Failed to write jt400 jar", "error", err)
		}
		// write class into classes/ directory
		classDir := filepath.Join(tmp, "classes")
		if err := os.Mkdir(classDir, 0755); err != nil {
			logger.Error("Failed to create class dir", "error", err)
		}
		classFile := filepath.Join(classDir, "JDBCRunner.class")
		if err := os.WriteFile(classFile, jdbcRunnerClass, 0644); err != nil {
			logger.Error("Failed to write JDBCRunner.class", "error", err)
		}
		// update config
		config.JarPath = jarPath
		config.ClassDir = classDir
	}

	db := database.NewJDBCRunnerService(config, logger)
	s := &DISReaderService{config: config, db: db, logger: logger}
	// Keep s.tempDir = tmp if needed (if field exists)
	// s.tempDir = tmp
	return s
}

func (s DISReaderService) GetConfig() *types.Config {
	return s.config
}

// TestDISConnection tests the JDBC connection using the embedded Java class.
func (s DISReaderService) TestDISConnection() error {
	if err := s.db.Start(); err != nil {
		return fmt.Errorf("failed to start JDBC server: %w", err)
	}
	return s.db.Connect()
}

func (s *DISReaderService) Shutdown() {
	s.logger.Info("Shutting down DISReaderService")
	s.db.Shutdown()
}

func (s *DISReaderService) AttachShutdownHook() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		s.logger.Info("Signal received, shutting down", "signal", sig)
		s.Shutdown()
		os.Exit(0)
	}()
}
