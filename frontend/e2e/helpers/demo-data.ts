/**
 * Demo Data Loader
 *
 * Provides a complete set of demo data that can be loaded after registration.
 * This creates a realistic hotel scenario with properties, rooms, guests, and reservations.
 */

import { TEST_CONFIG } from "./test-config";

export interface DemoDataResult {
  property: { id: string; name: string };
  unitTypes: { id: string; name: string; base_rate: number }[];
  units: { id: string; unit_number: string; unit_type_id: string }[];
  guests: { id: string; first_name: string; last_name: string; phone: string }[];
  reservations: { id: string; guest_id: string; unit_id: string; status: string }[];
}

const DEMO_PROPERTY = {
  name: "Grand StayFlow Hotel",
  address: "123 Marine Drive",
  city: "Mumbai",
  state: "Maharashtra",
  country: "India",
  pincode: "400001",
  phone: "+912222001234",
  email: "info@grandstayflow.com",
};

const DEMO_UNIT_TYPES = [
  { name: "Standard Room", base_rate: 2500, max_occupancy: 2, amenities: ["WiFi", "AC", "TV"] },
  { name: "Deluxe Room", base_rate: 4500, max_occupancy: 2, amenities: ["WiFi", "AC", "TV", "Minibar", "Balcony"] },
  { name: "Suite", base_rate: 8000, max_occupancy: 4, amenities: ["WiFi", "AC", "TV", "Minibar", "Balcony", "Living Room", "Kitchen"] },
  { name: "Family Room", base_rate: 6000, max_occupancy: 5, amenities: ["WiFi", "AC", "TV", "Extra Beds", "Kids Area"] },
];

const DEMO_UNITS = [
  // Standard Rooms - Floor 1
  { unit_number: "101", floor: 1, type_index: 0 },
  { unit_number: "102", floor: 1, type_index: 0 },
  { unit_number: "103", floor: 1, type_index: 0 },
  { unit_number: "104", floor: 1, type_index: 0 },
  { unit_number: "105", floor: 1, type_index: 0 },
  // Deluxe Rooms - Floor 2
  { unit_number: "201", floor: 2, type_index: 1 },
  { unit_number: "202", floor: 2, type_index: 1 },
  { unit_number: "203", floor: 2, type_index: 1 },
  { unit_number: "204", floor: 2, type_index: 1 },
  // Suites - Floor 3
  { unit_number: "301", floor: 3, type_index: 2 },
  { unit_number: "302", floor: 3, type_index: 2 },
  // Family Rooms - Floor 1
  { unit_number: "106", floor: 1, type_index: 3 },
  { unit_number: "107", floor: 1, type_index: 3 },
];

const DEMO_GUESTS = [
  { first_name: "Rajesh", last_name: "Sharma", phone: "+919876543001", email: "rajesh.sharma@email.com", city: "Delhi", state: "Delhi", country: "India" },
  { first_name: "Priya", last_name: "Patel", phone: "+919876543002", email: "priya.patel@email.com", city: "Ahmedabad", state: "Gujarat", country: "India" },
  { first_name: "Amit", last_name: "Kumar", phone: "+919876543003", email: "amit.kumar@email.com", city: "Bangalore", state: "Karnataka", country: "India" },
  { first_name: "Sneha", last_name: "Reddy", phone: "+919876543004", email: "sneha.reddy@email.com", city: "Hyderabad", state: "Telangana", country: "India" },
  { first_name: "Vikram", last_name: "Singh", phone: "+919876543005", email: "vikram.singh@email.com", city: "Jaipur", state: "Rajasthan", country: "India" },
  { first_name: "Anita", last_name: "Desai", phone: "+919876543006", email: "anita.desai@email.com", city: "Pune", state: "Maharashtra", country: "India" },
  { first_name: "Suresh", last_name: "Nair", phone: "+919876543007", email: "suresh.nair@email.com", city: "Kochi", state: "Kerala", country: "India" },
  { first_name: "Kavita", last_name: "Joshi", phone: "+919876543008", email: "kavita.joshi@email.com", city: "Lucknow", state: "Uttar Pradesh", country: "India" },
];

/**
 * Load complete demo data via API calls.
 * Called after registration to populate the tenant with realistic data.
 */
export async function loadDemoData(token: string): Promise<DemoDataResult> {
  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };
  const apiUrl = TEST_CONFIG.API_URL;

  // 1. Create Property
  const propResp = await fetch(`${apiUrl}/api/v1/properties`, {
    method: "POST",
    headers,
    body: JSON.stringify(DEMO_PROPERTY),
  });
  const propData = await propResp.json();
  const property = propData.data;

  // 2. Create Unit Types
  const unitTypes: DemoDataResult["unitTypes"] = [];
  for (const ut of DEMO_UNIT_TYPES) {
    const resp = await fetch(`${apiUrl}/api/v1/properties/${property.id}/unit-types`, {
      method: "POST",
      headers,
      body: JSON.stringify(ut),
    });
    const data = await resp.json();
    unitTypes.push(data.data);
  }

  // 3. Create Units
  const units: DemoDataResult["units"] = [];
  for (const u of DEMO_UNITS) {
    const resp = await fetch(`${apiUrl}/api/v1/properties/${property.id}/units`, {
      method: "POST",
      headers,
      body: JSON.stringify({
        unit_number: u.unit_number,
        floor: u.floor,
        unit_type_id: unitTypes[u.type_index].id,
      }),
    });
    const data = await resp.json();
    units.push(data.data);
  }

  // 4. Create Guests
  const guests: DemoDataResult["guests"] = [];
  for (const g of DEMO_GUESTS) {
    const resp = await fetch(`${apiUrl}/api/v1/guests`, {
      method: "POST",
      headers,
      body: JSON.stringify(g),
    });
    const data = await resp.json();
    guests.push(data.data);
  }

  // 5. Create Reservations (upcoming dates)
  const reservations: DemoDataResult["reservations"] = [];
  const today = new Date();

  // Active reservations for first 3 guests
  for (let i = 0; i < 3; i++) {
    const checkIn = new Date(today);
    checkIn.setDate(today.getDate() + i + 1);
    const checkOut = new Date(checkIn);
    checkOut.setDate(checkIn.getDate() + 3);

    const resp = await fetch(`${apiUrl}/api/v1/reservations`, {
      method: "POST",
      headers,
      body: JSON.stringify({
        property_id: property.id,
        unit_id: units[i].id,
        guest_id: guests[i].id,
        check_in_date: checkIn.toISOString().split("T")[0],
        check_out_date: checkOut.toISOString().split("T")[0],
        rate_per_night: DEMO_UNIT_TYPES[0].base_rate,
        booking_source: "walk_in",
        adults: 2,
        children: 0,
        notes: "Demo reservation",
      }),
    });
    const data = await resp.json();
    if (data.data) {
      reservations.push(data.data);
    }
  }

  return { property, unitTypes, units, guests, reservations };
}

/**
 * Quick demo data load — minimal set for smoke testing
 */
export async function loadMinimalDemoData(token: string) {
  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };
  const apiUrl = TEST_CONFIG.API_URL;

  // Single property with 2 rooms and 1 guest
  const propResp = await fetch(`${apiUrl}/api/v1/properties`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      name: "Quick Test Hotel",
      address: "1 Test Street",
      city: "Mumbai",
      state: "Maharashtra",
      country: "India",
    }),
  });
  const property = (await propResp.json()).data;

  const utResp = await fetch(`${apiUrl}/api/v1/properties/${property.id}/unit-types`, {
    method: "POST",
    headers,
    body: JSON.stringify({ name: "Standard", base_rate: 2000, max_occupancy: 2 }),
  });
  const unitType = (await utResp.json()).data;

  const unit1Resp = await fetch(`${apiUrl}/api/v1/properties/${property.id}/units`, {
    method: "POST",
    headers,
    body: JSON.stringify({ unit_number: "101", floor: 1, unit_type_id: unitType.id }),
  });
  const unit1 = (await unit1Resp.json()).data;

  const guestResp = await fetch(`${apiUrl}/api/v1/guests`, {
    method: "POST",
    headers,
    body: JSON.stringify({ first_name: "Test", last_name: "Guest", phone: "+919999000001" }),
  });
  const guest = (await guestResp.json()).data;

  return { property, unitType, unit1, guest };
}
