"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Bell, BellRing, CheckCheck, Heart, MessageSquareReply, ThumbsDown, ThumbsUp, UserPlus } from "lucide-react";
import Button from "@/components/ui/Button";
import EmptyState from "@/components/ui/EmptyState";
import PageShell from "@/components/ui/PageShell";
import { Skeleton } from "@/components/ui/Skeleton";
import { useAuth } from "@/context/AuthContext";
import { apiClient, type NotificationItem } from "@/lib/api";
import { formatDateTime } from "@/lib/format";
import { cn, getErrorMessage } from "@/lib/utils";

type NotificationStatus = "all" | "unread";

const PAGE_SIZE = 20;
const notificationUnreadChangedEvent = "notification-unread-changed";

function notificationIcon(item: NotificationItem) {
  switch (item.type) {
    case "comment_liked":
      return <Heart className="h-4 w-4" />;
    case "post_commented":
      return <MessageSquareReply className="h-4 w-4" />;
    case "post_voted":
      return item.direction && item.direction < 0 ? <ThumbsDown className="h-4 w-4" /> : <ThumbsUp className="h-4 w-4" />;
    case "comment_replied":
      return <MessageSquareReply className="h-4 w-4" />;
    case "user_followed":
      return <UserPlus className="h-4 w-4" />;
    default:
      return <Bell className="h-4 w-4" />;
  }
}

function iconTone(item: NotificationItem) {
  switch (item.type) {
    case "comment_liked":
      return "bg-rose-50 text-rose-600 ring-rose-100";
    case "post_commented":
      return "bg-amber-50 text-amber-600 ring-amber-100";
    case "post_voted":
      return item.direction && item.direction < 0 ? "bg-red-50 text-danger ring-red-100" : "bg-sky-50 text-primary ring-sky-100";
    case "comment_replied":
      return "bg-emerald-50 text-emerald-600 ring-emerald-100";
    case "user_followed":
      return "bg-violet-50 text-violet-600 ring-violet-100";
    default:
      return "bg-surface-soft text-primary ring-border";
  }
}

function normalizeLink(link: string) {
  if (!link) return "/notifications";
  return link.startsWith("/") ? link : `/${link}`;
}

function emitUnreadCount(count: number) {
  if (typeof window === "undefined") return;
  window.dispatchEvent(new CustomEvent<number>(notificationUnreadChangedEvent, { detail: Math.max(0, count) }));
}

export default function NotificationsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [status, setStatus] = useState<NotificationStatus>("all");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");
  const [markingAll, setMarkingAll] = useState(false);
  const [pendingID, setPendingID] = useState("");

  useEffect(() => {
    if (authLoading) return;
    if (!user) {
      setLoading(false);
      return;
    }

    let cancelled = false;
    const isFirstPage = page === 1;

    const fetchNotifications = async () => {
      setLoading(isFirstPage);
      setLoadingMore(!isFirstPage);
      setError("");
      try {
        const response = await apiClient.getNotifications({ page, size: PAGE_SIZE, status });
        if (cancelled) return;
        const nextList = response.data.list ?? [];
        setNotifications((current) => {
          if (isFirstPage) return nextList;
          const existing = new Set(current.map((item) => item.notification_id));
          return [...current, ...nextList.filter((item) => !existing.has(item.notification_id))];
        });
        setTotal(response.data.total || 0);
        const nextUnreadCount = response.data.unread_count || 0;
        setUnreadCount(nextUnreadCount);
        emitUnreadCount(nextUnreadCount);
      } catch (err) {
        if (cancelled) return;
        setError(getErrorMessage(err, "通知加载失败，请稍后重试。"));
        if (isFirstPage) setNotifications([]);
      } finally {
        if (!cancelled) {
          setLoading(false);
          setLoadingMore(false);
        }
      }
    };

    fetchNotifications();
    return () => {
      cancelled = true;
    };
  }, [authLoading, page, status, user]);

  const switchStatus = (nextStatus: NotificationStatus) => {
    if (nextStatus === status) return;
    setStatus(nextStatus);
    setPage(1);
    setNotifications([]);
  };

  const markAllRead = async () => {
    setMarkingAll(true);
    setError("");
    try {
      await apiClient.markAllNotificationsRead();
      setUnreadCount(0);
      emitUnreadCount(0);
      setNotifications((items) => items.map((item) => ({ ...item, is_read: true })));
      if (status === "unread") {
        setNotifications([]);
        setTotal(0);
      }
    } catch (err) {
      setError(getErrorMessage(err, "全部已读失败，请稍后重试。"));
    } finally {
      setMarkingAll(false);
    }
  };

  const openNotification = async (item: NotificationItem) => {
    setPendingID(item.notification_id);
    setError("");
    try {
      if (!item.is_read) {
        await apiClient.markNotificationRead(item.notification_id);
        setUnreadCount((value) => {
          const nextValue = Math.max(0, value - 1);
          emitUnreadCount(nextValue);
          return nextValue;
        });
        setNotifications((items) => items.map((current) => (current.notification_id === item.notification_id ? { ...current, is_read: true } : current)));
      }
      router.push(normalizeLink(item.link));
    } catch (err) {
      setError(getErrorMessage(err, "标记已读失败，请稍后重试。"));
    } finally {
      setPendingID("");
    }
  };

  const hasMore = notifications.length < total;

  if (!authLoading && !user) {
    return (
      <PageShell>
        <EmptyState
          title="登录后查看通知"
          description="有人回复、点赞、投票或关注你时，通知会集中出现在这里。"
          action={
            <Link className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark" href="/login">
              去登录
            </Link>
          }
        />
      </PageShell>
    );
  }

  return (
    <PageShell>
      <section className="overflow-hidden rounded-lg border border-border bg-surface shadow-sm">
        <div className="border-b border-border bg-[linear-gradient(135deg,rgba(69,123,157,0.12),rgba(255,255,255,0.92))] p-5">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div className="inline-flex items-center gap-2 rounded-full bg-white/75 px-3 py-1 text-xs font-semibold text-primary shadow-sm ring-1 ring-white/80">
                <BellRing className="h-3.5 w-3.5" />
                通知中心
              </div>
              <h1 className="mt-3 text-2xl font-bold text-foreground">与你相关的新动态</h1>
              <p className="mt-2 text-sm text-muted-strong">{unreadCount > 0 ? `你还有 ${unreadCount} 条未读通知。` : "所有通知都已处理。"}</p>
            </div>
            <Button variant="outline" disabled={markingAll || unreadCount === 0} onClick={markAllRead}>
              <CheckCheck className="h-4 w-4" />
              {markingAll ? "处理中" : "全部已读"}
            </Button>
          </div>

          <div className="mt-5 inline-flex rounded-lg bg-white/65 p-1 shadow-sm">
            {(["all", "unread"] as const).map((item) => (
              <button
                key={item}
                onClick={() => switchStatus(item)}
                className={cn(
                  "h-9 rounded-md px-4 text-sm font-medium transition",
                  status === item ? "bg-primary text-white shadow-sm" : "text-muted-strong hover:bg-white hover:text-foreground"
                )}
              >
                {item === "all" ? "全部" : "未读"}
              </button>
            ))}
          </div>
        </div>

        {error ? <div className="mx-5 mt-5 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">{error}</div> : null}

        <div className="p-5">
          {loading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, index) => (
                <div key={index} className="rounded-lg border border-border bg-surface p-4">
                  <Skeleton className="h-4 w-44" />
                  <Skeleton className="mt-3 h-4 w-2/3" />
                </div>
              ))}
            </div>
          ) : notifications.length > 0 ? (
            <div className="space-y-3">
              {notifications.map((item) => (
                <button
                  key={item.notification_id}
                  onClick={() => openNotification(item)}
                  disabled={pendingID === item.notification_id}
                  className={cn(
                    "group flex w-full items-start gap-3 rounded-lg border px-4 py-4 text-left transition",
                    item.is_read ? "border-border bg-white hover:border-border-strong hover:bg-surface-soft" : "border-[#457b9d]/25 bg-[#f3f8fb] hover:border-[#457b9d]/40 hover:bg-[#edf6fb]",
                    pendingID === item.notification_id && "opacity-60"
                  )}
                >
                  <span className={cn("mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg ring-1", iconTone(item))}>{notificationIcon(item)}</span>
                  <span className="min-w-0 flex-1">
                    <span className="flex flex-wrap items-center gap-2">
                      <span className="font-semibold text-foreground">{item.title}</span>
                      {!item.is_read ? <span className="rounded-full bg-danger px-2 py-0.5 text-[11px] font-bold text-white">未读</span> : null}
                    </span>
                    <span className="mt-1 block text-sm leading-6 text-muted-strong">{item.content}</span>
                    <span className="mt-2 block text-xs text-muted">{formatDateTime(item.create_time)}</span>
                  </span>
                </button>
              ))}
            </div>
          ) : (
            <EmptyState title={status === "unread" ? "没有未读通知" : "还没有通知"} description={status === "unread" ? "新的互动通知会在这里提醒你。" : "当有人回复、点赞、投票或关注你时，会出现在这里。"} />
          )}

          {!loading && hasMore ? (
            <div className="mt-5 flex justify-center">
              <Button variant="outline" disabled={loadingMore} onClick={() => setPage((value) => value + 1)}>
                {loadingMore ? "加载中" : "加载更多"}
              </Button>
            </div>
          ) : null}
        </div>
      </section>
    </PageShell>
  );
}
