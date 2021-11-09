package wol

import "testing"

func TestWol(t *testing.T) {
	err := SendWakeOnLanPacket("78:84:3c:2a:ac:08")
	if err != nil {
		t.Fatal(err)
	}
}
