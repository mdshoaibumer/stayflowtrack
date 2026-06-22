"use client";

import React, { useState, useCallback, useRef, useEffect } from "react";
import { useApi } from "@/hooks/useApi";

// Types
interface CalendarUnit {
  id: string;
  unit_number: string;
  unit_type_name: string;
  floor: number;
  status: string;
}

interface CalendarEntry {
  reservation_id: string;
  guest_name: string;
  unit_id: string;
  check_in_date: string;
  check_out_date: string;
  status: string;
  nights: number;
}

interface CalendarViewData {
  property_id: string;
  property_name: string;
  start_date: string;
  end_date: string;
  units: CalendarUnit[];
  entries: CalendarEntry[];
}

interface OccupancyStats {
  total_units: number;
  occupied_units: number;
  available_units: number;
  occupancy_rate: number;
  revenue_potential: number;
}

// Color map for reservation statuses
const statusColors: Record<string, string> = {
  confirmed: "bg-blue-500 hover:bg-blue-600",
  checked_in: "bg-green-500 hover:bg-green-600",
  checked_out: "bg-gray-400",
  cancelled: "bg-red-300 line-through",
  pending: "bg-yellow-400 hover:bg-yellow-500",
};

const unitStatusColors: Record<string, string> = {
  available: "bg-green-50",
  occupied: "bg-green-100",
  reserved: "bg-blue-50",
  cleaning: "bg-yellow-50",
  maintenance: "bg-red-50",
};

// Helper: generate date array
function generateDates(start: Date, days: number): Date[] {
  const dates: Date[] = [];
  for (let i = 0; i < days; i++) {
    const d = new Date(start);
    d.setDate(d.getDate() + i);
    dates.push(d);
  }
  return dates;
}

function formatDate(d: Date): string {
  return d.toISOString().split("T")[0];
}

function daysBetween(start: string, end: string): number {
  const s = new Date(start);
  const e = new Date(end);
  return Math.round((e.getTime() - s.getTime()) / (86400000));
}

// Props
interface ReservationCalendarProps {
  propertyId: string;
  initialDate?: string;
  daysToShow?: number;
  onMoveBooking?: (reservationId: string, newUnitId: string, newCheckIn: string, newCheckOut: string) => Promise<void>;
  onCellClick?: (unitId: string, date: string) => void;
  onEntryClick?: (entry: CalendarEntry) => void;
}

export default function ReservationCalendar({
  propertyId,
  initialDate,
  daysToShow = 14,
  onMoveBooking,
  onCellClick,
  onEntryClick,
}: ReservationCalendarProps) {
  const api = useApi();
  const [startDate, setStartDate] = useState<Date>(
    initialDate ? new Date(initialDate) : new Date()
  );
  const [calendarData, setCalendarData] = useState<CalendarViewData | null>(null);
  const [occupancy, setOccupancy] = useState<OccupancyStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Drag state
  const [dragEntry, setDragEntry] = useState<CalendarEntry | null>(null);
  const [dragTarget, setDragTarget] = useState<{ unitId: string; date: string } | null>(null);
  const gridRef = useRef<HTMLDivElement>(null);

  const dates = generateDates(startDate, daysToShow);

  // Fetch calendar data
  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    const start = formatDate(startDate);
    const end = formatDate(dates[dates.length - 1]);

    try {
      const [calData, occData] = await Promise.all([
        api.get<CalendarViewData>(`/api/v1/calendar/${propertyId}`, { start, end }),
        api.get<OccupancyStats>(`/api/v1/calendar/${propertyId}/occupancy`, { date: start }).catch(() => null),
      ]);

      setCalendarData(calData);

      if (occData) {
        setOccupancy(occData);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- dates derived from startDate+daysToShow, api is stable
  }, [propertyId, startDate, daysToShow]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Navigation
  const navigateDays = (days: number) => {
    const newDate = new Date(startDate);
    newDate.setDate(newDate.getDate() + days);
    setStartDate(newDate);
  };

  // Find entries for a unit on a date (exclude cancelled/checked_out)
  const getEntryForCell = (unitId: string, date: string): CalendarEntry | null => {
    if (!calendarData) return null;
    return (
      calendarData.entries.find(
        (e) => e.unit_id === unitId && date >= e.check_in_date && date < e.check_out_date && e.status !== "cancelled" && e.status !== "checked_out"
      ) || null
    );
  };

  // Check if date is start of entry (exclude cancelled)
  const isEntryStart = (unitId: string, date: string): boolean => {
    if (!calendarData) return false;
    return calendarData.entries.some(
      (e) => e.unit_id === unitId && e.check_in_date === date && e.status !== "cancelled" && e.status !== "checked_out"
    );
  };

  // Drag handlers
  const handleDragStart = (e: React.DragEvent, entry: CalendarEntry) => {
    setDragEntry(entry);
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", entry.reservation_id);
  };

  const handleDragOver = (e: React.DragEvent, unitId: string, date: string) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    setDragTarget({ unitId, date });
  };

  const handleDragLeave = () => {
    setDragTarget(null);
  };

  const handleDrop = async (e: React.DragEvent, unitId: string, date: string) => {
    e.preventDefault();
    setDragTarget(null);

    if (!dragEntry || !onMoveBooking) return;

    const nights = daysBetween(dragEntry.check_in_date, dragEntry.check_out_date);
    const newCheckIn = date;
    const newCheckOutDate = new Date(date);
    newCheckOutDate.setDate(newCheckOutDate.getDate() + nights);
    const newCheckOut = formatDate(newCheckOutDate);

    try {
      await onMoveBooking(dragEntry.reservation_id, unitId, newCheckIn, newCheckOut);
      await fetchData();
    } catch {
      // Error handled by parent
    }
    setDragEntry(null);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
        <p className="text-red-700">{error}</p>
        <button onClick={fetchData} className="mt-2 text-sm text-red-600 underline">
          Retry
        </button>
      </div>
    );
  }

  if (!calendarData) return null;

  return (
    <div className="space-y-4">
      {/* Header with navigation and occupancy */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <button
            onClick={() => navigateDays(-daysToShow)}
            className="px-3 py-1 text-sm border rounded hover:bg-gray-100"
          >
            ← Prev
          </button>
          <button
            onClick={() => setStartDate(new Date())}
            className="px-3 py-1 text-sm border rounded hover:bg-gray-100"
          >
            Today
          </button>
          <button
            onClick={() => navigateDays(daysToShow)}
            className="px-3 py-1 text-sm border rounded hover:bg-gray-100"
          >
            Next →
          </button>
          <span className="ml-4 text-sm font-medium text-gray-700">
            {formatDate(startDate)} — {formatDate(dates[dates.length - 1])}
          </span>
        </div>

        {occupancy && (
          <div className="flex items-center gap-4 text-sm">
            <span className="text-green-700 font-medium">
              {occupancy.occupancy_rate.toFixed(0)}% Occupied
            </span>
            <span className="text-gray-500">
              {occupancy.occupied_units}/{occupancy.total_units} units
            </span>
          </div>
        )}
      </div>

      {/* Calendar Grid */}
      <div className="overflow-x-auto border rounded-lg" ref={gridRef}>
        <div
          className="grid min-w-max"
          style={{
            gridTemplateColumns: `180px repeat(${daysToShow}, minmax(80px, 1fr))`,
          }}
        >
          {/* Header row - dates */}
          <div className="sticky left-0 z-10 bg-gray-100 border-b border-r p-2 font-semibold text-xs text-gray-600">
            Unit
          </div>
          {dates.map((date) => {
            const isToday = formatDate(date) === formatDate(new Date());
            const isWeekend = date.getDay() === 0 || date.getDay() === 6;
            return (
              <div
                key={formatDate(date)}
                className={`border-b border-r p-1 text-center text-xs ${
                  isToday ? "bg-blue-100 font-bold" : isWeekend ? "bg-gray-50" : "bg-gray-100"
                }`}
              >
                <div>{date.toLocaleDateString("en", { weekday: "short" })}</div>
                <div>{date.getDate()}</div>
              </div>
            );
          })}

          {/* Unit rows */}
          {calendarData.units.map((unit) => (
            <React.Fragment key={unit.id}>
              {/* Unit label */}
              <div
                className={`sticky left-0 z-10 border-b border-r p-2 text-xs ${
                  unitStatusColors[unit.status] || "bg-white"
                }`}
              >
                <div className="font-medium">{unit.unit_number}</div>
                <div className="text-gray-500">{unit.unit_type_name}</div>
              </div>

              {/* Date cells */}
              {dates.map((date) => {
                const dateStr = formatDate(date);
                const entry = getEntryForCell(unit.id, dateStr);
                const isStart = entry && isEntryStart(unit.id, dateStr);
                const isDragOver =
                  dragTarget?.unitId === unit.id && dragTarget?.date === dateStr;

                return (
                  <div
                    key={`${unit.id}-${dateStr}`}
                    className={`border-b border-r min-h-[40px] relative ${
                      isDragOver ? "bg-blue-100 ring-2 ring-blue-400" : ""
                    } ${!entry ? "hover:bg-gray-50 cursor-pointer" : ""}`}
                    onDragOver={(e) => handleDragOver(e, unit.id, dateStr)}
                    onDragLeave={handleDragLeave}
                    onDrop={(e) => handleDrop(e, unit.id, dateStr)}
                    onClick={() => {
                      if (!entry && onCellClick) onCellClick(unit.id, dateStr);
                    }}
                  >
                    {entry && isStart && (
                      <div
                        className={`absolute inset-y-0 left-0 rounded-sm m-[2px] px-1 flex items-center text-white text-[10px] font-medium cursor-move overflow-hidden whitespace-nowrap ${
                          statusColors[entry.status] || "bg-gray-500"
                        }`}
                        style={{
                          width: `calc(${daysBetween(entry.check_in_date, entry.check_out_date)} * 100% - 4px)`,
                          zIndex: 5,
                        }}
                        draggable={
                          entry.status === "confirmed" || entry.status === "pending"
                        }
                        onDragStart={(e) => handleDragStart(e, entry)}
                        onClick={(e) => {
                          e.stopPropagation();
                          if (onEntryClick) onEntryClick(entry);
                        }}
                        title={`${entry.guest_name} (${entry.status})\n${entry.check_in_date} → ${entry.check_out_date}`}
                      >
                        {entry.guest_name}
                      </div>
                    )}
                  </div>
                );
              })}
            </React.Fragment>
          ))}
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4 text-xs text-gray-600">
        <span className="font-medium">Status:</span>
        {Object.entries(statusColors).map(([status, color]) => (
          <div key={status} className="flex items-center gap-1">
            <div className={`w-3 h-3 rounded-sm ${color}`}></div>
            <span className="capitalize">{status.replaceAll("_", " ")}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
