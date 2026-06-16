package notify

import (
	"testing"
	"time"
)

type scheduleTestEnt struct {
	Time     string
	Schedule string
	Result   bool
}

func TestNotifySchedule(t *testing.T) {
	for _, e := range []scheduleTestEnt{
		{
			Time:     "2023-01-02T15:04:05+09:00",
			Schedule: "2 15:00-16:00",
			Result:   true,
		},
		{
			Time:     "2023-01-02T15:04:05+09:00",
			Schedule: "31 15:00-16:00",
			Result:   false,
		},
		{
			Time:     "2023-01-02T15:04:05+09:00",
			Schedule: "* 15:00-16:00",
			Result:   true,
		},
		{
			Time:     "2023-01-02T15:04:05+09:00",
			Schedule: "Fri 15:00-16:00",
			Result:   false,
		},
		{
			Time:     "2023-07-12T19:04:05+09:00",
			Schedule: "Wed 18:00-20:00",
			Result:   true,
		},
		{
			Time:     "2023-07-31T19:04:05+09:00",
			Schedule: "Last 18:00-20:00",
			Result:   true,
		},
		{
			Time:     "2023-02-28T19:04:05+09:00",
			Schedule: "Last 18:00-20:00",
			Result:   true,
		},
		{
			Time:     "2023-12-28T19:04:05+09:00",
			Schedule: "Last 18:00-20:00",
			Result:   false,
		},
		{
			Time:     "2023-12-28T19:04:05+09:00",
			Schedule: "Last 18:00-20:00",
			Result:   false,
		},
		{
			Time:     "2024-02-29T19:04:05+09:00",
			Schedule: "Last 18:00-20:00",
			Result:   true,
		},
		{
			Time:     "2024-02-28T19:04:05+09:00",
			Schedule: "Last 18:00-20:00",
			Result:   false,
		},
		{
			Time:     "2023-07-14T06:58:05+09:00",
			Schedule: "Fri 05:00-06:30",
			Result:   false,
		},
	} {
		tm, err := time.Parse(time.RFC3339, e.Time)
		if err != nil {
			t.Fatal(err)
		}
		if isExcludeTime(e.Schedule, tm.UnixNano()) != e.Result {
			t.Fatalf("ent=%+v", e)
		}
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"https://discord.com/api/webhooks/123", false},
		{"http://localhost:8080/webhook", false},
		{"http://192.168.1.100/webhook", false},
		{"https://example.com", false},
		{"ftp://example.com", true},
		{"javascript:alert(1)", true},
		{"/relative/path", true},
		{"", true},
	}
	for _, tc := range tests {
		got, err := ValidateURL(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateURL(%q) err = %v, wantErr = %v", tc.input, err, tc.wantErr)
		}
		if err == nil && got != tc.input {
			t.Errorf("ValidateURL(%q) got = %q, want = %q", tc.input, got, tc.input)
		}
	}
}
