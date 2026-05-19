package ratio_setting

import "testing"

func TestGetDisplayGroupRatioFallsBackToRealRatio(t *testing.T) {
	originalGroupRatio := GroupRatio2JSONString()
	originalDisplayGroupRatio := DisplayGroupRatio2JSONString()
	t.Cleanup(func() {
		_ = UpdateGroupRatioByJSONString(originalGroupRatio)
		_ = UpdateDisplayGroupRatioByJSONString(originalDisplayGroupRatio)
	})

	if err := UpdateGroupRatioByJSONString(`{"default":1.5}`); err != nil {
		t.Fatalf("failed to set group ratio: %v", err)
	}
	if err := UpdateDisplayGroupRatioByJSONString(`{}`); err != nil {
		t.Fatalf("failed to clear display group ratio: %v", err)
	}

	if got := GetDisplayGroupRatio("default"); got != 1.5 {
		t.Fatalf("expected fallback display ratio 1.5, got %v", got)
	}
}

func TestGetDisplayGroupRatioUsesExplicitValue(t *testing.T) {
	originalGroupRatio := GroupRatio2JSONString()
	originalDisplayGroupRatio := DisplayGroupRatio2JSONString()
	t.Cleanup(func() {
		_ = UpdateGroupRatioByJSONString(originalGroupRatio)
		_ = UpdateDisplayGroupRatioByJSONString(originalDisplayGroupRatio)
	})

	if err := UpdateGroupRatioByJSONString(`{"default":1.5}`); err != nil {
		t.Fatalf("failed to set group ratio: %v", err)
	}
	if err := UpdateDisplayGroupRatioByJSONString(`{"default":1.0}`); err != nil {
		t.Fatalf("failed to set display group ratio: %v", err)
	}

	if got := GetDisplayGroupRatio("default"); got != 1.0 {
		t.Fatalf("expected explicit display ratio 1.0, got %v", got)
	}
}

func TestCheckDisplayGroupRatioRejectsNegativeValue(t *testing.T) {
	if err := CheckDisplayGroupRatio(`{"default":-1}`); err == nil {
		t.Fatal("expected negative display ratio validation error")
	}
}
