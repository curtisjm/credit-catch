package credits

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GeneratePeriodsForCard creates credit_periods rows for all credit definitions
// on a user's card. Called when a card is added to a user's collection.
//
// Monthly credits: anchored to statementCloseDay.
// Annual/one-time credits: anchored to openedDate (skipped if nil).
func GeneratePeriodsForCard(ctx context.Context, db *pgxpool.Pool, userCardID string, statementCloseDay int, openedDate *time.Time) error {
	// Look up which card this user_card points to, and get its credit definitions.
	rows, err := db.Query(ctx,
		`SELECT cd.id, cd.period
		 FROM credit_definitions cd
		 JOIN user_cards uc ON uc.card_catalog_id = cd.card_catalog_id
		 WHERE uc.id = $1`, userCardID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type creditDef struct {
		ID     string
		Period string
	}

	var defs []creditDef
	for rows.Next() {
		var d creditDef
		if err := rows.Scan(&d.ID, &d.Period); err != nil {
			return err
		}
		defs = append(defs, d)
	}

	now := time.Now()

	for _, d := range defs {
		var start, end time.Time

		switch d.Period {
		case "monthly":
			start, end = currentMonthlyPeriod(now, statementCloseDay)
		case "annual":
			if openedDate == nil {
				continue // no opened date → skip auto-generation
			}
			start, end = currentAnnualPeriod(now, *openedDate)
		case "one_time":
			if openedDate == nil {
				continue
			}
			// One-time credit: from opened date, no natural end.
			// Use a far-future end so it always appears as "current".
			start = *openedDate
			end = time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
		default:
			continue
		}

		_, err := db.Exec(ctx,
			`INSERT INTO credit_periods (user_card_id, credit_definition_id, period_start, period_end)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (user_card_id, credit_definition_id, period_start) DO NOTHING`,
			userCardID, d.ID, start, end,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// currentMonthlyPeriod returns the period boundaries for today given a
// statement close day. If close day is 15:
//   - Before the 15th: period is prev-month 16th → this month 15th
//   - On/after the 15th: period is this month 16th → next month 15th
func currentMonthlyPeriod(now time.Time, closeDay int) (time.Time, time.Time) {
	y, m, d := now.Date()

	// Clamp close day to actual days in month.
	closeDay = clampDay(y, m, closeDay)

	if d <= closeDay {
		// We're in the period that ends this month on closeDay.
		prevMonth := time.Date(y, m-1, 1, 0, 0, 0, 0, time.UTC)
		startDay := clampDay(prevMonth.Year(), prevMonth.Month(), closeDay) + 1
		// Handle case where startDay exceeds days in prev month (close day is last day).
		if startDay > daysInMonth(prevMonth.Year(), prevMonth.Month()) {
			// Start is the 1st of current month.
			return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC),
				time.Date(y, m, closeDay, 0, 0, 0, 0, time.UTC)
		}
		return time.Date(prevMonth.Year(), prevMonth.Month(), startDay, 0, 0, 0, 0, time.UTC),
			time.Date(y, m, closeDay, 0, 0, 0, 0, time.UTC)
	}

	// We're past closeDay, so we're in the period that starts this month.
	startDay := closeDay + 1
	nextMonth := time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC)
	endDay := clampDay(nextMonth.Year(), nextMonth.Month(), closeDay)
	return time.Date(y, m, startDay, 0, 0, 0, 0, time.UTC),
		time.Date(nextMonth.Year(), nextMonth.Month(), endDay, 0, 0, 0, 0, time.UTC)
}

// currentAnnualPeriod returns the annual period boundaries for today given
// the card's opened date. The annual cycle resets on the anniversary.
func currentAnnualPeriod(now time.Time, opened time.Time) (time.Time, time.Time) {
	annivMonth := opened.Month()
	annivDay := opened.Day()

	// Find the most recent anniversary on or before today.
	thisYearAnniv := time.Date(now.Year(), annivMonth, clampDay(now.Year(), annivMonth, annivDay), 0, 0, 0, 0, time.UTC)

	var start time.Time
	if !thisYearAnniv.After(now) {
		start = thisYearAnniv
	} else {
		start = time.Date(now.Year()-1, annivMonth, clampDay(now.Year()-1, annivMonth, annivDay), 0, 0, 0, 0, time.UTC)
	}

	// End is the day before the next anniversary.
	nextAnniv := time.Date(start.Year()+1, annivMonth, clampDay(start.Year()+1, annivMonth, annivDay), 0, 0, 0, 0, time.UTC)
	end := nextAnniv.AddDate(0, 0, -1)

	return start, end
}

func clampDay(year int, month time.Month, day int) int {
	max := daysInMonth(year, month)
	if day > max {
		return max
	}
	return day
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
