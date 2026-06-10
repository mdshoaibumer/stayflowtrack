"use client";

import { useState } from "react";
import ProtectedRoute from "@/components/auth/ProtectedRoute";
import Sidebar from "@/components/layout/Sidebar";
import TopNav from "@/components/layout/TopNav";
import DemoDataDialog from "@/components/shared/DemoDataDialog";
import CommandPalette from "@/components/shared/CommandPalette";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <ProtectedRoute>
      <div className="flex h-screen overflow-hidden bg-[hsl(210,40%,98%)]">
        {/* Mobile overlay */}
        {mobileMenuOpen && (
          <div
            className="fixed inset-0 bg-black/40 backdrop-blur-sm z-40 lg:hidden transition-opacity duration-200"
            onClick={() => setMobileMenuOpen(false)}
          />
        )}

        {/* Sidebar - desktop */}
        <div className="hidden lg:flex">
          <Sidebar
            collapsed={sidebarCollapsed}
            onToggle={() => setSidebarCollapsed(!sidebarCollapsed)}
          />
        </div>

        {/* Sidebar - mobile */}
        <div
          className={`fixed inset-y-0 left-0 z-50 lg:hidden transition-transform duration-200 ${
            mobileMenuOpen ? "translate-x-0" : "-translate-x-full"
          }`}
        >
          <Sidebar
            collapsed={false}
            onToggle={() => setMobileMenuOpen(false)}
            onMobileClose={() => setMobileMenuOpen(false)}
          />
        </div>

        {/* Main content */}
        <div className="flex-1 flex flex-col min-w-0">
          <TopNav onMenuToggle={() => setMobileMenuOpen(!mobileMenuOpen)} />
          <main id="main-content" className="flex-1 overflow-y-auto p-4 lg:p-6 scroll-smooth">
            {children}
          </main>
        </div>

        {/* Demo data dialog for new users */}
        <DemoDataDialog />
        <CommandPalette />
      </div>
    </ProtectedRoute>
  );
}
