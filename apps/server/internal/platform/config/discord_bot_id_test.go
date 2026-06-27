package config

import "testing"

func TestIsMisplacedBotProfileID(t *testing.T) {
	if !IsMisplacedBotProfileID("1519615970128171068") {
		t.Fatal("snowflake should be misplaced")
	}
	if !IsMisplacedBotProfileID("app-1520303229869756487") {
		t.Fatal("app- prefix should be misplaced")
	}
	if IsMisplacedBotProfileID("designer") {
		t.Fatal("expert codename should not be misplaced")
	}
}
