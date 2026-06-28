"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/contexts/AuthContext";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

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
  { name: "Suite", base_rate: 8000, max_occupancy: 4, amenities: ["WiFi", "AC", "TV", "Minibar", "Balcony", "Living Room"] },
  { name: "Family Room", base_rate: 6000, max_occupancy: 5, amenities: ["WiFi", "AC", "TV", "Extra Beds", "Kids Area"] },
];

const DEMO_UNITS = [
  { unit_number: "101", floor: 1, type_index: 0 },
  { unit_number: "102", floor: 1, type_index: 0 },
  { unit_number: "103", floor: 1, type_index: 0 },
  { unit_number: "104", floor: 1, type_index: 0 },
  { unit_number: "105", floor: 1, type_index: 3 },
  { unit_number: "106", floor: 1, type_index: 3 },
  { unit_number: "201", floor: 2, type_index: 1 },
  { unit_number: "202", floor: 2, type_index: 1 },
  { unit_number: "203", floor: 2, type_index: 1 },
  { unit_number: "204", floor: 2, type_index: 1 },
  { unit_number: "301", floor: 3, type_index: 2 },
  { unit_number: "302", floor: 3, type_index: 2 },
];

const DEMO_GUESTS = [
  { first_name: "Rajesh", last_name: "Sharma", phone: "+919876543001", email: "rajesh.sharma@email.com", city: "Delhi", state: "Delhi", country: "India", nationality: "Indian" },
  { first_name: "Priya", last_name: "Patel", phone: "+919876543002", email: "priya.patel@email.com", city: "Ahmedabad", state: "Gujarat", country: "India", nationality: "Indian" },
  { first_name: "Amit", last_name: "Kumar", phone: "+919876543003", email: "amit.kumar@email.com", city: "Bangalore", state: "Karnataka", country: "India", nationality: "Indian" },
  { first_name: "Sneha", last_name: "Reddy", phone: "+919876543004", email: "sneha.reddy@email.com", city: "Hyderabad", state: "Telangana", country: "India", nationality: "Indian" },
  { first_name: "Vikram", last_name: "Singh", phone: "+919876543005", email: "vikram.singh@email.com", city: "Jaipur", state: "Rajasthan", country: "India", nationality: "Indian" },
  { first_name: "Anita", last_name: "Desai", phone: "+919876543006", email: "anita.desai@email.com", city: "Pune", state: "Maharashtra", country: "India", nationality: "Indian" },
  { first_name: "Suresh", last_name: "Nair", phone: "+919876543007", email: "suresh.nair@email.com", city: "Kochi", state: "Kerala", country: "India", nationality: "Indian" },
  { first_name: "Kavita", last_name: "Joshi", phone: "+919876543008", email: "kavita.joshi@email.com", city: "Lucknow", state: "Uttar Pradesh", country: "India", nationality: "Indian" },
];

interface DemoDataDialogProps {
  onComplete?: () => void;
}

export default function DemoDataDialog({ onComplete }: DemoDataDialogProps) {
  const { accessToken } = useAuth();
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [progress, setProgress] = useState("");
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Show dialog only for first-time users
    const demoShown = localStorage.getItem("demo_data_shown");
    if (demoShown || !accessToken) return;

    // Also check server-side: if user already has properties, skip the dialog
    const checkProperties = async () => {
      try {
        const resp = await fetch(`${API_BASE}/api/v1/properties`, {
          headers: { Authorization: `Bearer ${accessToken}` },
        });
        if (resp.ok) {
          const data = await resp.json();
          const properties = data.data || [];
          if (properties.length > 0) {
            // User already has data, mark as shown and don't display
            localStorage.setItem("demo_data_shown", "true");
            return;
          }
        }
      } catch {
        // On error, fall through and show the dialog
      }
      setVisible(true);
    };
    checkProperties();
  }, [accessToken]);

  const handleSkip = () => {
    localStorage.setItem("demo_data_shown", "true");
    localStorage.setItem("demo_data_loaded", "false");
    setVisible(false);
    onComplete?.();
  };

  const handleLoadDemoData = async () => {
    if (!accessToken) return;
    setLoading(true);
    setError(null);

    const headers = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    };

    try {
      // Step 1: Create Property
      setProgress("Creating property...");
      const propResp = await fetch(`${API_BASE}/api/v1/properties`, {
        method: "POST",
        headers,
        body: JSON.stringify(DEMO_PROPERTY),
      });
      if (!propResp.ok) throw new Error("Failed to create property");
      const propData = await propResp.json();
      const property = propData.data;

      // Step 2: Create Unit Types
      setProgress("Setting up room types...");
      const unitTypes: { id: string; name: string }[] = [];
      for (const ut of DEMO_UNIT_TYPES) {
        const resp = await fetch(`${API_BASE}/api/v1/properties/${property.id}/unit-types`, {
          method: "POST",
          headers,
          body: JSON.stringify(ut),
        });
        if (resp.ok) {
          const data = await resp.json();
          unitTypes.push(data.data);
        }
      }

      // Step 3: Create Units
      setProgress("Adding rooms...");
      const units: { id: string; unit_number: string }[] = [];
      for (const u of DEMO_UNITS) {
        const resp = await fetch(`${API_BASE}/api/v1/properties/${property.id}/units`, {
          method: "POST",
          headers,
          body: JSON.stringify({
            unit_number: u.unit_number,
            floor: String(u.floor),
            unit_type_id: unitTypes[u.type_index]?.id,
          }),
        });
        if (resp.ok) {
          const data = await resp.json();
          units.push(data.data);
        }
      }

      // Step 4: Create Guests
      setProgress("Adding sample guests...");
      const guests: { id: string }[] = [];
      for (const g of DEMO_GUESTS) {
        const resp = await fetch(`${API_BASE}/api/v1/guests`, {
          method: "POST",
          headers,
          body: JSON.stringify(g),
        });
        if (resp.ok) {
          const data = await resp.json();
          guests.push(data.data);
        }
      }

      // Step 5: Create sample reservations & operations lifecycle
      setProgress("Creating bookings & check-ins...");
      const today = new Date();
      const todayStr = today.toISOString().split("T")[0];

      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      const tomorrowStr = tomorrow.toISOString().split("T")[0];

      const dayAfterTomorrow = new Date();
      dayAfterTomorrow.setDate(dayAfterTomorrow.getDate() + 2);
      const dayAfterTomorrowStr = dayAfterTomorrow.toISOString().split("T")[0];

      const checkOutFuture = new Date();
      checkOutFuture.setDate(checkOutFuture.getDate() + 3);
      const checkOutFutureStr = checkOutFuture.toISOString().split("T")[0];

      const checkInFuture = new Date();
      checkInFuture.setDate(checkInFuture.getDate() + 4);
      const checkInFutureStr = checkInFuture.toISOString().split("T")[0];

      const checkOutFarFuture = new Date();
      checkOutFarFuture.setDate(checkOutFarFuture.getDate() + 7);
      const checkOutFarFutureStr = checkOutFarFuture.toISOString().split("T")[0];

      // Helper: past dates
      const daysAgo = (n: number) => {
        const d = new Date();
        d.setDate(d.getDate() - n);
        return d.toISOString().split("T")[0];
      };

      // Helper function to handle POST with logging
      const postJSON = async (url: string, dataObj: any) => {
        try {
          const resp = await fetch(url, {
            method: "POST",
            headers,
            body: JSON.stringify(dataObj),
          });
          if (!resp.ok) {
            const errBody = await resp.json().catch(() => ({}));
            console.warn(`POST ${url} failed:`, errBody);
            return null;
          }
          return await resp.json();
        } catch (e) {
          console.error(`POST ${url} error:`, e);
          return null;
        }
      };

      // Helper function to handle GET
      const getJSON = async (url: string) => {
        try {
          const resp = await fetch(url, { headers });
          if (!resp.ok) return null;
          return await resp.json();
        } catch (e) {
          return null;
        }
      };

      // ─────────────────────────────────────────────────────────────────
      // HISTORICAL RESERVATIONS (Past checkouts — builds up ₹1,00,000+ revenue)
      // ─────────────────────────────────────────────────────────────────
      setProgress("Loading historical revenue data...");

      // Guest 5 (Anita Desai) - Suite, 5 nights, 12 days ago → 7 days ago = ₹40,000
      const hist1Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[10]?.id, // Suite 301
        guest_id: guests[5]?.id,
        check_in_date: daysAgo(12),
        check_out_date: daysAgo(7),
        rate_per_night: 8000,
        booking_source: "booking_com",
        num_guests: 2,
        advance_amount: 8000,
        advance_method: "upi",
        advance_reference: "UPI-ADV-001",
        notes: "Corporate guest - repeat visitor",
      });
      if (hist1Data?.data) {
        const res = hist1Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res.id}/confirm`, {});
        await postJSON(`${API_BASE}/api/v1/operations/check-in`, {
          reservation_id: res.id,
          assigned_unit_id: units[10]?.id,
          deposit_amount: 5000,
          deposit_method: "card",
          id_document_type: "aadhaar",
        });
        const folioData = await getJSON(`${API_BASE}/api/v1/billing/folios/reservation/${res.id}`);
        const folio = folioData?.data || folioData;
        if (folio?.id) {
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "room_charge", description: "Room Charges - 5 Nights Suite", quantity: 5, unit_price: 8000 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "food_beverage", description: "Room Service - Dinner", quantity: 3, unit_price: 800 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "minibar", description: "Minibar Consumption", quantity: 1, unit_price: 1500 });
          await postJSON(`${API_BASE}/api/v1/billing/payments`, { folio_id: folio.id, payment_type: "payment", payment_method: "card", amount: 43900, reference_number: "TXN-HIST-001" });
          await postJSON(`${API_BASE}/api/v1/operations/check-out`, { reservation_id: res.id, payment_method: "card" });
        }
      }

      // Guest 6 (Suresh Nair) - Deluxe, 4 nights, 10 days ago → 6 days ago = ₹18,000
      const hist2Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[6]?.id, // Deluxe 201
        guest_id: guests[6]?.id,
        check_in_date: daysAgo(10),
        check_out_date: daysAgo(6),
        rate_per_night: 4500,
        booking_source: "phone",
        num_guests: 2,
        advance_amount: 4500,
        advance_method: "bank_transfer",
        advance_reference: "NEFT-9928374",
        notes: "Anniversary trip",
      });
      if (hist2Data?.data) {
        const res = hist2Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res.id}/confirm`, {});
        await postJSON(`${API_BASE}/api/v1/operations/check-in`, {
          reservation_id: res.id,
          assigned_unit_id: units[6]?.id,
          deposit_amount: 2000,
          deposit_method: "upi",
          id_document_type: "driving_license",
        });
        const folioData = await getJSON(`${API_BASE}/api/v1/billing/folios/reservation/${res.id}`);
        const folio = folioData?.data || folioData;
        if (folio?.id) {
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "room_charge", description: "Room Charges - 4 Nights Deluxe", quantity: 4, unit_price: 4500 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "spa", description: "Couple Spa Package", quantity: 1, unit_price: 3000 });
          await postJSON(`${API_BASE}/api/v1/billing/payments`, { folio_id: folio.id, payment_type: "payment", payment_method: "upi", amount: 21000, reference_number: "UPI-HIST-002" });
          await postJSON(`${API_BASE}/api/v1/operations/check-out`, { reservation_id: res.id, payment_method: "upi" });
        }
      }

      // Guest 7 (Kavita Joshi) - Family Room, 3 nights, 8 days ago → 5 days ago = ₹18,000
      const hist3Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[4]?.id, // Family 105
        guest_id: guests[7]?.id,
        check_in_date: daysAgo(8),
        check_out_date: daysAgo(5),
        rate_per_night: 6000,
        booking_source: "whatsapp",
        num_guests: 4,
        advance_amount: 6000,
        advance_method: "upi",
        advance_reference: "UPI-ADV-003",
        notes: "Family vacation with kids",
      });
      if (hist3Data?.data) {
        const res = hist3Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res.id}/confirm`, {});
        await postJSON(`${API_BASE}/api/v1/operations/check-in`, {
          reservation_id: res.id,
          assigned_unit_id: units[4]?.id,
          deposit_amount: 3000,
          deposit_method: "cash",
          id_document_type: "aadhaar",
        });
        const folioData = await getJSON(`${API_BASE}/api/v1/billing/folios/reservation/${res.id}`);
        const folio = folioData?.data || folioData;
        if (folio?.id) {
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "room_charge", description: "Room Charges - 3 Nights Family Room", quantity: 3, unit_price: 6000 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "extra_bed", description: "Extra Bed Charges", quantity: 3, unit_price: 500 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "food_beverage", description: "Breakfast Buffet (4 pax × 3 days)", quantity: 12, unit_price: 350 });
          await postJSON(`${API_BASE}/api/v1/billing/payments`, { folio_id: folio.id, payment_type: "payment", payment_method: "cash", amount: 23700, reference_number: "CASH-HIST-003" });
          await postJSON(`${API_BASE}/api/v1/operations/check-out`, { reservation_id: res.id, payment_method: "cash" });
        }
      }

      // Guest 1 (Rajesh) - historical stay, Standard, 3 nights, 15 days ago → 12 days ago = ₹7,500
      const hist4Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[2]?.id, // Standard 103
        guest_id: guests[0]?.id,
        check_in_date: daysAgo(15),
        check_out_date: daysAgo(12),
        rate_per_night: 2500,
        booking_source: "repeat",
        num_guests: 1,
        advance_amount: 2500,
        advance_method: "cash",
        advance_reference: "CASH-REC-101",
        notes: "Returning guest - loyal customer",
      });
      if (hist4Data?.data) {
        const res = hist4Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res.id}/confirm`, {});
        await postJSON(`${API_BASE}/api/v1/operations/check-in`, {
          reservation_id: res.id,
          assigned_unit_id: units[2]?.id,
          deposit_amount: 1000,
          deposit_method: "cash",
          id_document_type: "aadhaar",
        });
        const folioData = await getJSON(`${API_BASE}/api/v1/billing/folios/reservation/${res.id}`);
        const folio = folioData?.data || folioData;
        if (folio?.id) {
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "room_charge", description: "Room Charges - 3 Nights", quantity: 3, unit_price: 2500 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "laundry", description: "Laundry Service", quantity: 1, unit_price: 450 });
          await postJSON(`${API_BASE}/api/v1/billing/payments`, { folio_id: folio.id, payment_type: "payment", payment_method: "cash", amount: 7950, reference_number: "CASH-HIST-004" });
          await postJSON(`${API_BASE}/api/v1/operations/check-out`, { reservation_id: res.id, payment_method: "cash" });
        }
      }

      // Guest 3 (Amit) - historical stay, Deluxe, 2 nights, 5 days ago → 3 days ago = ₹9,000
      const hist5Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[7]?.id, // Deluxe 202
        guest_id: guests[2]?.id,
        check_in_date: daysAgo(5),
        check_out_date: daysAgo(3),
        rate_per_night: 4500,
        booking_source: "corporate",
        num_guests: 1,
        advance_amount: 9000,
        advance_method: "bank_transfer",
        advance_reference: "CORP-INV-2024-088",
        notes: "Corporate booking - TechCorp India",
      });
      if (hist5Data?.data) {
        const res = hist5Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res.id}/confirm`, {});
        await postJSON(`${API_BASE}/api/v1/operations/check-in`, {
          reservation_id: res.id,
          assigned_unit_id: units[7]?.id,
          deposit_amount: 0,
          id_document_type: "passport",
        });
        const folioData = await getJSON(`${API_BASE}/api/v1/billing/folios/reservation/${res.id}`);
        const folio = folioData?.data || folioData;
        if (folio?.id) {
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "room_charge", description: "Room Charges - 2 Nights Deluxe", quantity: 2, unit_price: 4500 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "parking", description: "Valet Parking", quantity: 2, unit_price: 200 });
          await postJSON(`${API_BASE}/api/v1/billing/payments`, { folio_id: folio.id, payment_type: "payment", payment_method: "bank_transfer", amount: 9400, reference_number: "NEFT-CORP-088" });
          await postJSON(`${API_BASE}/api/v1/operations/check-out`, { reservation_id: res.id, payment_method: "bank_transfer" });
        }
      }

      // ─────────────────────────────────────────────────────────────────
      // CURRENT / ACTIVE RESERVATIONS
      // ─────────────────────────────────────────────────────────────────
      setProgress("Creating active bookings...");

      // 1. Rajesh Sharma in Room 101 (Standard Room) - ACTIVE checked-in today
      const res1Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[0]?.id,
        guest_id: guests[0]?.id,
        check_in_date: todayStr,
        check_out_date: checkOutFutureStr,
        rate_per_night: 2500,
        booking_source: "walk_in",
        num_guests: 2,
        advance_amount: 2500,
        advance_method: "cash",
        advance_reference: "REC-2024-1102",
      });

      if (res1Data?.data) {
        const res1 = res1Data.data;
        // Confirm reservation
        await postJSON(`${API_BASE}/api/v1/reservations/${res1.id}/confirm`, {});
        
        // Check-in with deposit
        await postJSON(`${API_BASE}/api/v1/operations/check-in`, {
          reservation_id: res1.id,
          assigned_unit_id: units[0]?.id,
          deposit_amount: 1000,
          deposit_method: "cash",
          id_document_type: "aadhaar",
        });

        // Get Folio and add charges + payments
        const folioData = await getJSON(`${API_BASE}/api/v1/billing/folios/reservation/${res1.id}`);
        const folio = folioData?.data || folioData;
        if (folio?.id) {
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "room_charge", description: "Room Charge - Night 1", quantity: 1, unit_price: 2500 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "minibar", description: "Snacks & Refreshments", quantity: 1, unit_price: 350 });
          await postJSON(`${API_BASE}/api/v1/billing/payments`, { folio_id: folio.id, payment_type: "payment", payment_method: "upi", amount: 2850, reference_number: "UPI-10293847" });
        }
      }

      // 2. Priya Patel in Room 201 (Deluxe Room) - ARRIVAL tomorrow
      const res2Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[6]?.id,
        guest_id: guests[1]?.id,
        check_in_date: tomorrowStr,
        check_out_date: checkOutFutureStr,
        rate_per_night: 4500,
        booking_source: "ota_makemytrip",
        num_guests: 2,
        advance_amount: 4500,
        advance_method: "card",
        advance_reference: "MMT-BK-9928374",
      });
      if (res2Data?.data) {
        const res2 = res2Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res2.id}/confirm`, {});
      }

      // 3. Amit Kumar in Room 302 (Suite) - CHECKED OUT today
      const res3Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[11]?.id,
        guest_id: guests[2]?.id,
        check_in_date: todayStr,
        check_out_date: tomorrowStr,
        rate_per_night: 8000,
        booking_source: "website",
        num_guests: 3,
        advance_amount: 8000,
        advance_method: "card",
        advance_reference: "WEB-PAY-554433",
      });
      if (res3Data?.data) {
        const res3 = res3Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res3.id}/confirm`, {});
        await postJSON(`${API_BASE}/api/v1/operations/check-in`, {
          reservation_id: res3.id,
          assigned_unit_id: units[11]?.id,
          deposit_amount: 2000,
          deposit_method: "card",
          id_document_type: "driving_license",
        });

        const folioData = await getJSON(`${API_BASE}/api/v1/billing/folios/reservation/${res3.id}`);
        const folio = folioData?.data || folioData;
        if (folio?.id) {
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "room_charge", description: "Room Charge - 1 Night Suite", quantity: 1, unit_price: 8000 });
          await postJSON(`${API_BASE}/api/v1/billing/charges`, { folio_id: folio.id, category: "food_beverage", description: "In-Room Dining", quantity: 1, unit_price: 1200 });
          await postJSON(`${API_BASE}/api/v1/billing/payments`, { folio_id: folio.id, payment_type: "payment", payment_method: "card", amount: 9200, reference_number: "TXN-554433" });
          await postJSON(`${API_BASE}/api/v1/operations/check-out`, { reservation_id: res3.id, payment_method: "card" });
        }
      }

      // 4. Sneha Reddy in Room 102 (Standard Room) - FUTURE booking
      const res4Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[1]?.id,
        guest_id: guests[3]?.id,
        check_in_date: checkInFutureStr,
        check_out_date: checkOutFarFutureStr,
        rate_per_night: 2500,
        booking_source: "airbnb",
        num_guests: 1,
        advance_amount: 5000,
        advance_method: "upi",
        advance_reference: "ABNB-RES-88271",
      });
      if (res4Data?.data) {
        const res4 = res4Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res4.id}/confirm`, {});
      }

      // 5. Vikram Singh in Room 203 (Deluxe Room) - ARRIVAL today (Pending Check-in)
      const res5Data = await postJSON(`${API_BASE}/api/v1/reservations`, {
        property_id: property.id,
        unit_id: units[8]?.id,
        guest_id: guests[4]?.id,
        check_in_date: todayStr,
        check_out_date: dayAfterTomorrowStr,
        rate_per_night: 4500,
        booking_source: "walk_in",
        num_guests: 2,
        advance_amount: 4500,
        advance_method: "cash",
        advance_reference: "REC-2024-1105",
      });
      if (res5Data?.data) {
        const res5 = res5Data.data;
        await postJSON(`${API_BASE}/api/v1/reservations/${res5.id}/confirm`, {});
      }

      // Step 6: Create Housekeeping Tasks
      setProgress("Generating housekeeping schedules...");
      // Task 1: Room 103 (Standard Room) - Dirty
      await postJSON(`${API_BASE}/api/v1/housekeeping/tasks`, {
        property_id: property.id,
        unit_id: units[2]?.id,
        task_type: "stay_clean",
        priority: "medium",
        notes: "Routine deep clean after checkout"
      });

      // Task 2: Room 204 (Deluxe Room) - Cleaning in progress
      const hk2Data = await postJSON(`${API_BASE}/api/v1/housekeeping/tasks`, {
        property_id: property.id,
        unit_id: units[9]?.id,
        task_type: "stay_clean",
        priority: "high",
        notes: "Pre-arrival deep clean for VIP guest"
      });
      if (hk2Data?.data?.id) {
        await postJSON(`${API_BASE}/api/v1/housekeeping/tasks/status`, { task_id: hk2Data.data.id, status: "cleaning" });
      }

      // Task 3: Room 104 (Standard Room) - Ready / inspected
      const hk3Data = await postJSON(`${API_BASE}/api/v1/housekeeping/tasks`, {
        property_id: property.id,
        unit_id: units[3]?.id,
        task_type: "stay_clean",
        priority: "low",
        notes: "Routine light cleaning completed"
      });
      if (hk3Data?.data?.id) {
        await postJSON(`${API_BASE}/api/v1/housekeeping/tasks/status`, { task_id: hk3Data.data.id, status: "ready" });
      }

      // Task 4: Suite 302 - Deep cleaning after checkout
      await postJSON(`${API_BASE}/api/v1/housekeeping/tasks`, {
        property_id: property.id,
        unit_id: units[11]?.id,
        task_type: "checkout_clean",
        priority: "high",
        notes: "Post-checkout deep clean - Suite"
      });

      // Step 7: Create Laundry orders & rate card
      setProgress("Generating laundry orders...");

      // Create rate card for the property
      await postJSON(`${API_BASE}/api/v1/laundry/rate-card`, {
        property_id: property.id,
        item_type: "shirt",
        service_type: "wash_iron",
        price: 70,
      });
      await postJSON(`${API_BASE}/api/v1/laundry/rate-card`, {
        property_id: property.id,
        item_type: "trouser",
        service_type: "wash_iron",
        price: 80,
      });
      await postJSON(`${API_BASE}/api/v1/laundry/rate-card`, {
        property_id: property.id,
        item_type: "suit",
        service_type: "dry_clean",
        price: 250,
      });
      await postJSON(`${API_BASE}/api/v1/laundry/rate-card`, {
        property_id: property.id,
        item_type: "saree",
        service_type: "dry_clean",
        price: 200,
      });

      // Laundry order for Rajesh (active guest)
      await postJSON(`${API_BASE}/api/v1/laundry/orders`, {
        property_id: property.id,
        guest_id: guests[0]?.id,
        order_type: "guest",
        items: [
          { item_type: "shirt", service_type: "wash_iron", quantity: 3, unit_price: 70 },
          { item_type: "trouser", service_type: "wash_iron", quantity: 2, unit_price: 80 },
          { item_type: "suit", service_type: "dry_clean", quantity: 1, unit_price: 250 },
        ],
        notes: "Express delivery by 6 PM",
      });

      // Step 8: Create notification template
      setProgress("Setting up notifications...");
      await postJSON(`${API_BASE}/api/v1/notifications/templates`, {
        event_type: "booking_confirmation",
        channel: "whatsapp",
        template_body: "Hello {{guest_name}}, your booking at {{property_name}} is confirmed! Check-in: {{check_in_date}}, Check-out: {{check_out_date}}. Booking ID: {{booking_id}}",
        is_active: true,
      });
      await postJSON(`${API_BASE}/api/v1/notifications/templates`, {
        event_type: "check_in_welcome",
        channel: "whatsapp",
        template_body: "Welcome to {{property_name}}, {{guest_name}}! You are in Room {{room_number}}. WiFi: StayFlow-Guest / Password: guest2024. For any assistance, call the front desk.",
        is_active: true,
      });

      setProgress("");
      setDone(true);
      localStorage.setItem("demo_data_shown", "true");
      localStorage.setItem("demo_data_loaded", "true");

      // Auto-close after showing success
      setTimeout(() => {
        setVisible(false);
        onComplete?.();
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load demo data");
      setLoading(false);
    }
  };

  if (!visible) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" data-testid="demo-data-dialog">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4 p-6">
        <div className="text-center">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-blue-100 mb-4">
            <svg className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />
            </svg>
          </div>

          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            Welcome to StayFlow!
          </h3>

          {!loading && !done && (
            <>
              <p className="text-sm text-gray-600 mb-6">
                Would you like to load demo data? This will create a sample hotel with rooms,
                guests, and bookings so you can explore all features immediately.
              </p>

              <div className="bg-gray-50 rounded-md p-3 mb-6 text-left">
                <p className="text-xs font-medium text-gray-700 mb-2">Demo data includes:</p>
                <ul className="text-xs text-gray-600 space-y-1">
                  <li>• 1 Property with 12 rooms across 4 room types</li>
                  <li>• 8 Sample guests with full profiles</li>
                  <li>• 10 Bookings with ₹1,00,000+ revenue history</li>
                  <li>• Complete billing: charges, payments, advances & deposits</li>
                  <li>• Housekeeping tasks (Dirty, Cleaning, Ready states)</li>
                  <li>• Laundry orders with rate cards</li>
                  <li>• WhatsApp notification templates</li>
                </ul>
              </div>

              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                  {error}
                </div>
              )}

              <div className="flex gap-3">
                <button
                  onClick={handleSkip}
                  data-testid="skip-demo-data"
                  className="flex-1 rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Skip
                </button>
                <button
                  onClick={handleLoadDemoData}
                  data-testid="load-demo-data"
                  className="flex-1 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
                >
                  Load Demo Data
                </button>
              </div>
            </>
          )}

          {loading && !done && (
            <div data-testid="demo-loading">
              <div className="flex items-center justify-center mb-4">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
              <p className="text-sm text-gray-600">{progress}</p>
            </div>
          )}

          {done && (
            <div data-testid="demo-success">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-green-100 mb-4">
                <svg className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                </svg>
              </div>
              <p className="text-sm text-green-700 font-medium">Demo data loaded successfully!</p>
              <p className="text-xs text-gray-500 mt-1">Redirecting to dashboard...</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
