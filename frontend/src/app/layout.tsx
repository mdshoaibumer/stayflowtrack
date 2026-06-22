import type { Metadata, Viewport } from "next";
import { Inter, Karla } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/contexts/AuthContext";
import { ToastProvider } from "@/components/shared/Toast";

const inter = Inter({ subsets: ["latin"], display: "swap", variable: "--font-inter" });
const karla = Karla({ subsets: ["latin"], display: "swap", variable: "--font-karla" });

export const metadata: Metadata = {
  title: {
    default: "StayFlow — Property Management Platform",
    template: "%s | StayFlow",
  },
  description: "Reservations, billing, housekeeping, laundry, and operations — all in one platform for service apartments and boutique hotels.",
  keywords: ["property management", "hotel management", "service apartments", "PMS", "billing", "housekeeping"],
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  themeColor: "#1E3A5F",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${inter.variable} ${karla.variable}`}>
      <body className={`${karla.className} antialiased`}>
        <a href="#main-content" className="skip-link">Skip to content</a>
        <AuthProvider>
          <ToastProvider>{children}</ToastProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
