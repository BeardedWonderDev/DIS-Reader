package disreader

import (
	"github.com/BeardedWonderDev/DIS-Reader/internal"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

func NewDISReaderService(config *types.Config) types.DISReaderService {
	return internal.NewDISReaderService(config)
}
