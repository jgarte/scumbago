package scumbag

import (
	"testing"

	"fmt"
)

func newTestBot() (*Scumbag, error) {
	configFile := "../config/bot.json.test"
	logFilename := "../log/test.log"
	environment := "test"
	return NewBot(&configFile, &logFilename, &environment)
}

func TestVersionString(t *testing.T) {
	expected := fmt.Sprintf("scumbag v%s-%s", Version, BuildTag)
	if expected != VersionString() {
		t.Error("Version string is incorrect")
	}
}

func TestNewBot(t *testing.T) {
	_, err := newTestBot()
	if err != nil {
		t.Errorf("Error creating bot: %s", err)
	}
}

func TestAdmin(t *testing.T) {
	bot, _ := newTestBot()

	if bot.Admin("admin_nick") == false {
		t.Error("admin_nick should be an admin")
	}

	if bot.Admin("not_an_admin") == true {
		t.Error("not an admin")
	}
}
