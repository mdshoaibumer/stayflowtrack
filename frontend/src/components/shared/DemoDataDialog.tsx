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
];

const DEMO_UNITS = [
  { unit_number: "101", floor: 1, type_index: 0 },
  { unit_number: "102", floor: 1, type_index: 0 },
  { unit_number: "103", floor: 1, type_index: 0 },
  { unit_number: "201", floor: 2, type_index: 1 },
  { unit_number: "202", floor: 2, type_index: 1 },
  { unit_number: "203", floor: 2, type_index: 1 },
  { unit_number: "301", floor: 3, type_index: 2 },
  { unit_number: "302", floor: 3, type_index: 2 },
];

const DEMO_GUESTS = [
  { first_name: "Rajesh", last_name: "Sharma", phone: "+919876543001", email: "rajesh.sharma@email.com", city: "Delhi", state: "Delhi", country: "India" },
  { first_name: "Priya", last_name: "Patel", phone: "+919876543002", email: "priya.patel@email.com", city: "Ahmedabad", state: "Gujarat", country: "India" },
  { first_name: "Amit", last_name: "Kumar", phone: "+919876543003", email: "amit.kumar@email.com", city: "Bangalore", state: "Karnataka", country: "India" },
  { first_name: "Sneha", last_name: "Reddy", phone: "+919876543004", email: "sneha.reddy@email.com", city: "Hyderabad", state: "Telangana", country: "India" },
  { first_name: "Vikram", last_name: "Singh", phone: "+919876543005", email: "vikram.singh@email.com", city: "Jaipur", state: "Rajasthan", country: "India" },
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
    // Show dialog only for first-time users (check localStorage flag)
    const demoShown = localStorage.getItem("demo_data_shown");
    if (!demoShown && accessToken) {
      setVisible(true);
    }
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
      for (const u of DEMO_UNITS) {
        await fetch(`${API_BASE}/api/v1/properties/${property.id}/units`, {
          method: "POST",
          headers,
          body: JSON.stringify({
            unit_number: u.unit_number,
            floor: u.floor,
            unit_type_id: unitTypes[u.type_index]?.id,
          }),
        });
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

      // Step 5: Create sample reservations
      setProgress("Creating sample bookings...");
      const today = new Date();
      for (let i = 0; i < Math.min(3, guests.length); i++) {
        const checkIn = new Date(today);
        checkIn.setDate(today.getDate() + i + 1);
        const checkOut = new Date(checkIn);
        checkOut.setDate(checkIn.getDate() + 3);

        await fetch(`${API_BASE}/api/v1/reservations`, {
          method: "POST",
          headers,
          body: JSON.stringify({
            property_id: property.id,
            unit_id: DEMO_UNITS[i] ? undefined : undefined, // Will need unit IDs
            guest_id: guests[i]?.id,
            check_in_date: checkIn.toISOString().split("T")[0],
            check_out_date: checkOut.toISOString().split("T")[0],
            rate_per_night: DEMO_UNIT_TYPES[0].base_rate,
            booking_source: "walk_in",
            adults: 2,
            children: 0,
          }),
        });
      }

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
                  <li>• 1 Hotel property (Grand StayFlow Hotel)</li>
                  <li>• 3 Room types (Standard, Deluxe, Suite)</li>
                  <li>• 8 Rooms across 3 floors</li>
                  <li>• 5 Sample guests with details</li>
                  <li>• 3 Upcoming reservations</li>
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
