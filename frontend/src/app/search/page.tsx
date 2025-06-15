"use client";

import { Suspense, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useSearchParams, useRouter } from "next/navigation";
import { Search } from "lucide-react";
import PostCard from "@/components/ui/PostCard";
import EmptyState from "@/components/ui/EmptyState";
import PageShell from "@/components/ui/PageShell";
import { PostCardSkeleton } from "@/components/ui/Skeleton";
import { apiClient, type PostListItem, type PostsResponse } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { getErrorMessage } from "@/lib/utils";

const PAGE_SIZE = 10;

function normalizePosts(data: PostsResponse | PostListItem[] | undefined) {
  if (!data) return { list: [] as PostListItem[], total: undefined as number | undefined };
  if (Array.isArray(data)) return { list: data, total: undefined as number | undefined };
  return { list: data.list ?? [], total: data.total ?? 0 };
}

function SearchContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user, loading: authLoading } = useAuth();

  const keyword = searchParams.get("keyword") || "";
  const communityName = searchParams.get("community_name") || "";
  const userName = searchParams.get("user_name") || "";

  const [posts, setPosts] = useState<PostListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState("");
  const loadMoreRef = useRef<HTMLDivElement | null>(null);
  const requestIdRef = useRef(0);
  const isFetchingRef = useRef(false);

  const hasSearchParams = keyword || communityName || userName;

  useEffect(() => {
    if (authLoading) return;

    if (!user) {
      setLoading(false);
      return;
    }

    if (!hasSearchParams) {
      setLoading(false);
      setPosts([]);
      return;
    }

    const requestId = requestIdRef.current + 1;
    requestIdRef.current = requestId;
    const isFirstPage = page === 1;

    const fetchPosts = async () => {
      isFetchingRef.current = true;
      setLoading(isFirstPage);
      setLoadingMore(!isFirstPage);
      setError("");
      try {
        const response = await apiClient.getPosts({
          page,
          size: PAGE_SIZE,
          keyword: keyword || undefined,
          community_name: communityName || undefined,
          user_name: userName || undefined,
          sort_by: "create_time",
          order: "desc",
        });
        if (requestId !== requestIdRef.current) return;
        const normalized = normalizePosts(response.data);
        setPosts((currentPosts) => {
          if (isFirstPage) return normalized.list;
          const existingIds = new Set(currentPosts.map((post) => post.post_id));
          const nextPosts = normalized.list.filter((post) => !existingIds.has(post.post_id));
          return [...currentPosts, ...nextPosts];
        });
        setTotal(normalized.total ?? 0);
        setHasMore(normalized.list.length === PAGE_SIZE && (normalized.total === undefined || page * PAGE_SIZE < normalized.total));
      } catch (err) {
        if (requestId !== requestIdRef.current) return;
        setHasMore(false);
        setError(getErrorMessage(err, "搜索失败，请稍后再试。"));
        setPosts([]);
      } finally {
        if (requestId === requestIdRef.current) {
          isFetchingRef.current = false;
          setLoading(false);
          setLoadingMore(false);
        }
      }
    };

    fetchPosts();
  }, [page, user, authLoading, keyword, communityName, userName, hasSearchParams]);

  useEffect(() => {
    const node = loadMoreRef.current;
    if (!node || loading || loadingMore || !hasMore || error) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry.isIntersecting || isFetchingRef.current) return;
        isFetchingRef.current = true;
        setPage((value) => value + 1);
      },
      { rootMargin: "240px 0px" }
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [loading, loadingMore, hasMore, error]);

  const buildSearchDescription = () => {
    const parts = [];
    if (keyword) parts.push(`关键字"${keyword}"`);
    if (communityName) parts.push(`社区"${communityName}"`);
    if (userName) parts.push(`用户"${userName}"`);
    return parts.join(" + ");
  };

  if (!user && !authLoading) {
    return (
      <PageShell>
        <EmptyState
          title="登录后使用搜索功能"
          description="搜索功能需要登录后才能使用。"
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
    <PageShell>
      <section className="rounded-lg border border-border bg-surface p-5">
        <div className="flex items-center gap-3">
          <Search className="h-6 w-6 text-primary" />
          <div>
            <h1 className="text-2xl font-bold text-foreground">搜索结果</h1>
            <p className="mt-1 text-sm text-muted-strong">
              {hasSearchParams
                ? `正在搜索 ${buildSearchDescription()}`
                : "请输入搜索条件"}
            </p>
          </div>
        </div>
      </section>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">
          {error}
        </div>
      ) : null}

      {!hasSearchParams ? (
        <EmptyState
          title="请输入搜索条件"
          description="在顶部搜索栏输入关键字、社区名或用户名来搜索帖子。"
          action={
            <Link
              href="/"
              className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
            >
              返回首页
            </Link>
          }
        />
      ) : null}

      <div className="space-y-3">
        {loading
          ? Array.from({ length: 5 }).map((_, index) => <PostCardSkeleton key={index} />)
          : posts.map((post) => <PostCard key={post.post_id} post={post} />)}
      </div>

      {!loading && loadingMore ? (
        <div className="space-y-3">
          {Array.from({ length: 2 }).map((_, index) => (
            <PostCardSkeleton key={`loading-more-${index}`} />
          ))}
        </div>
      ) : null}

      {!loading && hasSearchParams && posts.length === 0 && !error ? (
        <EmptyState
          title="没有找到匹配的帖子"
          description="尝试使用不同的关键字或筛选条件。"
          action={
            <Link
              href="/"
              className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
            >
              返回首页
            </Link>
          }
        />
      ) : null}

      {!loading && posts.length > 0 ? (
        <div ref={loadMoreRef} className="flex justify-center py-3 text-sm text-muted-strong">
          {hasMore ? "继续向下滚动加载更多" : total > PAGE_SIZE || page > 1 ? "已经到底了" : null}
        </div>
      ) : null}
    </PageShell>
  );
}

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <PageShell>
          <section className="rounded-lg border border-border bg-surface p-5">
            <div className="flex items-center gap-3">
              <Search className="h-6 w-6 text-primary" />
              <div>
                <h1 className="text-2xl font-bold text-foreground">搜索结果</h1>
                <p className="mt-1 text-sm text-muted-strong">正在加载...</p>
              </div>
            </div>
          </section>
          <div className="space-y-3">
            {Array.from({ length: 5 }).map((_, index) => (
              <PostCardSkeleton key={index} />
            ))}
          </div>
        </PageShell>
      }
    >
      <SearchContent />
    </Suspense>
  );
}
