package credits

import (
	"testing"
	"time"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestCurrentMonthlyPeriod_BeforeCloseDay(t *testing.T) {
	// Close day 15, today is March 10 → period is Feb 16 to Mar 15.
	now := date(2026, time.March, 10)
	start, end := currentMonthlyPeriod(now, 15)

	wantStart := date(2026, time.February, 16)
	wantEnd := date(2026, time.March, 15)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentMonthlyPeriod_OnCloseDay(t *testing.T) {
	// Close day 15, today is March 15 → still in Feb 16 to Mar 15 period.
	now := date(2026, time.March, 15)
	start, end := currentMonthlyPeriod(now, 15)

	wantStart := date(2026, time.February, 16)
	wantEnd := date(2026, time.March, 15)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentMonthlyPeriod_AfterCloseDay(t *testing.T) {
	// Close day 15, today is March 20 → period is Mar 16 to Apr 15.
	now := date(2026, time.March, 20)
	start, end := currentMonthlyPeriod(now, 15)

	wantStart := date(2026, time.March, 16)
	wantEnd := date(2026, time.April, 15)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentMonthlyPeriod_CloseDay31_InFeb(t *testing.T) {
	// Close day 31, today is Feb 15 → close day clamps to 28.
	// Period: Jan 29 (31+1=32→clamped: prev month close=31, +1=Jan 32→Feb 1? No...)
	// Actually: close day 31 in Feb clamps to 28. Today (15) <= 28, so:
	//   prev month = Jan, close day in Jan = 31. Start = Jan 32 → but that overflows.
	//   Start should be Feb 1 (day after Jan 31). End = Feb 28.
	now := date(2026, time.February, 15)
	start, end := currentMonthlyPeriod(now, 31)

	wantStart := date(2026, time.February, 1)
	wantEnd := date(2026, time.February, 28)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentMonthlyPeriod_CloseDay1(t *testing.T) {
	// Close day 1, today is March 1 → period is Feb 2 to Mar 1.
	now := date(2026, time.March, 1)
	start, end := currentMonthlyPeriod(now, 1)

	wantStart := date(2026, time.February, 2)
	wantEnd := date(2026, time.March, 1)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentMonthlyPeriod_CloseDay1_AfterCloseDay(t *testing.T) {
	// Close day 1, today is March 15 → period is Mar 2 to Apr 1.
	now := date(2026, time.March, 15)
	start, end := currentMonthlyPeriod(now, 1)

	wantStart := date(2026, time.March, 2)
	wantEnd := date(2026, time.April, 1)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentMonthlyPeriod_YearBoundary(t *testing.T) {
	// Close day 10, today is Jan 5 → period is Dec 11 to Jan 10.
	now := date(2026, time.January, 5)
	start, end := currentMonthlyPeriod(now, 10)

	wantStart := date(2025, time.December, 11)
	wantEnd := date(2026, time.January, 10)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentAnnualPeriod_BeforeAnniversary(t *testing.T) {
	// Opened Sep 3, today is Apr 8 → period is Sep 3, 2025 to Sep 2, 2026.
	now := date(2026, time.April, 8)
	opened := date(2023, time.September, 3)
	start, end := currentAnnualPeriod(now, opened)

	wantStart := date(2025, time.September, 3)
	wantEnd := date(2026, time.September, 2)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentAnnualPeriod_OnAnniversary(t *testing.T) {
	// Opened Sep 3, today IS Sep 3 → new period starts: Sep 3, 2026 to Sep 2, 2027.
	now := date(2026, time.September, 3)
	opened := date(2023, time.September, 3)
	start, end := currentAnnualPeriod(now, opened)

	wantStart := date(2026, time.September, 3)
	wantEnd := date(2027, time.September, 2)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentAnnualPeriod_AfterAnniversary(t *testing.T) {
	// Opened Sep 3, today is Oct 15 → period is Sep 3, 2026 to Sep 2, 2027.
	now := date(2026, time.October, 15)
	opened := date(2023, time.September, 3)
	start, end := currentAnnualPeriod(now, opened)

	wantStart := date(2026, time.September, 3)
	wantEnd := date(2027, time.September, 2)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestCurrentAnnualPeriod_LeapDayOpened(t *testing.T) {
	// Opened Feb 29 (leap year), current year is non-leap → anniversary clamps to Feb 28.
	now := date(2027, time.March, 1)
	opened := date(2024, time.February, 29)
	start, end := currentAnnualPeriod(now, opened)

	// Anniversary clamps to Feb 28 in 2027 (non-leap).
	wantStart := date(2027, time.February, 28)
	wantEnd := date(2028, time.February, 28) // 2028 is leap, so Feb 29 - 1 = Feb 28

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestDaysInMonth(t *testing.T) {
	tests := []struct {
		year  int
		month time.Month
		want  int
	}{
		{2026, time.January, 31},
		{2026, time.February, 28},
		{2024, time.February, 29}, // leap year
		{2026, time.April, 30},
		{2026, time.December, 31},
	}

	for _, tt := range tests {
		got := daysInMonth(tt.year, tt.month)
		if got != tt.want {
			t.Errorf("daysInMonth(%d, %v) = %d, want %d", tt.year, tt.month, got, tt.want)
		}
	}
}

func TestClampDay(t *testing.T) {
	tests := []struct {
		year  int
		month time.Month
		day   int
		want  int
	}{
		{2026, time.February, 31, 28},
		{2024, time.February, 31, 29},
		{2026, time.January, 31, 31},
		{2026, time.April, 30, 30},
		{2026, time.April, 31, 30},
		{2026, time.March, 15, 15},
	}

	for _, tt := range tests {
		got := clampDay(tt.year, tt.month, tt.day)
		if got != tt.want {
			t.Errorf("clampDay(%d, %v, %d) = %d, want %d", tt.year, tt.month, tt.day, got, tt.want)
		}
	}
}
