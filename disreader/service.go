package disreader

import (
	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/internal"
	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

func NewDISReaderService(config *types.DISConfig, logger *slog.Logger) (types.DISReaderService, error) {
	return internal.NewDISReaderService(config, logger)
}

// NewDISReaderServiceWithAuth allows callers to supply a custom AgentAuthenticator for bridge mode.
// Pass nil to retain the default config-based authenticator.
func NewDISReaderServiceWithAuth(config *types.DISConfig, logger *slog.Logger, auth bridge.AgentAuthenticator) (types.DISReaderService, error) {
	return internal.NewDISReaderServiceWithAuth(config, logger, auth)
}
