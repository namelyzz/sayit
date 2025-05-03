"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { Clock3, Flame, PenLine } from "lucide-react";
import PostCard from "@/components/ui/PostCard";
import Button from "@/components/ui/Button";
import EmptyState from "@/components/ui/EmptyState";
import PageShell from "@/components/ui/PageShell";
import { PostCardSkeleton } from "@/components/ui/Skeleton";
import { apiClient, type PostListItem, type PostsResponse } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { cn, getErrorMessage } from "@/lib/utils";

type SortBy = "create_time" | "score";

function normalizePosts(data: PostsResponse | PostListItem[] | undefined) {
  if (!data) return { list: [] as PostListItem[], total: 0 };
  if (Array.isArray(data)) return { list: data, total: data.length };
  return { list: data.list ?? [], total: data.total ?? 0 };
}

export default function Home() {
  const { user, loading: authLoading } = useAuth();
  const [posts, setPosts] = useState<PostListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [sortBy, setSortBy] = useState<SortBy>("create_time");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState("");
  const [needLogin, setNeedLogin] = useState(false);

  useEffect(() => {
    if (authLoading) return;

    const fetchPosts = async () => {
      setLoading(true);
      setError("");
      try {
        const response = await apiClient.getPosts({
          page,
          size: 10,
          sort_by: sortBy,
          order: "desc",
        });
        const normalized = normalizePosts(response.data);
        setPosts(normalized.list);
        setTotal(normalized.total);
        setNeedLogin(false);
      } catch (err) {
        if (!user) {
          setNeedLogin(true);
          setPosts([]);
        } else {
          setError(getErrorMessage(err, "帖子加载失败，请确认后端服务已启动。"));
          setPosts([]);
        }
      } finally {
        setLoading(false);
      }
    };

    fetchPosts();
  }, [sortBy, page, user, authLoading]);

  const hasNextPage = useMemo(() => posts.length === 10 && (total === 0 || page * 10 < total), [page, posts, total]);

  return (
    <PageShell>
      <section className="rounded-lg border border-border bg-surface p-5">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold text-foreground">今天大家在聊什么</h1>
            <p className="mt-2 text-sm text-muted-strong">按时间或热度浏览社区里的新想法。</p>
          </div>
          <Link
            href="/submit"
            className="inline-flex h-10 items-center justify-center gap-2 rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
          >
            <PenLine className="h-4 w-4" />
            发布帖子
          </Link>
        </div>

        <div className="mt-5 inline-flex rounded-lg bg-surface-soft p-1">
          <button
            onClick={() => {
              setSortBy("create_time");
              setPage(1);
            }}
            className={cn(
              "inline-flex h-9 items-center gap-2 rounded-md px-3 text-sm font-medium transition",
              sortBy === "create_time" ? "bg-surface text-foreground" : "text-muted hover:text-foreground"
            )}
          >
            <Clock3 className="h-4 w-4" />
            最新
          </button>
          <button
            onClick={() => {
              setSortBy("score");
              setPage(1);
            }}
            className={cn(
              "inline-flex h-9 items-center gap-2 rounded-md px-3 text-sm font-medium transition",
              sortBy === "score" ? "bg-surface text-foreground" : "text-muted hover:text-foreground"
            )}
          >
            <Flame className="h-4 w-4" />
            热门
          </button>
        </div>
      </section>

      {error ? <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">{error}</div> : null}

      {needLogin ? (
        <EmptyState
          title="登录后查看帖子内容"
          description="SayIt 会根据你的登录状态返回可浏览的帖子列表。"
          action={
            <Link
              href="/login"
              className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
            >
              去登录
            </Link>
          }
        />
      ) : null}

      <div className="space-y-3">
        {loading
          ? Array.from({ length: 5 }).map((_, index) => <PostCardSkeleton key={index} />)
          : posts.map((post) => <PostCard key={post.post_id} post={post} />)}
      </div>

      {!loading && !needLogin && posts.length === 0 ? (
        <EmptyState
          title="暂时还没有帖子"
          description="成为第一个发起讨论的人，让社区有个漂亮的开场。"
          action={
            <Link
              href="/submit"
              className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
            >
              发布第一篇
            </Link>
          }
        />
      ) : null}

      {total > 10 || page > 1 ? (
        <div className="flex items-center justify-center gap-3 pt-2">
          <Button variant="outline" disabled={page === 1} onClick={() => setPage((value) => Math.max(1, value - 1))}>
            上一页
          </Button>
          <span className="text-sm font-medium text-muted-strong">第 {page} 页</span>
          <Button variant="outline" disabled={!hasNextPage} onClick={() => setPage((value) => value + 1)}>
            下一页
          </Button>
        </div>
      ) : null}
    </PageShell>
  );
}
