package internal

import (
	_ "embed"
	"os"
	"os/signal"
	"syscall"

	"fmt"

	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type DISReaderService struct {
	config *types.Config
	db     *database.JDBCRunnerService
	logger *slog.Logger
}

func NewDISReaderService(config *types.Config, logger *slog.Logger) *DISReaderService {
	if logger == nil {
		logger = slog.Default()
	}
	db := database.NewJDBCRunnerService(config, logger)

	return &DISReaderService{config: config, db: db, logger: logger}
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
