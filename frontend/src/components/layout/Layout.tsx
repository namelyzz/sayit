"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";
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
  const pathname = usePathname();
  const isUserPage = pathname.startsWith("/user/");
  const shouldShowRightSidebar = showRightSidebar && !isUserPage;

  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto min-h-screen max-w-[1440px] bg-chrome">
        <Navbar />

        <div className="grid grid-cols-1 lg:grid-cols-[292px_minmax(0,1fr)]">
          {showSidebar ? <Sidebar /> : null}

          <div className="min-w-0 bg-background px-4 pb-24 pt-5 md:px-6 lg:min-h-[calc(100vh-4rem)] lg:rounded-tl-[24px] lg:pb-10">
            <div
              className={`grid w-full gap-6 ${
                shouldShowRightSidebar
                  ? "xl:grid-cols-[minmax(0,1fr)_360px]"
                  : "mx-auto max-w-[1120px]"
              }`}
            >
              <main className="min-w-0">{children}</main>
              {shouldShowRightSidebar ? <RightSidebar /> : null}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
