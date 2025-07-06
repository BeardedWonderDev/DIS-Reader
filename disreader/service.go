package disreader

import (
	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/internal"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

func NewDISReaderService(config *types.DISConfig, logger *slog.Logger) (types.DISReaderService, error) {
	return internal.NewDISReaderService(config, logger)
}
