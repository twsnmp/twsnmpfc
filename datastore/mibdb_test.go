package datastore

import (
	"testing"

	"github.com/gosnmp/gosnmp"
)

func TestPrintHintedMIBIntVal(t *testing.T) {
	tests := []struct {
		val  int32
		hint string
		us   bool
		want string
	}{
		{1234, "d-3", false, "1.234"},
		{1234, "d-2", false, "12.34"},
		{1234, "d-1", false, "123.4"},
		{1234, "d-0", false, "1234"},
		{1, "d-3", false, "0.001"},
		{12, "d-3", false, "0.012"},
		{123, "d-3", false, "0.123"},
		{-1234, "d-3", false, "-1.234"},
	}
	for _, tt := range tests {
		got := PrintHintedMIBIntVal(tt.val, tt.hint, tt.us)
		if got != tt.want {
			t.Errorf("PrintHintedMIBIntVal(%d, %q, %v) = %q, want %q", tt.val, tt.hint, tt.us, got, tt.want)
		}
	}
}

func TestGetMIBValueString(t *testing.T) {
	// 1 day + 1 second = 86400*100 + 1*100 = 8640000 + 100 = 8640100 ticks
	ticks := uint32(8640100)
	pdu := gosnmp.SnmpPDU{
		Type:  gosnmp.TimeTicks,
		Value: ticks,
	}
	got := GetMIBValueString("test", &pdu, false)
	want := "8640100(1 days, 1s)"
	if got != want {
		t.Errorf("GetMIBValueString(TimeTicks) = %q, want %q", got, want)
	}
}
