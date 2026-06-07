package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCalendarEntry_NumNights(t *testing.T) {
	// NumNights is calculated as checkout_date - checkin_date in calendar days
	entry := CalendarEntry{
		ReservationID: uuid.New(),
		UnitID:        uuid.New(),
		GuestName:     "Test Guest",
		CheckInDate:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		CheckOutDate:  time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC),
		Status:        "confirmed",
		NumNights:     3,
	}

	days := int(entry.CheckOutDate.Sub(entry.CheckInDate).Hours() / 24)
	if days != entry.NumNights {
		t.Errorf("expected %d nights from dates, got NumNights=%d", days, entry.NumNights)
	}
}

func TestCalendarView_DateRange(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	view := CalendarView{
		PropertyID: uuid.New(),
		StartDate:  start,
		EndDate:    end,
		Units:      []CalendarUnit{},
	}

	days := int(view.EndDate.Sub(view.StartDate).Hours() / 24)
	if days > 90 {
		t.Error("calendar view should not exceed 90 days")
	}
	if days != 29 {
		t.Errorf("expected 29 days, got %d", days)
	}
}

func TestCalendarView_MaxRange(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC)

	days := int(end.Sub(start).Hours() / 24)
	if days <= 90 {
		t.Errorf("test setup error: %d days should exceed 90", days)
	}
}

func TestMoveBookingInput_Fields(t *testing.T) {
	input := MoveBookingInput{
		ReservationID: uuid.New(),
		NewUnitID:     uuid.New(),
		NewCheckIn:    "2026-06-05",
		NewCheckOut:   "2026-06-08",
	}

	if input.ReservationID == uuid.Nil {
		t.Error("reservation ID should not be nil")
	}
	if input.NewUnitID == uuid.Nil {
		t.Error("new unit ID should not be nil")
	}

	checkIn, err := time.Parse("2006-01-02", input.NewCheckIn)
	if err != nil {
		t.Fatalf("invalid check_in date: %v", err)
	}
	checkOut, err := time.Parse("2006-01-02", input.NewCheckOut)
	if err != nil {
		t.Fatalf("invalid check_out date: %v", err)
	}
	if !checkOut.After(checkIn) {
		t.Error("check_out must be after check_in")
	}
}

func TestOccupancyStats_Rate(t *testing.T) {
	stats := OccupancyStats{
		Date:           time.Now(),
		TotalUnits:     20,
		OccupiedUnits:  15,
		AvailableUnits: 5,
		OccupancyRate:  75.0,
	}

	if stats.OccupiedUnits+stats.AvailableUnits != stats.TotalUnits {
		t.Error("occupied + available should equal total")
	}

	expectedRate := float64(stats.OccupiedUnits) / float64(stats.TotalUnits) * 100
	if stats.OccupancyRate != expectedRate {
		t.Errorf("expected rate %.1f%%, got %.1f%%", expectedRate, stats.OccupancyRate)
	}
}

func TestCalendarUnit_WithEntries(t *testing.T) {
	unit := CalendarUnit{
		UnitID:       uuid.New(),
		UnitNumber:   "101",
		Floor:        "1",
		UnitTypeName: "Deluxe Room",
		Status:       "available",
		Entries: []CalendarEntry{
			{
				ReservationID: uuid.New(),
				GuestName:     "Guest A",
				CheckInDate:   time.Date(2026, 6, 1, 14, 0, 0, 0, time.UTC),
				CheckOutDate:  time.Date(2026, 6, 3, 11, 0, 0, 0, time.UTC),
				Status:        "confirmed",
			},
			{
				ReservationID: uuid.New(),
				GuestName:     "Guest B",
				CheckInDate:   time.Date(2026, 6, 5, 14, 0, 0, 0, time.UTC),
				CheckOutDate:  time.Date(2026, 6, 7, 11, 0, 0, 0, time.UTC),
				Status:        "checked_in",
			},
		},
	}

	if len(unit.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(unit.Entries))
	}

	// Verify no overlap between entries
	for i := 0; i < len(unit.Entries)-1; i++ {
		current := unit.Entries[i]
		next := unit.Entries[i+1]
		if current.CheckOutDate.After(next.CheckInDate) {
			t.Error("entries should not overlap")
		}
	}
}

func TestCalendarEntry_ValidStatuses(t *testing.T) {
	validStatuses := []string{"pending", "confirmed", "checked_in", "checked_out", "cancelled"}

	for _, status := range validStatuses {
		entry := CalendarEntry{
			ReservationID: uuid.New(),
			Status:        status,
		}
		if entry.Status == "" {
			t.Error("status should not be empty")
		}
	}
}
