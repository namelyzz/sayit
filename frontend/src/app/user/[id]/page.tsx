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
  Users,
  X,
} from "lucide-react";
import Badge from "@/components/ui/Badge";
import Button from "@/components/ui/Button";
import EmptyState from "@/components/ui/EmptyState";
import PageShell from "@/components/ui/PageShell";
import PostCard from "@/components/ui/PostCard";
import { Textarea } from "@/components/ui/Field";
import { PostCardSkeleton } from "@/components/ui/Skeleton";
import { useAuth } from "@/context/AuthContext";
import { apiClient, type PostListItem, type PostsResponse } from "@/lib/api";
import { formatCount, formatDateTime, formatShortDate } from "@/lib/format";
import {
  DEFAULT_SIGNATURE,
  PREVIEW_REGISTER_DATE,
  previewComments,
  previewFollowers,
  previewFollowing,
  type PreviewPerson,
} from "@/lib/user-preview";
import { cn, getErrorMessage } from "@/lib/utils";

type ProfileListType = "followers" | "following" | null;

const SIGNATURE_STORAGE_PREFIX = "sayit:profile-signature:";

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

function PeopleDialog({
  type,
  people,
  onClose,
}: {
  type: Exclude<ProfileListType, null>;
  people: PreviewPerson[];
  onClose: () => void;
}) {
  const title = type === "followers" ? "粉丝列表" : "关注列表";
  const description =
    type === "followers" ? "这些用户关注了你。" : "这些是你正在关注的用户。";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/35 px-4 backdrop-blur-sm">
      <div className="w-full max-w-lg rounded-lg bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.18)]">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-semibold text-foreground">{title}</h3>
            <p className="mt-1 text-sm text-muted">{description}</p>
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
          {people.map((person) => (
            <div
              key={person.id}
              className="flex items-start gap-3 rounded-lg border border-border bg-surface px-4 py-3"
            >
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-[#457b9d]/12 text-sm font-semibold text-primary">
                {person.name.slice(0, 1)}
              </div>
              <div className="min-w-0">
                <p className="text-sm font-semibold text-foreground">{person.name}</p>
                <p className="mt-1 text-sm leading-6 text-muted-strong">{person.note}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export default function UserManagementPage() {
  const params = useParams();
  const profileId = params.id as string;
  const { user, loading: authLoading } = useAuth();

  const isSelf = user?.user_id === profileId;
  const [signature, setSignature] = useState(DEFAULT_SIGNATURE);
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [posts, setPosts] = useState<PostListItem[]>([]);
  const [error, setError] = useState("");
  const [activeList, setActiveList] = useState<ProfileListType>(null);

  useEffect(() => {
    if (typeof window === "undefined") return;

    const stored = window.localStorage.getItem(`${SIGNATURE_STORAGE_PREFIX}${profileId}`);
    setSignature(stored || DEFAULT_SIGNATURE);
  }, [profileId]);

  useEffect(() => {
    if (authLoading) return;
    if (!user) {
      setLoading(false);
      return;
    }

    const fetchPosts = async () => {
      setLoading(true);
      setError("");

      try {
        if (!isSelf) {
          setPosts([]);
          return;
        }

        const response = await apiClient.getPosts({
          page: 1,
          size: 50,
          sort_by: "create_time",
          order: "desc",
        });
        const allPosts = normalizePosts(response.data);
        setPosts(allPosts.filter((post) => post.user_name === user.user_name));
      } catch (err) {
        setError(getErrorMessage(err, "用户帖子加载失败，请稍后再试。"));
        setPosts([]);
      } finally {
        setLoading(false);
      }
    };

    fetchPosts();
  }, [authLoading, isSelf, user]);

  const postCount = posts.length;
  const latestPostTime = posts[0]?.create_time;
  const postScore = posts.reduce((sum, post) => sum + (post.like_count ?? 0), 0);
  const displayName = isSelf ? user?.user_name ?? "我的账号" : `用户 ${profileId.slice(0, 6)}`;
  const avatarText = displayName.slice(0, 1).toUpperCase();
  const people = activeList === "followers" ? previewFollowers : previewFollowing;

  const profileMeta = useMemo(
    () => [
      {
        label: "注册时间",
        value: formatShortDate(PREVIEW_REGISTER_DATE),
        icon: <CalendarDays className="h-4 w-4" />,
      },
      {
        label: "最后一个帖子时间",
        value: latestPostTime ? formatShortDate(latestPostTime) : "无",
        icon: <Clock3 className="h-4 w-4" />,
      },
      {
        label: "帖子总分",
        value: formatCount(postScore),
        icon: <Heart className="h-4 w-4" />,
      },
    ],
    [latestPostTime, postScore]
  );

  const handleSaveSignature = () => {
    if (typeof window === "undefined") return;
    window.localStorage.setItem(`${SIGNATURE_STORAGE_PREFIX}${profileId}`, signature.trim() || DEFAULT_SIGNATURE);
    setSaved(true);
    window.setTimeout(() => setSaved(false), 1800);
  };

  if (!authLoading && !user) {
    return (
      <PageShell className="!max-w-4xl">
        <EmptyState
          title="登录后查看个人中心"
          description="这里会展示你的粉丝、关注、帖子和最新评论。"
          action={
            <Link
              href="/login"
              className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
            >
              去登录
            </Link>
          }
        />
      </PageShell>
    );
  }

  return (
    <>
      <PageShell className="!max-w-none space-y-6">
        <section className="overflow-hidden rounded-lg border border-border bg-surface shadow-sm">
          <div className="bg-[linear-gradient(135deg,rgba(69,123,157,0.16),rgba(123,191,207,0.16))] px-6 py-6 md:px-7">
            <div className="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge tone="primary">个人中心</Badge>
                  <Badge tone="neutral">前端预览</Badge>
                </div>

                <div className="mt-5 flex items-start gap-4">
                  <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-[22px] bg-white text-2xl font-semibold text-primary shadow-sm">
                    {avatarText}
                  </div>
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-3">
                      <h1 className="text-2xl font-bold text-foreground">{displayName}</h1>
                      {isSelf ? <Badge tone="neutral">我的页面</Badge> : null}
                    </div>
                    <p className="mt-2 text-sm text-muted">ID {profileId}</p>
                    <p className="mt-3 max-w-2xl text-sm leading-7 text-muted-strong">
                      {signature.trim() || DEFAULT_SIGNATURE}
                    </p>
                  </div>
                </div>
              </div>

              <div className="grid w-full gap-3 sm:grid-cols-2 lg:max-w-[360px]">
                <button
                  onClick={() => setActiveList("followers")}
                  className="rounded-lg border border-white/70 bg-white/88 px-4 py-4 text-left shadow-sm transition hover:border-[#457b9d]/20 hover:shadow-md"
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-[#457b9d]/12 text-primary">
                      <Users className="h-5 w-5" />
                    </span>
                    <span className="text-xl font-semibold text-foreground">
                      {formatCount(previewFollowers.length)}
                    </span>
                  </div>
                  <p className="mt-4 text-sm font-semibold text-foreground">粉丝</p>
                  <p className="mt-1 text-xs leading-5 text-muted">看看是谁在关注你</p>
                </button>

                <button
                  onClick={() => setActiveList("following")}
                  className="rounded-lg border border-white/70 bg-white/88 px-4 py-4 text-left shadow-sm transition hover:border-[#457b9d]/20 hover:shadow-md"
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-[#457b9d]/12 text-primary">
                      <Heart className="h-5 w-5" />
                    </span>
                    <span className="text-xl font-semibold text-foreground">
                      {formatCount(previewFollowing.length)}
                    </span>
                  </div>
                  <p className="mt-4 text-sm font-semibold text-foreground">关注</p>
                  <p className="mt-1 text-xs leading-5 text-muted">看看你正在关注谁</p>
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
              extra={saved ? <span className="text-xs font-medium text-primary">已保存到本地预览</span> : null}
            >
              {isSelf ? (
                <div className="space-y-4">
                  <Textarea
                    rows={4}
                    maxLength={120}
                    value={signature}
                    onChange={(event) => setSignature(event.target.value)}
                    placeholder="写一句让别人认识你的话。"
                  />
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <p className="text-sm text-muted">当前为前端预览版，签名会先保存在本地浏览器。</p>
                    <Button onClick={handleSaveSignature}>
                      <PencilLine className="h-4 w-4" />
                      保存签名
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="rounded-lg border border-border bg-surface-soft px-4 py-4 text-sm leading-7 text-muted-strong">
                  后续接入后端后，这里会展示该用户公开的个性签名。
                </div>
              )}
            </SectionCard>
          </div>

          <SectionCard title="关系预览">
            <div className="space-y-3">
              <div className="rounded-lg border border-border bg-surface-soft px-4 py-4">
                <p className="text-sm font-semibold text-foreground">粉丝</p>
                <p className="mt-1 text-sm leading-6 text-muted-strong">
                  点击顶部统计卡，可查看关注你的用户列表。
                </p>
              </div>
              <div className="rounded-lg border border-border bg-surface-soft px-4 py-4">
                <p className="text-sm font-semibold text-foreground">关注</p>
                <p className="mt-1 text-sm leading-6 text-muted-strong">
                  点击顶部统计卡，可查看你关注的用户列表。
                </p>
              </div>
            </div>
          </SectionCard>
        </div>

        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
          <SectionCard
            title="最新帖子"
            extra={<span className="text-sm font-medium text-muted-strong">共 {formatCount(postCount)} 条</span>}
          >
            <div className="space-y-3">
              {loading ? (
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
            extra={<span className="text-sm font-medium text-muted-strong">共 {formatCount(previewComments.length)} 条</span>}
          >
            <div className="space-y-3">
              {previewComments.map((comment) => (
                <article key={comment.id} className="rounded-lg border border-border bg-surface-soft px-4 py-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="text-sm font-semibold text-foreground">{comment.postTitle}</p>
                      <p className="mt-2 text-sm leading-6 text-muted-strong">{comment.excerpt}</p>
                    </div>
                    <span className="inline-flex shrink-0 items-center gap-1 rounded-md bg-white px-2 py-1 text-xs font-medium text-muted">
                      <MessageSquareText className="h-3.5 w-3.5" />
                      预览
                    </span>
                  </div>
                  <p className="mt-3 text-xs text-muted">{formatDateTime(comment.createdAt)}</p>
                </article>
              ))}

              <div className="rounded-lg border border-dashed border-border-strong bg-surface px-4 py-4 text-sm leading-6 text-muted">
                评论系统接入后，这里会按时间展示你的最新评论，并同步评论总数。
              </div>
            </div>
          </SectionCard>
        </div>
      </PageShell>

      {activeList ? <PeopleDialog type={activeList} people={people} onClose={() => setActiveList(null)} /> : null}
    </>
  );
}
