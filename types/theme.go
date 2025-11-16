package types

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ColorValue struct {
	Error           tcell.Color
	Warning         tcell.Color
	Notice          tcell.Color
	WindowColor     tcell.Color
	ModalColor      tcell.Color
	CommandBarColor tcell.Color
	PrimaryText     tcell.Color
	SecondaryText   tcell.Color
	AccentColor     tcell.Color
	AccentTextColor tcell.Color
	BorderColor     tcell.Color
	TableHeader     tcell.Color
	TableRow        tcell.Color
	StatusBarBg     tcell.Color
	StatusBarText   tcell.Color
}

type ComponentStyle struct {
	ButtonStyle         tcell.Style
	PlaceholderStyle    tcell.Style
	FieldStyle          tcell.Style
	ListMainTextStyle   tcell.Style
	ListSelectedStyle   tcell.Style
	ListBorderStyle     tcell.Style
	TextAreaStyle       tcell.Style
	TableHeaderStyle    tcell.Style
	TableCellStyle      tcell.Style
	TableSelectedStyle  tcell.Style
	StatusBarStyle      tcell.Style
	PickerCellStyle     tcell.Style
	PickerSelectedStyle tcell.Style
}

type Theme struct {
	Name   string
	Colors ColorValue
	Style  ComponentStyle
}

const DefaultThemeName = "high-contrast-dark"

var (
	DefualtTerminalTheme = Theme{
		Name: "Terminal",
		Colors: ColorValue{
			Error:           tcell.GetColor("red"),
			Warning:         tcell.GetColor("darkred"),
			Notice:          tcell.GetColor("silver"),
			WindowColor:     tcell.GetColor("#444444"),
			ModalColor:      tcell.GetColor("#111111"),
			CommandBarColor: tcell.GetColor("#333333"),
			PrimaryText:     tcell.ColorWhite,
			SecondaryText:   tcell.GetColor("lightgray"),
			AccentColor:     tcell.GetColor("#00A5FF"),
			AccentTextColor: tcell.ColorBlack,
			BorderColor:     tcell.GetColor("#666666"),
			TableHeader:     tcell.GetColor("#555555"),
			TableRow:        tcell.GetColor("#333333"),
			StatusBarBg:     tcell.GetColor("#222222"),
			StatusBarText:   tcell.ColorWhite,
		},
		Style: ComponentStyle{
			ButtonStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#5500FF")).
				Foreground(tcell.ColorWhite).Bold(true),
			PlaceholderStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#666666")).
				Foreground(tcell.ColorBlack).Italic(true),
			FieldStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#666666")).
				Foreground(tcell.ColorWhite),
			ListMainTextStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#444444")).
				Foreground(tcell.ColorWhite),
			ListSelectedStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#777777")).
				Foreground(tcell.ColorBlack).Bold(true),
			ListBorderStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#444444")).
				Foreground(tcell.GetColor("#AAAAAA")),
			TextAreaStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#444444")).
				Foreground(tcell.ColorWhite),
			TableHeaderStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#555555")).
				Foreground(tcell.ColorWhite).Bold(true),
			TableCellStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#333333")).
				Foreground(tcell.ColorWhite),
			TableSelectedStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#00A5FF")).
				Foreground(tcell.ColorBlack),
			StatusBarStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#222222")).
				Foreground(tcell.ColorWhite).Bold(true),
			PickerCellStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#444444")).
				Foreground(tcell.ColorWhite),
			PickerSelectedStyle: tcell.StyleDefault.
				Background(tcell.GetColor("#00A5FF")).
				Foreground(tcell.ColorBlack).Bold(true),
		},
	}
)

var HighContrastDarkTheme = Theme{
	Name: "High Contrast Dark",
	Colors: ColorValue{
		Error:           tcell.GetColor("#ff6b6b"),
		Warning:         tcell.GetColor("#ffa94d"),
		Notice:          tcell.GetColor("#9be7ff"),
		WindowColor:     tcell.GetColor("#10131a"),
		ModalColor:      tcell.GetColor("#161b26"),
		CommandBarColor: tcell.GetColor("#0c111a"),
		PrimaryText:     tcell.GetColor("#f4f7ff"),
		SecondaryText:   tcell.GetColor("#9aa8c3"),
		AccentColor:     tcell.GetColor("#3d9df6"),
		AccentTextColor: tcell.GetColor("#050608"),
		BorderColor:     tcell.GetColor("#2d3445"),
		TableHeader:     tcell.GetColor("#1e2433"),
		TableRow:        tcell.GetColor("#0f131c"),
		StatusBarBg:     tcell.GetColor("#0b1220"),
		StatusBarText:   tcell.GetColor("#e7f1ff"),
	},
	Style: ComponentStyle{
		ButtonStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#3d9df6")).
			Foreground(tcell.GetColor("#050608")).Bold(true),
		PlaceholderStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#1b2332")).
			Foreground(tcell.GetColor("#9aa8c3")).Italic(true),
		FieldStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#1b2332")).
			Foreground(tcell.GetColor("#f4f7ff")),
		ListMainTextStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#10131a")).
			Foreground(tcell.GetColor("#f4f7ff")),
		ListSelectedStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#3d9df6")).
			Foreground(tcell.GetColor("#050608")).Bold(true),
		ListBorderStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#10131a")).
			Foreground(tcell.GetColor("#2d3445")),
		TextAreaStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#10131a")).
			Foreground(tcell.GetColor("#f4f7ff")),
		TableHeaderStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#1e2433")).
			Foreground(tcell.GetColor("#f4f7ff")).Bold(true),
		TableCellStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#0f131c")).
			Foreground(tcell.GetColor("#e1e6f6")),
		TableSelectedStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#3d9df6")).
			Foreground(tcell.GetColor("#050608")).Bold(true),
		StatusBarStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#0b1220")).
			Foreground(tcell.GetColor("#e7f1ff")).Bold(true),
		PickerCellStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#10131a")).
			Foreground(tcell.GetColor("#f4f7ff")),
		PickerSelectedStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#3d9df6")).
			Foreground(tcell.GetColor("#050608")).Bold(true),
	},
}

var themeCatalog = map[string]*Theme{
	"terminal":           &DefualtTerminalTheme,
	"terminal-dark":      &DefualtTerminalTheme,
	"high-contrast-dark": &HighContrastDarkTheme,
}

// ResolveTheme returns a clone of a known theme by name, falling back to the default terminal theme.
func ResolveTheme(name string) *Theme {
	key := strings.TrimSpace(strings.ToLower(name))
	if key == "" {
		key = DefaultThemeName
	}
	if base, ok := themeCatalog[key]; ok {
		clone := *base
		clone.Colors = base.Colors
		clone.Style = base.Style
		return &clone
	}
	clone := DefualtTerminalTheme
	return &clone
}
