package types

import "testing"

func TestResolveThemeDefaultsToCatalogDefault(t *testing.T) {
	theme := ResolveTheme("")
	expected := ResolveTheme(DefaultThemeName)
	if theme == nil || expected == nil {
		t.Fatalf("expected non-nil themes")
	}
	if theme.Name != expected.Name {
		t.Fatalf("expected default theme name %s, got %s", expected.Name, theme.Name)
	}
}

func TestResolveThemeReturnsCloneForKnownKey(t *testing.T) {
	first := ResolveTheme("aurora-dark")
	second := ResolveTheme("Aurora-Dark")
	if first == nil || second == nil {
		t.Fatalf("expected non-nil themes")
	}
	if first == second {
		t.Fatalf("expected distinct clones, got same pointer")
	}
	if first.Name != second.Name {
		t.Fatalf("expected names to match, got %s vs %s", first.Name, second.Name)
	}
}

func TestResolveThemeFallsBackOnUnknown(t *testing.T) {
	unknown := ResolveTheme("not-a-theme")
	if unknown.Name != DefualtTerminalTheme.Name {
		t.Fatalf("expected fallback to terminal theme, got %s", unknown.Name)
	}
}
