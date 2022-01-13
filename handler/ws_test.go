package handler

import (
	"os"
	"testing"
)

func init() {
	os.Setenv("GO_ENV", "testing")
}

func TestEjectConnection(t *testing.T) {
	conn := WebSocketConnection{}
	err := ejectConnection(&conn)
	if err != nil {
		t.Error("Error when ejectConnection", err)
	}
}
