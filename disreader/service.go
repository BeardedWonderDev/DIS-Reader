package disreader

import (
	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/internal"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

func NewDISReaderService(config *types.Config, logger *slog.Logger) types.DISReaderService {
	return internal.NewDISReaderService(config, logger)
}
