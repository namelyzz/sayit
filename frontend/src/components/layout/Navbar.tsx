"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Bell, ChevronRight, Compass, Home, LogIn, LogOut, PenLine, Search } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import Button from "@/components/ui/Button";
import { apiClient } from "@/lib/api";
import { cn } from "@/lib/utils";

const notificationUnreadChangedEvent = "notification-unread-changed";

const mobileLinks = [
  { href: "/", label: "首页", icon: Home },
  { href: "/submit", label: "发布", icon: PenLine },
  { href: "/about", label: "关于", icon: Compass },
];

export default function Navbar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const [unreadCount, setUnreadCount] = useState(0);
  const avatarText = user?.user_name?.slice(0, 1).toUpperCase() || "S";
  const unreadLabel = unreadCount > 99 ? "99+" : String(unreadCount);

  const refreshUnreadCount = useCallback(async () => {
    if (!user) {
      setUnreadCount(0);
      return;
    }
    try {
      const response = await apiClient.getNotificationUnreadCount();
      setUnreadCount(response.data.count || 0);
    } catch {
      setUnreadCount(0);
    }
  }, [user]);

  useEffect(() => {
    refreshUnreadCount();
    if (!user) return;

    const interval = window.setInterval(refreshUnreadCount, 45000);
    const handleFocus = () => refreshUnreadCount();
    const handleUnreadChanged = (event: Event) => {
      const detail = (event as CustomEvent<number>).detail;
      if (typeof detail === "number") {
        setUnreadCount(Math.max(0, detail));
        return;
      }
      refreshUnreadCount();
    };
    window.addEventListener("focus", handleFocus);
    window.addEventListener(notificationUnreadChangedEvent, handleUnreadChanged);

    return () => {
      window.clearInterval(interval);
      window.removeEventListener("focus", handleFocus);
      window.removeEventListener(notificationUnreadChangedEvent, handleUnreadChanged);
    };
  }, [refreshUnreadCount, user]);

  return (
    <>
      <header className="z-30 bg-transparent">
        <div className="flex h-16 items-center justify-between gap-4 px-5 md:px-7">
          <Link href="/" className="flex shrink-0 items-center gap-3">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-base font-bold text-white shadow-sm">
              S
            </span>
            <span className="leading-tight">
              <span className="block text-base font-bold tracking-normal text-slate-950">SayIt</span>
              <span className="hidden text-xs font-medium text-slate-800/80 sm:block">晒意</span>
            </span>
          </Link>

          <form
            className="hidden min-w-0 flex-1 justify-center md:flex"
            role="search"
            onSubmit={(event) => event.preventDefault()}
          >
            <div className="relative w-full max-w-xl">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
              <input
                type="search"
                aria-label="搜索帖子或社区"
                placeholder="搜索帖子或社区"
                className="h-10 w-full rounded-lg bg-white/90 pl-9 pr-3 text-sm text-foreground shadow-sm transition placeholder:text-slate-500 focus:bg-white focus:outline-none focus:ring-4 focus:ring-white/50"
              />
            </div>
          </form>

          <div className="flex shrink-0 items-center gap-2">
            {user ? (
              <Link
                href="/notifications"
                className="relative hidden h-10 w-10 items-center justify-center rounded-lg text-slate-800 transition hover:bg-white/35 md:inline-flex"
                aria-label={unreadCount > 0 ? `通知，${unreadCount} 条未读` : "通知"}
                title="通知"
              >
                <Bell className="h-5 w-5" />
                {unreadCount > 0 ? (
                  <span className="absolute -right-1 -top-1 min-w-5 rounded-full bg-danger px-1.5 py-0.5 text-center text-[10px] font-bold leading-4 text-white shadow-sm ring-2 ring-white/80">
                    {unreadLabel}
                  </span>
                ) : null}
              </Link>
            ) : null}

            <Link
              href="/submit"
              className="hidden h-10 items-center justify-center gap-2 rounded-lg bg-primary px-4 text-sm font-medium text-white shadow-sm transition hover:bg-primary-dark md:inline-flex"
            >
              <PenLine className="h-4 w-4" />
              发布
            </Link>

            {user ? (
              <div className="flex items-center gap-2">
                <Link
                  href={`/user/${user.user_id}`}
                  className="group inline-flex h-10 items-center gap-2 rounded-lg border border-white/60 bg-white/55 px-2.5 pr-3 text-sm font-medium text-slate-900 shadow-sm transition hover:border-white/80 hover:bg-white/80 hover:shadow-md"
                  aria-label="进入个人中心"
                  title="进入个人中心"
                >
                  <span className="flex h-7 w-7 items-center justify-center rounded-full bg-primary text-xs font-semibold text-white ring-2 ring-white/70 transition group-hover:scale-[1.04]">
                    {avatarText}
                  </span>
                  <span className="hidden max-w-28 truncate md:inline">{user.user_name}</span>
                  <ChevronRight className="hidden h-4 w-4 text-slate-500 transition group-hover:text-slate-700 md:inline" />
                </Link>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={logout}
                  className="hidden text-slate-800 hover:bg-white/35 md:inline-flex"
                  aria-label="退出登录"
                >
                  <LogOut className="h-5 w-5" />
                </Button>
              </div>
            ) : (
              <Link
                href="/login"
                className="inline-flex h-10 items-center justify-center gap-2 rounded-lg bg-white/45 px-3 text-sm font-medium text-slate-900 transition hover:bg-white/60"
              >
                <LogIn className="h-4 w-4" />
                登录
              </Link>
            )}
          </div>
        </div>
      </header>

      <nav className="fixed inset-x-0 bottom-0 z-40 bg-surface/95 px-4 pb-3 pt-2 backdrop-blur-xl md:hidden">
        <div className="mx-auto grid max-w-md grid-cols-3 gap-1">
          {mobileLinks.map((item) => {
            const Icon = item.icon;
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "flex flex-col items-center gap-1 rounded-lg px-2 py-2 text-xs font-medium transition",
                  active ? "bg-[#e0eef8] text-primary" : "text-muted hover:bg-surface-soft"
                )}
              >
                <Icon className="h-5 w-5" />
                {item.label}
              </Link>
            );
          })}
        </div>
      </nav>
    </>
  );
}
