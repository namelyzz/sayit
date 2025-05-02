"use client";

import Link from "next/link";
import { Award, MessageSquareText, PenLine, ShieldCheck, UserRound } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import Badge from "@/components/ui/Badge";

export default function RightSidebar() {
  const { user } = useAuth();

  return (
    <aside className="sticky top-21 hidden h-[calc(100vh-5.25rem)] overflow-y-auto xl:block scrollbar-thin">
      <div className="space-y-4">
        <section className="rounded-lg border border-border bg-surface p-5 shadow-sm">
          {user ? (
            <>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary text-white">
                  <UserRound className="h-6 w-6" />
                </div>
                <div className="min-w-0">
                  <p className="truncate text-base font-semibold text-foreground">{user.user_name}</p>
                  <p className="truncate text-xs text-muted">ID {user.user_id}</p>
                </div>
              </div>
              <div className="mt-5 grid grid-cols-2 gap-3">
                <div className="rounded-lg bg-surface-soft p-3">
                  <p className="text-lg font-bold text-foreground">0</p>
                  <p className="text-xs text-muted">发布</p>
                </div>
                <div className="rounded-lg bg-surface-soft p-3">
                  <p className="text-lg font-bold text-foreground">0</p>
                  <p className="text-xs text-muted">关注者</p>
                </div>
              </div>
              <Link
                href="/submit"
                className="mt-4 inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-primary text-sm font-medium text-white transition hover:bg-primary-dark"
              >
                <PenLine className="h-4 w-4" />
                写一篇帖子
              </Link>
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
