/**
 * Shared API response & domain types for StayFlow Track
 * Eliminates `any` usage across the application
 */

// Generic paginated response
export interface PaginatedResponse<T> {
  data: T[];
  meta?: {
    total: number;
    total_pages: number;
    page: number;
    per_page: number;
  };
}

// API list response (can be array or paginated)
export type ListResponse<T> = T[] | PaginatedResponse<T>;

// Reservation
export interface ReservationRecord {
  id: string;
  reservation_id?: string;
  guest_name: string;
  guest_id: string;
  unit_number: string;
  unit_id: string;
  status: string;
  check_in_date: string;
  check_out_date: string;
  total_amount?: number;
  notes?: string;
  created_at?: string;
}

// Guest
export interface GuestRecord {
  id: string;
  full_name: string;
  first_name?: string;
  last_name?: string;
  email: string;
  phone: string;
  id_type?: string;
  id_number?: string;
  nationality?: string;
  total_stays: number;
  last_stay_date: string;
  created_at: string;
}

// Unit
export interface UnitRecord {
  id: string;
  unit_number: string;
  status: string;
  unit_type_name?: string;
  unit_type_id?: string;
  floor?: number;
  guest_name?: string;
}

// Folio & Billing
export interface FolioRecord {
  id: string;
  reservation_id: string;
  guest_name: string;
  unit_number: string;
  status: string;
  total_charges: number;
  total_payments: number;
  balance: number;
  created_at: string;
}

export interface LineItem {
  id: string;
  folio_id: string;
  description: string;
  amount: number;
  category: string;
  status: string;
  created_at: string;
}

export interface PaymentRecord {
  id: string;
  folio_id: string;
  amount: number;
  method: string;
  reference?: string;
  created_at: string;
}

// Housekeeping
export interface HousekeepingTask {
  id: string;
  unit_id: string;
  unit_number: string;
  task_type: string;
  status: string;
  priority: string;
  assigned_to?: string;
  notes?: string;
  created_at: string;
}

// Laundry
export interface LaundryOrder {
  id: string;
  guest_name: string;
  unit_number: string;
  status: string;
  total_amount: number;
  items_count: number;
  created_at: string;
}

export interface RateCard {
  id: string;
  item_name: string;
  category: string;
  regular_price: number;
  express_price: number;
}

// Dashboard
export interface DashboardMetrics {
  date: string;
  occupancy: {
    total_units: number;
    occupied_units: number;
    available_units: number;
    occupancy_rate: number;
  };
  revenue: {
    today: number;
    this_week: number;
    this_month: number;
    currency: string;
  };
  operations: {
    check_ins_today: number;
    check_outs_today: number;
    expected_arrivals: number;
    expected_departures: number;
  };
  housekeeping: { counts: Record<string, number>; total: number };
  laundry: { counts: Record<string, number>; total: number };
  pending_payments: {
    pending_count: number;
    pending_amount: number;
    overdue_count: number;
  };
}

// Helper to extract array from API response
export function extractArray<T>(data: ListResponse<T>): T[] {
  return Array.isArray(data) ? data : data?.data || [];
}
