package authUI

import (
	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
)

type Auth struct {
	UI            typesUI.UI
	Authenticated bool
}

func NewAuthService() Auth {
	return Auth{}
}

func (a Auth) SetUI(ui typesUI.UI) {
	a.UI = ui
}

func (a Auth) GetUI() typesUI.UI {
	return a.UI
}

func (a Auth) IsAuthenticated() bool {
	return a.Authenticated
}

func (a Auth) Authenticate() error {
	a.UI.PrintLog(entity.Log{
		Content: "🌏 Verifying DIS Connection to [blue]" + txtServerURL.GetText() + ", connecting...",
		Type:    entity.LOG_INFO,
	})

	if conErr := a.UI.GetDIS().TestDISConnection(); conErr != nil {
		return conErr
	}

	a.Authenticated = true
	a.UI.PrintLog(entity.Log{
		Content: "DIS Connection [green] Succesful",
		Type:    entity.LOG_INFO,
	})

	return nil
}
