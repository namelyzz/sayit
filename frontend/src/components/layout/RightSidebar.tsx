"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Award, ChevronRight, MessageSquareText, ShieldCheck } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import Badge from "@/components/ui/Badge";
import { apiClient, type UserProfile } from "@/lib/api";
import { formatCount } from "@/lib/format";
import { previewFollowing } from "@/lib/user-preview";

function StatCard({ value, label }: { value: string; label: string }) {
  return (
    <div className="rounded-lg bg-surface-soft p-3 text-center">
      <p className="text-lg font-bold text-foreground">{value}</p>
      <p className="mt-1 text-xs text-muted">{label}</p>
    </div>
  );
}

export default function RightSidebar() {
  const { user, loading: authLoading } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);

  useEffect(() => {
    if (authLoading || !user) {
      if (!user) setProfile(null);
      return;
    }

    const fetchProfile = async () => {
      try {
        const response = await apiClient.getUserProfile(user.user_id);
        setProfile(response.data);
      } catch {
        setProfile(null);
      }
    };

    fetchProfile();
  }, [authLoading, user]);

  return (
    <aside className="sticky top-5 hidden h-[calc(100vh-1.25rem)] overflow-y-auto xl:block scrollbar-thin">
      <div className="space-y-4">
        <section className="rounded-lg border border-border bg-surface p-5 shadow-sm">
          {user ? (
            <>
              <Link
                href={`/user/${user.user_id}`}
                className="group block rounded-lg border border-transparent p-1 transition hover:border-border hover:bg-surface-soft/70"
                aria-label="进入个人中心"
                title="进入个人中心"
              >
                <div className="flex items-center gap-3">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary text-white ring-2 ring-white shadow-sm transition group-hover:scale-[1.03]">
                    <span className="text-lg font-semibold">{user.user_name.slice(0, 1).toUpperCase()}</span>
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <p className="truncate text-base font-semibold text-foreground">{user.user_name}</p>
                      <span className="inline-flex items-center rounded-md bg-[#457b9d]/10 px-2 py-0.5 text-[11px] font-medium text-primary">
                        个人中心
                      </span>
                    </div>
                    <p className="mt-1 text-xs text-muted">点击头像或昵称进入管理页面</p>
                  </div>
                  <ChevronRight className="h-4 w-4 shrink-0 text-slate-400 transition group-hover:text-primary" />
                </div>
              </Link>

              <div className="mt-5 grid grid-cols-3 gap-3">
                <StatCard value={formatCount(profile?.post_count ?? 0)} label="发布" />
                <StatCard value={formatCount(profile?.post_score ?? 0)} label="热度" />
                <StatCard value={formatCount(previewFollowing.length)} label="关注" />
              </div>
            </>
          ) : (
            <>
              <Badge tone="primary">欢迎来到 SayIt</Badge>
              <h2 className="mt-3 text-lg font-semibold text-foreground">把想法晒出来，让讨论慢慢长出来。</h2>
              <p className="mt-2 text-sm leading-6 text-muted">
                登录后可以发布帖子、关注社区、参与投票和评论。
              </p>
              <div className="mt-4 grid grid-cols-2 gap-2">
                <Link
                  href="/login"
                  className="inline-flex h-10 items-center justify-center rounded-lg bg-primary text-sm font-medium text-white transition hover:bg-primary-dark"
                >
                  登录
                </Link>
                <Link
                  href="/signup"
                  className="inline-flex h-10 items-center justify-center rounded-lg border border-border bg-surface text-sm font-medium text-muted-strong transition hover:bg-surface-soft"
                >
                  注册
                </Link>
              </div>
            </>
          )}
        </section>

        <section className="rounded-lg border border-border bg-surface p-5 shadow-sm">
          <div className="flex items-center gap-2">
            <MessageSquareText className="h-5 w-5 text-primary" />
            <h2 className="text-sm font-semibold text-foreground">社区公告</h2>
          </div>
          <div className="mt-4 space-y-4">
            <div>
              <div className="mb-1 flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-primary" />
                <span className="text-xs font-semibold text-primary">进行中</span>
              </div>
              <p className="text-sm font-medium text-foreground">夏季摄影与生活记录征集</p>
              <p className="mt-1 text-xs text-muted">分享一张照片和背后的故事。</p>
            </div>
            <div>
              <div className="mb-1 flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-accent" />
                <span className="text-xs font-semibold text-accent">即将开始</span>
              </div>
              <p className="text-sm font-medium text-foreground">独立开发者圆桌 AMA</p>
              <p className="mt-1 text-xs text-muted">明天 20:00 开放提问。</p>
            </div>
          </div>
        </section>

        <section className="rounded-lg border border-border bg-surface p-5 shadow-sm">
          <div className="flex items-center gap-2">
            <Award className="h-5 w-5 text-accent" />
            <h2 className="text-sm font-semibold text-foreground">社区规则</h2>
          </div>
          <ul className="mt-4 space-y-3 text-sm leading-6 text-muted-strong">
            <li className="flex gap-2">
              <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
              尊重他人，表达观点时保留善意。
            </li>
            <li className="flex gap-2">
              <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
              不发布垃圾广告、恶意引战或违法内容。
            </li>
            <li className="flex gap-2">
              <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
              保护隐私，不公开他人的敏感信息。
            </li>
          </ul>
        </section>

        <footer className="px-2 text-center text-xs text-muted">
          <p className="rainbow-text text-lg font-bold leading-7">晒一个有意思的灵魂</p>
          <div className="mt-2 flex justify-center gap-4">
            <Link href="/about" className="hover:text-primary">
              关于
            </Link>
            <Link href="/help" className="hover:text-primary">
              帮助
            </Link>
          </div>
        </footer>
      </div>
    </aside>
  );
}
