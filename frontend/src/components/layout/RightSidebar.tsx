"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { ChevronRight, Flame, MessageSquareText, ShieldCheck, Users, Heart, FileText, X } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import Badge from "@/components/ui/Badge";
import Button from "@/components/ui/Button";
import EmptyState from "@/components/ui/EmptyState";
import PostCard from "@/components/ui/PostCard";
import { PostCardSkeleton } from "@/components/ui/Skeleton";
import {
  apiClient,
  type PostListItem,
  type PostsResponse,
  type UserFollowItem,
  type UserFollowListResponse,
  type UserProfile,
} from "@/lib/api";
import { formatCount } from "@/lib/format";
import { getErrorMessage } from "@/lib/utils";

type PanelType = "posts" | "followers" | "following" | null;

const POST_PAGE_SIZE = 5;

function normalizePosts(data: PostsResponse | PostListItem[] | undefined) {
  if (!data) return { list: [] as PostListItem[], total: undefined as number | undefined };
  if (Array.isArray(data)) return { list: data, total: undefined as number | undefined };
  return { list: data.list ?? [], total: data.total ?? 0 };
}

function StatButton({
  icon,
  label,
  value,
  onClick,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className="rounded-lg bg-surface-soft p-3 text-left transition hover:bg-white hover:shadow-sm"
      type="button"
    >
      <div className="flex items-center justify-between gap-3">
        <span className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-[#457b9d]/12 text-primary">{icon}</span>
        <span className="text-lg font-bold text-foreground">{value}</span>
      </div>
      <p className="mt-3 text-xs font-medium text-muted">{label}</p>
    </button>
  );
}

function UserListDialog({
  type,
  people,
  total,
  loading,
  currentUserId,
  canManageFollowing,
  pendingUserId,
  onFollow,
  onUnfollow,
  onClose,
}: {
  type: "followers" | "following";
  people: UserFollowItem[];
  total: number;
  loading: boolean;
  currentUserId?: string;
  canManageFollowing: boolean;
  pendingUserId: string;
  onFollow: (id: string) => void;
  onUnfollow: (id: string) => void;
  onClose: () => void;
}) {
  const title = type === "followers" ? "粉丝列表" : "关注列表";
  const description = type === "followers" ? "这些用户关注了你。" : "这些是你正在关注的用户。";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/35 px-4 backdrop-blur-sm" onClick={onClose}>
      <div className="w-full max-w-lg rounded-lg bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.18)]" onClick={(event) => event.stopPropagation()}>
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-semibold text-foreground">{title}</h3>
            <p className="mt-1 text-sm text-muted">{description} 共 {formatCount(total)} 人。</p>
          </div>
          <button
            onClick={onClose}
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-muted transition hover:bg-surface-soft hover:text-foreground"
            aria-label="关闭列表"
            type="button"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="mt-5 space-y-3">
          {loading ? (
            <div className="rounded-lg border border-border bg-surface-soft px-4 py-8 text-center text-sm text-muted">加载中...</div>
          ) : people.length > 0 ? (
            people.map((person) => (
              <div
                key={person.user_id}
                className="flex items-start justify-between gap-3 rounded-lg border border-border bg-surface px-4 py-3 transition hover:border-[#457b9d]/25 hover:bg-surface-soft"
              >
                <Link href={`/user/${person.user_id}`} onClick={onClose} className="flex min-w-0 flex-1 items-start gap-3">
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-[#457b9d]/12 text-sm font-semibold text-primary">
                    {person.user_name.slice(0, 1).toUpperCase()}
                  </div>
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="text-sm font-semibold text-foreground">{person.user_name}</p>
                      {person.is_mutual ? (
                        <span className="rounded-md bg-[#457b9d]/10 px-2 py-0.5 text-[11px] font-medium text-primary">
                          互相关注
                        </span>
                      ) : null}
                    </div>
                    <p className="mt-1 line-clamp-2 text-sm leading-6 text-muted-strong">
                      {person.signature || "这个人很懒，还没有留下签名。"}
                    </p>
                  </div>
                </Link>

                <div className="shrink-0 pt-1">
                  {currentUserId && person.user_id !== currentUserId && type === "followers" && !person.is_following ? (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={pendingUserId === person.user_id}
                      onClick={() => onFollow(person.user_id)}
                    >
                      {pendingUserId === person.user_id ? "处理中" : "关注Ta"}
                    </Button>
                  ) : null}
                  {canManageFollowing && currentUserId && person.user_id !== currentUserId && type === "following" ? (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={pendingUserId === person.user_id}
                      onClick={() => onUnfollow(person.user_id)}
                    >
                      {pendingUserId === person.user_id ? "处理中" : "取消关注"}
                    </Button>
                  ) : null}
                </div>
              </div>
            ))
          ) : (
            <div className="rounded-lg border border-border bg-surface-soft px-4 py-8 text-center text-sm text-muted">
              {type === "followers" ? "还没有粉丝。" : "还没有关注任何用户。"}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function PostsDialog({
  posts,
  loading,
  loadingMore,
  hasMore,
  total,
  error,
  loadMoreRef,
  onClose,
}: {
  posts: PostListItem[];
  loading: boolean;
  loadingMore: boolean;
  hasMore: boolean;
  total: number;
  error: string;
  loadMoreRef: React.RefObject<HTMLDivElement | null>;
  onClose: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/35 px-4 backdrop-blur-sm" onClick={onClose}>
      <div className="flex h-[min(80vh,760px)] w-full max-w-4xl flex-col rounded-lg bg-white shadow-[0_18px_50px_rgba(15,23,42,0.18)]" onClick={(event) => event.stopPropagation()}>
        <div className="flex items-start justify-between gap-3 border-b border-border px-5 py-4">
          <div>
            <h3 className="text-lg font-semibold text-foreground">发布记录</h3>
            <p className="mt-1 text-sm text-muted">共 {formatCount(total || posts.length)} 条帖子。</p>
          </div>
          <button
            onClick={onClose}
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-muted transition hover:bg-surface-soft hover:text-foreground"
            aria-label="关闭帖子列表"
            type="button"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-5 py-4">
          {error ? <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">{error}</div> : null}

          <div className="space-y-3">
            {loading
              ? Array.from({ length: 3 }).map((_, index) => <PostCardSkeleton key={index} />)
              : posts.map((post) => <PostCard key={post.post_id} post={post} />)}
          </div>

          {!loading && loadingMore ? (
            <div className="mt-3 space-y-3">
              {Array.from({ length: 2 }).map((_, index) => <PostCardSkeleton key={`dialog-loading-${index}`} />)}
            </div>
          ) : null}

          {!loading && posts.length === 0 ? (
            <div className="mt-2">
              <EmptyState title="还没有发布过帖子" description="等第一篇内容出现后，这里会按时间顺序继续展开。" />
            </div>
          ) : null}

          {!loading && posts.length > 0 ? (
            <div ref={loadMoreRef} className="flex justify-center py-4 text-sm text-muted-strong">
              {hasMore ? "继续向下滚动加载更多" : total > POST_PAGE_SIZE || posts.length > POST_PAGE_SIZE ? "已经到底了" : null}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}

export default function RightSidebar() {
  const { user, loading: authLoading } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [activePanel, setActivePanel] = useState<PanelType>(null);
  const [panelError, setPanelError] = useState("");
  const [followPeople, setFollowPeople] = useState<UserFollowItem[]>([]);
  const [followTotal, setFollowTotal] = useState(0);
  const [followListLoading, setFollowListLoading] = useState(false);
  const [pendingFollowListUserId, setPendingFollowListUserId] = useState("");
  const [postItems, setPostItems] = useState<PostListItem[]>([]);
  const [postPage, setPostPage] = useState(1);
  const [postTotal, setPostTotal] = useState(0);
  const [postsLoading, setPostsLoading] = useState(false);
  const [postsLoadingMore, setPostsLoadingMore] = useState(false);
  const [postsHasMore, setPostsHasMore] = useState(true);
  const postsLoadMoreRef = useRef<HTMLDivElement | null>(null);
  const postsRequestIdRef = useRef(0);
  const postsFetchingRef = useRef(false);

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

  useEffect(() => {
    if (activePanel !== "posts" || !user) return;

    const requestId = postsRequestIdRef.current + 1;
    postsRequestIdRef.current = requestId;
    const isFirstPage = postPage === 1;

    const fetchPosts = async () => {
      postsFetchingRef.current = true;
      setPostsLoading(isFirstPage);
      setPostsLoadingMore(!isFirstPage);
      setPanelError("");

      try {
        const response = await apiClient.getUserPosts(user.user_id, {
          page: postPage,
          size: POST_PAGE_SIZE,
          sort_by: "create_time",
          order: "desc",
        });
        if (requestId !== postsRequestIdRef.current) return;

        const normalized = normalizePosts(response.data);
        setPostItems((current) => {
          if (isFirstPage) return normalized.list;
          const existingIds = new Set(current.map((post) => post.post_id));
          const nextItems = normalized.list.filter((post) => !existingIds.has(post.post_id));
          return [...current, ...nextItems];
        });
        setPostTotal(normalized.total ?? 0);
        setPostsHasMore(
          normalized.list.length === POST_PAGE_SIZE &&
            (normalized.total === undefined || postPage * POST_PAGE_SIZE < normalized.total)
        );
      } catch (error) {
        if (requestId !== postsRequestIdRef.current) return;
        setPanelError(getErrorMessage(error, "帖子列表加载失败，请稍后再试。"));
        setPostsHasMore(false);
      } finally {
        if (requestId === postsRequestIdRef.current) {
          postsFetchingRef.current = false;
          setPostsLoading(false);
          setPostsLoadingMore(false);
        }
      }
    };

    fetchPosts();
  }, [activePanel, postPage, user]);

  useEffect(() => {
    const node = postsLoadMoreRef.current;
    if (!node || activePanel !== "posts" || postsLoading || postsLoadingMore || !postsHasMore || panelError) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry.isIntersecting || postsFetchingRef.current) return;
        postsFetchingRef.current = true;
        setPostPage((value) => value + 1);
      },
      { rootMargin: "240px 0px" }
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [activePanel, postsLoading, postsLoadingMore, postsHasMore, panelError]);

  const openPostsPanel = () => {
    setActivePanel("posts");
    setPanelError("");
    setPostItems([]);
    setPostTotal(0);
    setPostsHasMore(true);
    setPostPage(1);
  };

  const openFollowPanel = async (type: "followers" | "following") => {
    if (!user) return;

    setActivePanel(type);
    setPanelError("");
    setFollowPeople([]);
    setFollowTotal(0);
    setFollowListLoading(true);
    setPendingFollowListUserId("");

    try {
      const response =
        type === "followers"
          ? await apiClient.getUserFollowers(user.user_id, { page: 1, size: 50 })
          : await apiClient.getUserFollowing(user.user_id, { page: 1, size: 50 });
      setFollowPeople((response.data as UserFollowListResponse).list ?? []);
      setFollowTotal((response.data as UserFollowListResponse).total ?? 0);
    } catch (error) {
      setPanelError(getErrorMessage(error, "列表加载失败，请稍后再试。"));
      setActivePanel(null);
    } finally {
      setFollowListLoading(false);
    }
  };

  const handleFollowFromList = async (targetUserId: string) => {
    if (!user || pendingFollowListUserId) return;

    setPendingFollowListUserId(targetUserId);
    setPanelError("");

    try {
      await apiClient.followUser(targetUserId);
      setFollowPeople((items) =>
        items.map((item) =>
          item.user_id === targetUserId ? { ...item, is_following: true, is_mutual: item.is_followed_by } : item
        )
      );
      setProfile((current) => (current && activePanel === "following" ? { ...current, following_count: current.following_count + 1 } : current));
    } catch (error) {
      setPanelError(getErrorMessage(error, "关注操作失败，请稍后再试。"));
    } finally {
      setPendingFollowListUserId("");
    }
  };

  const handleUnfollowFromList = async (targetUserId: string) => {
    if (!user || pendingFollowListUserId) return;

    setPendingFollowListUserId(targetUserId);
    setPanelError("");

    try {
      await apiClient.unfollowUser(targetUserId);
      setFollowPeople((items) => items.filter((item) => item.user_id !== targetUserId));
      setFollowTotal((current) => Math.max(0, current - 1));
      setProfile((current) =>
        current ? { ...current, following_count: Math.max(0, current.following_count - 1) } : current
      );
    } catch (error) {
      setPanelError(getErrorMessage(error, "取消关注失败，请稍后再试。"));
    } finally {
      setPendingFollowListUserId("");
    }
  };

  const signature = profile?.signature || "写下一句属于你的介绍，让别人更快认识你。";

  return (
    <>
      <aside className="sticky top-5 hidden h-[calc(100vh-1.25rem)] overflow-y-auto xl:block scrollbar-thin">
        <div className="space-y-4">
          <section className="rounded-lg border border-border bg-surface p-5 shadow-sm">
            {user ? (
              <>
                <Link
                  href={`/user/${user.user_id}`}
                  className="group block rounded-lg border border-transparent p-1 transition hover:border-border hover:bg-surface-soft/70"
                  aria-label="进入个人中心"
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
                        <span className="group/heat relative inline-flex items-center gap-1 rounded-md bg-orange-50 px-2 py-0.5 text-[11px] font-medium text-orange-600">
                          <Flame className="h-3.5 w-3.5" />
                          <span>{formatCount(profile?.post_score ?? 0)}</span>
                          <span className="pointer-events-none absolute right-0 top-full z-10 mt-2 w-64 rounded-xl border border-border bg-white px-3 py-2 text-xs font-normal leading-5 text-muted-strong opacity-0 shadow-[0_16px_34px_rgba(15,23,42,0.12)] transition group-hover/heat:translate-y-1 group-hover/heat:opacity-100">
                            这是用户所有帖子的投票得分。
                          </span>
                        </span>
                      </div>
                      <p className="mt-1 line-clamp-2 text-xs leading-5 text-muted">{signature}</p>
                    </div>
                    <ChevronRight className="h-4 w-4 shrink-0 text-slate-400 transition group-hover:text-primary" />
                  </div>
                </Link>

                <div className="mt-4 grid grid-cols-3 gap-3">
                  <StatButton icon={<FileText className="h-4 w-4" />} label="发布" value={formatCount(profile?.post_count ?? 0)} onClick={openPostsPanel} />
                  <StatButton icon={<Users className="h-4 w-4" />} label="粉丝" value={formatCount(profile?.follower_count ?? 0)} onClick={() => void openFollowPanel("followers")} />
                  <StatButton icon={<Heart className="h-4 w-4" />} label="关注" value={formatCount(profile?.following_count ?? 0)} onClick={() => void openFollowPanel("following")} />
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
              <MessageSquareText className="h-5 w-5 text-accent" />
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

      {activePanel === "posts" ? (
        <PostsDialog
          posts={postItems}
          loading={postsLoading}
          loadingMore={postsLoadingMore}
          hasMore={postsHasMore}
          total={postTotal}
          error={panelError}
          loadMoreRef={postsLoadMoreRef}
          onClose={() => setActivePanel(null)}
        />
      ) : null}

      {activePanel === "followers" || activePanel === "following" ? (
        <UserListDialog
          type={activePanel}
          people={followPeople}
          total={followTotal}
          loading={followListLoading}
          currentUserId={user?.user_id}
          canManageFollowing={activePanel === "following"}
          pendingUserId={pendingFollowListUserId}
          onFollow={handleFollowFromList}
          onUnfollow={handleUnfollowFromList}
          onClose={() => setActivePanel(null)}
        />
      ) : null}
    </>
  );
}
