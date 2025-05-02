"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Bell, Compass, Home, LogIn, LogOut, PenLine, Search, UserRound } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import Button from "@/components/ui/Button";
import { cn } from "@/lib/utils";

const mobileLinks = [
  { href: "/", label: "首页", icon: Home },
  { href: "/submit", label: "发布", icon: PenLine },
  { href: "/about", label: "关于", icon: Compass },
];

export default function Navbar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  return (
    <>
      <header className="sticky top-0 z-40 border-b border-border bg-surface/90 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-[1440px] items-center justify-between gap-4 px-4 md:px-6">
          <Link href="/" className="flex shrink-0 items-center gap-3">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-base font-bold text-white shadow-sm">
              S
            </span>
            <span className="leading-tight">
              <span className="block text-base font-bold tracking-normal text-foreground">SayIt</span>
              <span className="hidden text-xs text-muted sm:block">晒意</span>
            </span>
          </Link>

          <form
            className="hidden min-w-0 flex-1 justify-center md:flex"
            role="search"
            onSubmit={(event) => event.preventDefault()}
          >
            <div className="relative w-full max-w-xl">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
              <input
                type="search"
                aria-label="搜索帖子或社区"
                placeholder="搜索帖子或社区"
                className="h-10 w-full rounded-lg border border-border bg-surface-soft pl-9 pr-3 text-sm text-foreground transition placeholder:text-muted focus:border-primary focus:bg-surface focus:outline-none focus:ring-4 focus:ring-teal-100"
              />
            </div>
          </form>

          <div className="flex shrink-0 items-center gap-2">
            <Button variant="ghost" size="icon" className="hidden md:inline-flex" aria-label="通知">
              <Bell className="h-5 w-5" />
            </Button>

            <Link
              href="/submit"
              className="hidden h-10 items-center justify-center gap-2 rounded-lg bg-primary px-4 text-sm font-medium text-white shadow-sm transition hover:bg-primary-dark md:inline-flex"
            >
              <PenLine className="h-4 w-4" />
              发布
            </Link>

            {user ? (
              <div className="hidden items-center gap-2 md:flex">
                <Link
                  href={`/user/${user.user_id}`}
                  className="inline-flex h-10 items-center gap-2 rounded-lg bg-surface-soft px-3 text-sm font-medium text-muted-strong transition hover:bg-surface-muted"
                >
                  <UserRound className="h-4 w-4" />
                  {user.user_name}
                </Link>
                <Button variant="ghost" size="icon" onClick={logout} aria-label="退出登录">
                  <LogOut className="h-5 w-5" />
                </Button>
              </div>
            ) : (
              <Link
                href="/login"
                className="inline-flex h-10 items-center justify-center gap-2 rounded-lg border border-border bg-surface px-3 text-sm font-medium text-muted-strong transition hover:bg-surface-soft"
              >
                <LogIn className="h-4 w-4" />
                登录
              </Link>
            )}
          </div>
        </div>
      </header>

      <nav className="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-surface/95 px-4 pb-3 pt-2 backdrop-blur-xl md:hidden">
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
                  active ? "bg-teal-50 text-primary" : "text-muted hover:bg-surface-soft"
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
