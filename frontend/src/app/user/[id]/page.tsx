"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import {
  CalendarDays,
  Clock3,
  Heart,
  MessageSquareText,
  PencilLine,
  UserCheck,
  UserPlus,
  Users,
  X,
} from "lucide-react";
import Badge from "@/components/ui/Badge";
import Button from "@/components/ui/Button";
import EmptyState from "@/components/ui/EmptyState";
import { Textarea } from "@/components/ui/Field";
import PageShell from "@/components/ui/PageShell";
import PostCard from "@/components/ui/PostCard";
import { PostCardSkeleton } from "@/components/ui/Skeleton";
import { useAuth } from "@/context/AuthContext";
import { apiClient, type PostListItem, type PostsResponse, type UserCommentItem, type UserFollowItem, type UserProfile } from "@/lib/api";
import { formatCount, formatDateTime, formatShortDate } from "@/lib/format";
import { cn, getErrorMessage } from "@/lib/utils";

type ProfileListType = "followers" | "following" | null;

const profileBadges = [
  {
    icon: "💗",
    title: "欢迎",
    description: "收到第一个赞，说明你的表达已经被人看见。",
    tone: "from-rose-100 via-white to-rose-50 text-rose-500",
  },
  {
    icon: "🪐",
    title: "初次远行",
    description: "加入第二个社区，开始在不同角落留下足迹。",
    tone: "from-sky-100 via-white to-cyan-50 text-sky-600",
  },
  {
    icon: "🔥",
    title: "火花",
    description: "一篇帖子引起热烈讨论，点亮了首页的气氛。",
    tone: "from-amber-100 via-white to-orange-50 text-orange-500",
  },
  {
    icon: "🌱",
    title: "发芽",
    description: "写下第一篇帖子，让个人主页不再空白。",
    tone: "from-emerald-100 via-white to-lime-50 text-emerald-600",
  },
  {
    icon: "🎧",
    title: "认真倾听",
    description: "留下第一条评论，正式加入别人的对话。",
    tone: "from-violet-100 via-white to-fuchsia-50 text-violet-600",
  },
  {
    icon: "🏆",
    title: "常驻居民",
    description: "连续活跃了一段时间，已经成了社区里的熟面孔。",
    tone: "from-yellow-100 via-white to-amber-50 text-yellow-600",
  },
] as const;

function normalizePosts(data: PostsResponse | PostListItem[] | undefined) {
  if (!data) return [] as PostListItem[];
  return Array.isArray(data) ? data : data.list ?? [];
}

function SectionCard({
  title,
  extra,
  children,
  className,
}: {
  title: string;
  extra?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("rounded-lg border border-border bg-surface p-5 shadow-sm", className)}>
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-base font-semibold text-foreground">{title}</h2>
        {extra ? <div className="shrink-0">{extra}</div> : null}
      </div>
      <div className="mt-5">{children}</div>
    </section>
  );
}

function LoginLockedCard({ title, description }: { title: string; description: string }) {
  return (
    <div className="rounded-lg border border-dashed border-border-strong bg-surface-soft px-4 py-5 text-sm leading-6 text-muted-strong">
      <p className="font-semibold text-foreground">{title}</p>
      <p className="mt-1">{description}</p>
      <Link
        href="/login"
        className="mt-4 inline-flex h-9 items-center justify-center rounded-lg bg-primary px-3 text-sm font-medium text-white transition hover:bg-primary-dark"
      >
        登录后查看更多
      </Link>
    </div>
  );
}

function InfoItem({
  label,
  value,
  icon,
}: {
  label: string;
  value: string;
  icon: React.ReactNode;
}) {
  return (
    <div className="rounded-lg border border-border bg-surface-soft px-4 py-4">
      <div className="flex items-center gap-2 text-sm text-muted">
        <span className="text-primary">{icon}</span>
        <span>{label}</span>
      </div>
      <p className="mt-3 text-lg font-semibold text-foreground">{value}</p>
    </div>
  );
}

function BadgeCabinet() {
  return (
    <div className="rounded-[24px] border border-[#d6c2a3] bg-[linear-gradient(180deg,#f4ead8_0%,#ebdcc5_100%)] p-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.55),0_10px_24px_rgba(123,87,42,0.08)]">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        {profileBadges.map((badge) => (
          <div key={badge.title} className="group relative aspect-square">
            <div className="absolute inset-0 rounded-[18px] border border-[#b9966c] bg-[linear-gradient(135deg,rgba(255,255,255,0.5),rgba(148,111,71,0.08))] shadow-[inset_0_1px_0_rgba(255,255,255,0.65),inset_0_-8px_14px_rgba(134,97,57,0.08)] transition duration-200 group-hover:border-[#8e6a42] group-hover:shadow-[inset_0_1px_0_rgba(255,255,255,0.8),0_14px_24px_rgba(94,64,31,0.18)] group-hover:brightness-105" />
            <div className="absolute inset-[10px] rounded-[14px] border border-white/70 bg-[rgba(255,248,238,0.92)] transition group-hover:bg-white">
              <div
                className={cn(
                  "flex h-full items-center justify-center rounded-[13px] bg-gradient-to-br text-3xl shadow-[inset_0_1px_0_rgba(255,255,255,0.7)] transition duration-200 group-hover:scale-[1.04]",
                  badge.tone
                )}
              >
                <span aria-hidden="true">{badge.icon}</span>
              </div>
            </div>

            <div className="pointer-events-none absolute left-1/2 top-full z-10 mt-3 w-52 -translate-x-1/2 rounded-2xl border border-border bg-white px-4 py-3 text-left opacity-0 shadow-[0_18px_40px_rgba(15,23,42,0.14)] transition duration-200 group-hover:translate-y-1 group-hover:opacity-100">
              <div className="absolute left-1/2 top-0 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rotate-45 border-l border-t border-border bg-white" />
              <p className="text-sm font-semibold text-foreground">{badge.title}</p>
              <p className="mt-1 text-sm leading-6 text-muted-strong">{badge.description}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function PeopleDialog({
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
  type: Exclude<ProfileListType, null>;
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
  const description =
    type === "followers" ? "这些用户关注了这个主页。" : "这个主页正在关注这些用户。";

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

export default function UserManagementPage() {
  const params = useParams();
  const profileId = params.id as string;
  const { user, loading: authLoading } = useAuth();

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [posts, setPosts] = useState<PostListItem[]>([]);
  const [comments, setComments] = useState<UserCommentItem[]>([]);
  const [commentTotal, setCommentTotal] = useState(0);
  const [error, setError] = useState("");
  const [signatureDraft, setSignatureDraft] = useState("");
  const [savingSignature, setSavingSignature] = useState(false);
  const [signatureSaved, setSignatureSaved] = useState(false);
  const [isFollowingUser, setIsFollowingUser] = useState(false);
  const [savingFollow, setSavingFollow] = useState(false);
  const [activeList, setActiveList] = useState<ProfileListType>(null);
  const [followPeople, setFollowPeople] = useState<UserFollowItem[]>([]);
  const [followTotal, setFollowTotal] = useState(0);
  const [followListLoading, setFollowListLoading] = useState(false);
  const [pendingFollowListUserId, setPendingFollowListUserId] = useState("");

  useEffect(() => {
    if (authLoading) return;

    const fetchProfile = async () => {
      setLoading(true);
      setError("");

      try {
        const profileResponse = await apiClient.getUserProfile(profileId);
        const nextProfile = profileResponse.data;
        setProfile(nextProfile);
        setSignatureDraft(nextProfile.signature);
        setIsFollowingUser(nextProfile.is_following);

        const postsResponse = await apiClient.getUserPosts(profileId, {
          page: 1,
          size: 50,
          sort_by: "create_time",
          order: "desc",
        });
        setPosts(normalizePosts(postsResponse.data));

        const commentsResponse = await apiClient.getUserComments(profileId);
        setComments(commentsResponse.data.list);
        setCommentTotal(commentsResponse.data.total);
      } catch (err) {
        setError(getErrorMessage(err, "用户资料加载失败，请稍后再试。"));
        setProfile(null);
        setPosts([]);
        setComments([]);
        setCommentTotal(0);
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, [authLoading, profileId]);

  const isSelf = Boolean(profile?.is_self || user?.user_id === profileId);
  const postCount = profile?.post_count ?? posts.length;
  const latestPostTime = posts[0]?.create_time;
  const postHeat = profile?.post_score ?? posts.reduce((sum, post) => sum + (post.like_count ?? 0), 0);
  const commentHeat = profile?.comment_score ?? 0;
  const totalHeat = postHeat + commentHeat;
  const displayName = profile?.user_name ?? (isSelf ? user?.user_name ?? "我的账号" : `用户 ${profileId.slice(0, 6)}`);
  const signature = profile?.signature ?? "这个人很懒，还没有留下签名。";
  const followerCount = profile?.follower_count ?? 0;
  const followingCount = profile?.following_count ?? 0;
  const avatarText = displayName.slice(0, 1).toUpperCase();
  const canViewPrivateBlocks = Boolean(user);

  const profileMeta = useMemo(
    () => [
      {
        label: "注册时间",
        value: profile?.create_time ? formatShortDate(profile.create_time) : "加载中",
        icon: <CalendarDays className="h-4 w-4" />,
      },
      {
        label: "最后一个帖子时间",
        value: latestPostTime ? formatShortDate(latestPostTime) : "无",
        icon: <Clock3 className="h-4 w-4" />,
      },
      {
        label: "热度值",
        value: formatCount(totalHeat),
        icon: <Heart className="h-4 w-4" />,
      },
    ],
    [latestPostTime, totalHeat, profile?.create_time]
  );

  const handleSaveSignature = async () => {
    setSavingSignature(true);
    setSignatureSaved(false);
    setError("");

    try {
      const response = await apiClient.updateMe(signatureDraft);
      setProfile(response.data);
      setSignatureDraft(response.data.signature);
      setSignatureSaved(true);
      window.setTimeout(() => setSignatureSaved(false), 1800);
    } catch (err) {
      setError(getErrorMessage(err, "签名保存失败，请稍后再试。"));
    } finally {
      setSavingSignature(false);
    }
  };

  const handleToggleFollow = async () => {
    if (!profile || savingFollow) return;

    setSavingFollow(true);
    setError("");

    try {
      const response = isFollowingUser
        ? await apiClient.unfollowUser(profileId)
        : await apiClient.followUser(profileId);
      const nextFollowing = response.data.is_following;
      setIsFollowingUser(nextFollowing);
      setProfile({
        ...profile,
        is_following: nextFollowing,
        follower_count: Math.max(0, profile.follower_count + (nextFollowing ? 1 : -1)),
      });
    } catch (err) {
      setError(getErrorMessage(err, "关注操作失败，请稍后再试。"));
    } finally {
      setSavingFollow(false);
    }
  };

  const handleOpenFollowList = async (type: Exclude<ProfileListType, null>) => {
    if (!canViewPrivateBlocks) return;

    setActiveList(type);
    setFollowPeople([]);
    setFollowTotal(0);
    setFollowListLoading(true);
    setPendingFollowListUserId("");
    setError("");

    try {
      const response =
        type === "followers"
          ? await apiClient.getUserFollowers(profileId, { page: 1, size: 50 })
          : await apiClient.getUserFollowing(profileId, { page: 1, size: 50 });
      setFollowPeople(response.data.list ?? []);
      setFollowTotal(response.data.total ?? 0);
    } catch (err) {
      setError(getErrorMessage(err, "关注列表加载失败，请稍后再试。"));
      setActiveList(null);
    } finally {
      setFollowListLoading(false);
    }
  };

  const handleFollowFromList = async (targetUserId: string) => {
    if (pendingFollowListUserId) return;

    setPendingFollowListUserId(targetUserId);
    setError("");

    try {
      await apiClient.followUser(targetUserId);
      setFollowPeople((items) =>
        items.map((item) =>
          item.user_id === targetUserId
            ? { ...item, is_following: true, is_mutual: item.is_followed_by }
            : item
        )
      );
      if (user?.user_id === profileId) {
        setProfile((current) =>
          current ? { ...current, following_count: current.following_count + 1 } : current
        );
      }
    } catch (err) {
      setError(getErrorMessage(err, "关注操作失败，请稍后再试。"));
    } finally {
      setPendingFollowListUserId("");
    }
  };

  const handleUnfollowFromList = async (targetUserId: string) => {
    if (pendingFollowListUserId) return;

    setPendingFollowListUserId(targetUserId);
    setError("");

    try {
      await apiClient.unfollowUser(targetUserId);
      setFollowPeople((items) => items.filter((item) => item.user_id !== targetUserId));
      setFollowTotal((current) => Math.max(0, current - 1));
      if (user?.user_id === profileId) {
        setProfile((current) =>
          current ? { ...current, following_count: Math.max(0, current.following_count - 1) } : current
        );
      }
    } catch (err) {
      setError(getErrorMessage(err, "取消关注失败，请稍后再试。"));
    } finally {
      setPendingFollowListUserId("");
    }
  };

  const followAction = isSelf ? null : user ? (
    <Button
      variant={isFollowingUser ? "outline" : "primary"}
      onClick={handleToggleFollow}
      disabled={savingFollow || !profile}
      size="sm"
      className="h-8 px-3"
    >
      {isFollowingUser ? <UserCheck className="h-4 w-4" /> : <UserPlus className="h-4 w-4" />}
      {savingFollow ? "处理中" : isFollowingUser ? "已关注" : "关注"}
    </Button>
  ) : (
    <Link
      href="/login"
      className="inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border border-border bg-white px-3 text-xs font-medium text-muted-strong shadow-sm transition hover:border-[#457b9d]/30 hover:bg-[#457b9d]/8 hover:text-primary"
    >
      <UserPlus className="h-4 w-4" />
      登录关注
    </Link>
  );

  return (
    <>
      <PageShell className="!max-w-none space-y-6">
        <section className="overflow-hidden rounded-lg border border-border bg-surface shadow-sm">
          <div className="bg-[linear-gradient(135deg,rgba(69,123,157,0.16),rgba(123,191,207,0.16))] px-6 py-6 md:px-7">
            <div className="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge tone="primary">个人中心</Badge>
                </div>

                <div className="mt-5 flex items-start gap-4">
                  <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-[22px] bg-white text-2xl font-semibold text-primary shadow-sm">
                    {avatarText}
                  </div>
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-3">
                      <h1 className="text-2xl font-bold text-foreground">{displayName}</h1>
                      {isSelf ? <Badge tone="neutral">我的页面</Badge> : null}
                      {followAction}
                    </div>
                    <p className="mt-2 text-sm text-muted">ID {profileId}</p>
                    <p className="mt-3 max-w-2xl text-sm leading-7 text-muted-strong">
                      {signature}
                    </p>
                  </div>
                </div>
              </div>

              <div className="grid w-full gap-3 sm:grid-cols-2 lg:max-w-[360px]">
                <button
                  onClick={() => handleOpenFollowList("followers")}
                  className="rounded-lg border border-white/70 bg-white/88 px-4 py-4 text-left shadow-sm transition hover:border-[#457b9d]/20 hover:shadow-md"
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-[#457b9d]/12 text-primary">
                      <Users className="h-5 w-5" />
                    </span>
                    <span className="text-xl font-semibold text-foreground">
                      {formatCount(followerCount)}
                    </span>
                  </div>
                  <p className="mt-4 text-sm font-semibold text-foreground">粉丝</p>
                  <p className="mt-1 text-xs leading-5 text-muted">
                    {canViewPrivateBlocks ? "看看是谁在关注你" : "登录后查看关系详情"}
                  </p>
                </button>

                <button
                  onClick={() => handleOpenFollowList("following")}
                  className="rounded-lg border border-white/70 bg-white/88 px-4 py-4 text-left shadow-sm transition hover:border-[#457b9d]/20 hover:shadow-md"
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-[#457b9d]/12 text-primary">
                      <Heart className="h-5 w-5" />
                    </span>
                    <span className="text-xl font-semibold text-foreground">
                      {formatCount(followingCount)}
                    </span>
                  </div>
                  <p className="mt-4 text-sm font-semibold text-foreground">关注</p>
                  <p className="mt-1 text-xs leading-5 text-muted">
                    {canViewPrivateBlocks ? "看看你正在关注谁" : "登录后查看关系详情"}
                  </p>
                </button>
              </div>
            </div>
          </div>
        </section>

        {error ? (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">
            {error}
          </div>
        ) : null}

        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
          <div className="space-y-6">
            <SectionCard title="个人信息">
              <div className="grid gap-3 md:grid-cols-3">
                {profileMeta.map((item) => (
                  <InfoItem key={item.label} label={item.label} value={item.value} icon={item.icon} />
                ))}
              </div>
            </SectionCard>

            <SectionCard
              title="个性签名"
              extra={
                signatureSaved ? (
                  <span className="text-xs font-medium text-primary">已保存</span>
                ) : (
                  <span className="text-xs font-medium text-muted">编辑个人介绍</span>
                )
              }
            >
              {isSelf ? (
                <div className="space-y-4">
                  <Textarea
                    rows={4}
                    maxLength={120}
                    value={signatureDraft}
                    onChange={(event) => setSignatureDraft(event.target.value)}
                    placeholder="写一句让别人认识你的话。"
                  />
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <p className="text-sm text-muted">最多 120 个字符，写一句让别人记住你的话。</p>
                    <Button onClick={handleSaveSignature} disabled={savingSignature || !profile}>
                      <PencilLine className="h-4 w-4" />
                      {savingSignature ? "保存中" : "保存签名"}
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="rounded-lg border border-border bg-surface-soft px-4 py-4 text-sm leading-7 text-muted-strong">
                  {signature}
                </div>
              )}
            </SectionCard>
          </div>

          <SectionCard title="徽章柜">
            <div className="space-y-4">
              <p className="text-sm leading-6 text-muted-strong">
                用几枚小徽章收纳你的社区旅程。把鼠标移到格子上，可以看看每枚徽章代表什么。
              </p>
              <BadgeCabinet />
            </div>
          </SectionCard>
        </div>

        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
          <SectionCard
            title="最新帖子"
            extra={<span className="text-sm font-medium text-muted-strong">共 {formatCount(postCount)} 条</span>}
          >
            <div className="space-y-3">
              {!user ? (
                <LoginLockedCard title="帖子列表已锁定" description="登录后可以查看该用户发布的最新帖子并参与互动。" />
              ) : loading ? (
                <>
                  <PostCardSkeleton />
                  <PostCardSkeleton />
                </>
              ) : posts.length > 0 ? (
                posts.map((post) => <PostCard key={post.post_id} post={post} />)
              ) : (
                <EmptyState
                  title={isSelf ? "你还没有发布帖子" : "这位用户暂时还没有公开帖子"}
                  description={
                    isSelf
                      ? "第一篇帖子发出去之后，这里会自动展示最新内容。"
                      : "等后端资料接口接入后，这里会展示对方的最新帖子。"
                  }
                  action={
                    isSelf ? (
                      <Link
                        href="/submit"
                        className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
                      >
                        去写帖子
                      </Link>
                    ) : null
                  }
                />
              )}
            </div>
          </SectionCard>

          <SectionCard
            title="最新评论"
            extra={<span className="text-sm font-medium text-muted-strong">共 {formatCount(commentTotal)} 条</span>}
          >
            {canViewPrivateBlocks ? (
              comments.length > 0 ? (
                <div className="space-y-3">
                  {comments.map((comment) => (
                    <Link
                      key={comment.comment_id}
                      href={`/post/${comment.post_id}`}
                      className="block rounded-lg border border-border bg-surface-soft px-4 py-4 transition hover:border-primary/30 hover:shadow-sm"
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <p className="text-sm font-semibold text-foreground">{comment.post_title || "无标题"}</p>
                          <p className="mt-2 text-sm leading-6 text-muted-strong line-clamp-2">{comment.content}</p>
                        </div>
                        <span className="inline-flex shrink-0 items-center gap-1 rounded-md bg-white px-2 py-1 text-xs font-medium text-muted">
                          <MessageSquareText className="h-3.5 w-3.5" />
                          评论
                        </span>
                      </div>
                      <div className="mt-3 flex items-center gap-3 text-xs text-muted">
                        <span>{formatDateTime(comment.create_time)}</span>
                        {comment.like_count > 0 && (
                          <span className="flex items-center gap-1">
                            <Heart className="h-3 w-3" />
                            {formatCount(comment.like_count)}
                          </span>
                        )}
                      </div>
                    </Link>
                  ))}
                </div>
              ) : (
                <EmptyState
                  title="暂无评论"
                  description={isSelf ? "去帖子下面发表你的看法吧。" : "该用户还没有发表过评论。"}
                />
              )
            ) : (
              <LoginLockedCard title="评论信息已锁定" description="登录后可以查看更多评论活动和个人互动信息。" />
            )}
          </SectionCard>
        </div>
      </PageShell>

      {activeList ? (
        <PeopleDialog
          type={activeList}
          people={followPeople}
          total={followTotal}
          loading={followListLoading}
          currentUserId={user?.user_id}
          canManageFollowing={Boolean(user?.user_id === profileId)}
          pendingUserId={pendingFollowListUserId}
          onFollow={handleFollowFromList}
          onUnfollow={handleUnfollowFromList}
          onClose={() => setActiveList(null)}
        />
      ) : null}
    </>
  );
}
