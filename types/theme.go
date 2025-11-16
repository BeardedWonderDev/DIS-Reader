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

const DefaultThemeName = "aurora-dark"

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

var AuroraDarkTheme = Theme{
	Name: "Aurora Dark",
	Colors: ColorValue{
		Error:           tcell.GetColor("#ff7f87"),
		Warning:         tcell.GetColor("#ffd479"),
		Notice:          tcell.GetColor("#8fd5ff"),
		WindowColor:     tcell.GetColor("#23283b"),
		ModalColor:      tcell.GetColor("#2c3147"),
		CommandBarColor: tcell.GetColor("#1b1f2d"),
		PrimaryText:     tcell.GetColor("#f8fbff"),
		SecondaryText:   tcell.GetColor("#bac4dd"),
		AccentColor:     tcell.GetColor("#5be7c4"),
		AccentTextColor: tcell.GetColor("#071510"),
		BorderColor:     tcell.GetColor("#6f7ab8"),
		TableHeader:     tcell.GetColor("#2f3650"),
		TableRow:        tcell.GetColor("#262c3f"),
		StatusBarBg:     tcell.GetColor("#1f2435"),
		StatusBarText:   tcell.GetColor("#fefefe"),
	},
	Style: ComponentStyle{
		ButtonStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#ff956b")).
			Foreground(tcell.GetColor("#1b0c05")).Bold(true),
		PlaceholderStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#343b52")).
			Foreground(tcell.GetColor("#9aa7c8")).Italic(true),
		FieldStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#2e3449")).
			Foreground(tcell.GetColor("#f8fbff")),
		ListMainTextStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#23283b")).
			Foreground(tcell.GetColor("#f8fbff")),
		ListSelectedStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#ffd479")).
			Foreground(tcell.GetColor("#3a2a08")).Bold(true),
		ListBorderStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#23283b")).
			Foreground(tcell.GetColor("#6f7ab8")),
		TextAreaStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#23283b")).
			Foreground(tcell.GetColor("#f8fbff")),
		TableHeaderStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#2f3650")).
			Foreground(tcell.GetColor("#fefefe")).Bold(true),
		TableCellStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#262c3f")).
			Foreground(tcell.GetColor("#e4e9fb")),
		TableSelectedStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#5be7c4")).
			Foreground(tcell.GetColor("#071510")).Bold(true),
		StatusBarStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#1f2435")).
			Foreground(tcell.GetColor("#fefefe")).Bold(true),
		PickerCellStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#23283b")).
			Foreground(tcell.GetColor("#f8fbff")),
		PickerSelectedStyle: tcell.StyleDefault.
			Background(tcell.GetColor("#5be7c4")).
			Foreground(tcell.GetColor("#071510")).Bold(true),
	},
}

var themeCatalog = map[string]*Theme{
	"terminal":           &DefualtTerminalTheme,
	"terminal-dark":      &DefualtTerminalTheme,
	"high-contrast-dark": &AuroraDarkTheme,
	"aurora-dark":        &AuroraDarkTheme,
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
