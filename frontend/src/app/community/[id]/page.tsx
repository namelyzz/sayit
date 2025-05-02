"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Calendar, Heart, PenLine, Users } from "lucide-react";
import PostCard from "@/components/ui/PostCard";
import Button from "@/components/ui/Button";
import EmptyState from "@/components/ui/EmptyState";
import PageShell from "@/components/ui/PageShell";
import { PostCardSkeleton, Skeleton } from "@/components/ui/Skeleton";
import { apiClient, type CommunityDetail, type PostListItem, type PostsResponse } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { formatDate } from "@/lib/format";
import { cn, getErrorMessage } from "@/lib/utils";

function normalizePosts(data: PostsResponse | PostListItem[] | undefined) {
  if (!data) return [] as PostListItem[];
  return Array.isArray(data) ? data : data.list ?? [];
}

export default function CommunityPage() {
  const params = useParams();
  const router = useRouter();
  const communityId = params.id as string;
  const { user, loading: authLoading } = useAuth();

  const [community, setCommunity] = useState<CommunityDetail | null>(null);
  const [posts, setPosts] = useState<PostListItem[]>([]);
  const [isFollowed, setIsFollowed] = useState(false);
  const [loading, setLoading] = useState(true);
  const [followLoading, setFollowLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (authLoading) return;

    const fetchData = async () => {
      setLoading(true);
      setError("");
      try {
        const [communityResponse, postsResponse] = await Promise.all([
          apiClient.getCommunityDetail(communityId),
          apiClient.getPosts({ community_id: communityId, size: 20 }),
        ]);

        setCommunity(communityResponse.data);
        setPosts(normalizePosts(postsResponse.data));

        if (user) {
          try {
            const followResponse = await apiClient.isFollowedCommunity(communityId);
            setIsFollowed(followResponse.data?.is_followed ?? false);
          } catch {
            setIsFollowed(false);
          }
        }
      } catch (err) {
        setError(getErrorMessage(err, "社区数据加载失败，请确认后端服务已启动。"));
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [communityId, authLoading, user]);

  const handleFollowToggle = async () => {
    if (!user) {
      router.push("/login");
      return;
    }

    setFollowLoading(true);
    try {
      if (isFollowed) {
        await apiClient.unfollowCommunity(communityId);
        setIsFollowed(false);
      } else {
        await apiClient.followCommunity(communityId);
        setIsFollowed(true);
      }
      window.dispatchEvent(new CustomEvent("follow-refresh"));
    } catch (err) {
      const message = getErrorMessage(err, "操作失败，请重试。");
      if (message.includes("token") || message.includes("登录")) {
        router.push("/login");
      } else {
        setError(message);
      }
    } finally {
      setFollowLoading(false);
    }
  };

  if (loading) {
    return (
      <PageShell>
        <section className="rounded-lg border border-border bg-surface p-5 shadow-sm">
          <Skeleton className="h-7 w-48" />
          <Skeleton className="mt-4 h-4 w-full" />
          <Skeleton className="mt-2 h-4 w-2/3" />
        </section>
        <PostCardSkeleton />
        <PostCardSkeleton />
      </PageShell>
    );
  }

  if (error && !community) {
    return (
      <PageShell>
        <EmptyState
          title="社区加载失败"
          description={error}
          action={
            <Link
              href="/"
              className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
            >
              返回首页
            </Link>
          }
        />
      </PageShell>
    );
  }

  if (!community) {
    return (
      <PageShell>
        <EmptyState title="社区不存在" description="这个社区可能已经被移除，或者链接不正确。" />
      </PageShell>
    );
  }

  return (
    <PageShell>
      <section className="overflow-hidden rounded-lg border border-border bg-surface shadow-sm">
        <div className="h-24 bg-gradient-to-r from-teal-700 via-teal-600 to-amber-500" />
        <div className="p-5">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0">
              <div className="-mt-12 mb-4 flex h-16 w-16 items-center justify-center rounded-lg border-4 border-surface bg-primary text-2xl font-bold text-white shadow-sm">
                {community.name.slice(0, 1).toUpperCase()}
              </div>
              <h1 className="text-2xl font-bold text-foreground">{community.name}</h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-strong">
                {community.introduction || "这个社区还没有介绍。"}
              </p>
              <div className="mt-4 flex flex-wrap gap-3 text-sm text-muted">
                <span className="inline-flex items-center gap-1.5">
                  <Users className="h-4 w-4" />
                  社区成员
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <Calendar className="h-4 w-4" />
                  创建于 {formatDate(community.create_time)}
                </span>
              </div>
            </div>

            <div className="flex shrink-0 gap-2">
              {user ? (
                <Button
                  variant={isFollowed ? "primary" : "outline"}
                  onClick={handleFollowToggle}
                  disabled={followLoading}
                  className={cn(isFollowed && "bg-primary")}
                >
                  <Heart className={cn("h-4 w-4", isFollowed && "fill-current")} />
                  {isFollowed ? "已关注" : "关注"}
                </Button>
              ) : (
                <Link
                  href="/login"
                  className="inline-flex h-10 items-center justify-center rounded-lg border border-border bg-surface px-4 text-sm font-medium text-muted-strong transition hover:bg-surface-soft"
                >
                  登录后关注
                </Link>
              )}
              <Link
                href="/submit"
                className="inline-flex h-10 items-center justify-center gap-2 rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
              >
                <PenLine className="h-4 w-4" />
                发帖
              </Link>
            </div>
          </div>
        </div>
      </section>

      {error ? <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">{error}</div> : null}

      <div className="space-y-3">
        {posts.length > 0 ? posts.map((post) => <PostCard key={post.post_id} post={post} />) : null}
      </div>

      {posts.length === 0 ? (
        <EmptyState
          title="这个社区还没有帖子"
          description="发起第一个话题，让这里热闹起来。"
          action={
            <Link
              href="/submit"
              className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
            >
              发布帖子
            </Link>
          }
        />
      ) : null}
    </PageShell>
  );
}
