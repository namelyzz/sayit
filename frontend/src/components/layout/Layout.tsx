"use client";

import type { ReactNode } from "react";
import Navbar from "./Navbar";
import Sidebar from "./Sidebar";
import RightSidebar from "./RightSidebar";

interface LayoutProps {
  children: ReactNode;
  showSidebar?: boolean;
  showRightSidebar?: boolean;
}

export default function Layout({
  children,
  showSidebar = true,
  showRightSidebar = true,
}: LayoutProps) {
  return (
    <div className="min-h-screen">
      <Navbar />

      <div className="mx-auto grid w-full max-w-[1440px] grid-cols-1 gap-6 px-4 pb-24 pt-5 md:px-6 lg:grid-cols-[248px_minmax(0,1fr)] lg:pb-10 xl:grid-cols-[248px_minmax(0,760px)_320px]">
        {showSidebar ? <Sidebar /> : null}

        <main className="min-w-0">{children}</main>

        {showRightSidebar ? <RightSidebar /> : null}
      </div>
    </div>
  );
}
