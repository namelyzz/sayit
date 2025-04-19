"use client";

import Navbar from "./Navbar";
import Sidebar from "./Sidebar";
import RightSidebar from "./RightSidebar";

interface LayoutProps {
  children: React.ReactNode;
  showSidebar?: boolean;
  showRightSidebar?: boolean;
}

export default function Layout({
  children,
  showSidebar = true,
  showRightSidebar = true,
}: LayoutProps) {
  return (
    <div className="min-h-screen bg-gray-50">
      <Navbar />
      
      <div className="flex">
        {/* Left Sidebar */}
        {showSidebar && <Sidebar />}
        
        {/* Main Content */}
        <main className="flex-1 min-w-0">
          <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
            {children}
          </div>
        </main>
        
        {/* Right Sidebar */}
        {showRightSidebar && <RightSidebar />}
      </div>
    </div>
  );
}
