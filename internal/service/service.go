package service

import "github.com/BeardedWonderDev/DIS-Reader/types"

type Service interface {
	DIS() types.DISReaderService
}
