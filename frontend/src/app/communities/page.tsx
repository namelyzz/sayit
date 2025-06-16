"use client";

import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Globe, Plus, Search, Users } from "lucide-react";
import PageShell from "@/components/ui/PageShell";
import EmptyState from "@/components/ui/EmptyState";
import { apiClient, type CommunityListItem, type CommunityListResponse } from "@/lib/api";
import { cn, getErrorMessage } from "@/lib/utils";
import { getCommunityPalette } from "@/lib/community-colors";

const PAGE_SIZE = 20;

function CommunityCard({ community }: { community: CommunityListItem }) {
  const palette = getCommunityPalette(community.community_id || community.name);

  return (
    <Link
      href={`/community/${community.community_id}`}
      className="group rounded-xl border border-border bg-surface p-4 transition hover:border-primary/30 hover:shadow-md"
    >
      <div className="flex items-start gap-3">
        <div
          className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-sm font-bold"
          style={{
            backgroundColor: palette.bg,
            color: palette.text,
          }}
        >
          {community.name.charAt(0).toUpperCase()}
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="truncate font-semibold text-foreground group-hover:text-primary">
            {community.name}
          </h3>
          <p className="mt-1 line-clamp-2 text-sm text-muted-strong">
            {community.introduction || "暂无简介"}
          </p>
        </div>
      </div>
      <div className="mt-3 flex items-center gap-4 text-xs text-muted">
        <span className="flex items-center gap-1">
          <Users className="h-3.5 w-3.5" />
          {community.post_count} 帖子
        </span>
      </div>
    </Link>
  );
}

function CommunityCardSkeleton() {
  return (
    <div className="rounded-xl border border-border bg-surface p-4">
      <div className="flex items-start gap-3">
        <div className="h-10 w-10 animate-pulse rounded-lg bg-muted/30" />
        <div className="min-w-0 flex-1 space-y-2">
          <div className="h-5 w-24 animate-pulse rounded bg-muted/30" />
          <div className="h-4 w-full animate-pulse rounded bg-muted/30" />
        </div>
      </div>
      <div className="mt-3 flex items-center gap-4">
        <div className="h-4 w-16 animate-pulse rounded bg-muted/30" />
      </div>
    </div>
  );
}

function CommunitiesContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [communities, setCommunities] = useState<CommunityListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [keyword, setKeyword] = useState(searchParams.get("keyword") || "");
  const [searchInput, setSearchInput] = useState(searchParams.get("keyword") || "");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState("");
  const loadMoreRef = useRef<HTMLDivElement | null>(null);
  const requestIdRef = useRef(0);
  const isFetchingRef = useRef(false);

  // 监听 URL 参数变化
  useEffect(() => {
    const urlKeyword = searchParams.get("keyword") || "";
    setKeyword(urlKeyword);
    setSearchInput(urlKeyword);
    setPage(1);
    setCommunities([]);
  }, [searchParams]);

  const fetchCommunities = useCallback(
    async (pageNum: number, searchKeyword: string) => {
      const requestId = requestIdRef.current + 1;
      requestIdRef.current = requestId;
      const isFirstPage = pageNum === 1;

      isFetchingRef.current = true;
      setLoading(isFirstPage);
      setLoadingMore(!isFirstPage);
      setError("");

      try {
        const response = await apiClient.getCommunitiesWithSearch({
          page: pageNum,
          size: PAGE_SIZE,
          keyword: searchKeyword || undefined,
        });

        if (requestId !== requestIdRef.current) return;

        const data = response.data as CommunityListResponse;
        const list = data?.list ?? [];
        const total = data?.total ?? 0;

        setCommunities((current) => {
          if (isFirstPage) return list;
          const existingIds = new Set(current.map((c) => c.community_id));
          const newItems = list.filter((c) => !existingIds.has(c.community_id));
          return [...current, ...newItems];
        });

        setTotal(total);
        setHasMore(list.length === PAGE_SIZE && pageNum * PAGE_SIZE < total);
      } catch (err) {
        if (requestId !== requestIdRef.current) return;
        setHasMore(false);
        setError(getErrorMessage(err, "获取社区列表失败，请稍后再试。"));
        if (isFirstPage) setCommunities([]);
      } finally {
        if (requestId === requestIdRef.current) {
          isFetchingRef.current = false;
          setLoading(false);
          setLoadingMore(false);
        }
      }
    },
    []
  );

  useEffect(() => {
    fetchCommunities(page, keyword);
  }, [page, keyword, fetchCommunities]);

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

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    const params = new URLSearchParams(searchParams.toString());
    if (searchInput.trim()) {
      params.set("keyword", searchInput.trim());
    } else {
      params.delete("keyword");
    }
    router.push(`/communities?${params.toString()}`);
  };

  return (
    <PageShell>
      <section className="rounded-lg border border-border bg-surface p-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Globe className="h-6 w-6 text-primary" />
            <div>
              <h1 className="text-2xl font-bold text-foreground">社区列表</h1>
              <p className="mt-1 text-sm text-muted-strong">
                发现并加入感兴趣的社区
              </p>
            </div>
          </div>
          <Link
            href="/communities/create"
            className="inline-flex h-10 items-center justify-center gap-2 rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
          >
            <Plus className="h-4 w-4" />
            申请创建社区
          </Link>
        </div>
      </section>

      <form onSubmit={handleSearch} className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
          <input
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="搜索社区名称或简介..."
            className="h-10 w-full rounded-lg border border-border bg-surface pl-9 pr-4 text-sm text-foreground placeholder:text-muted focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
          />
        </div>
        <button
          type="submit"
          className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
        >
          搜索
        </button>
      </form>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">
          {error}
        </div>
      ) : null}

      <div className="grid gap-3 sm:grid-cols-2">
        {loading
          ? Array.from({ length: 6 }).map((_, index) => (
              <CommunityCardSkeleton key={index} />
            ))
          : communities.map((community) => (
              <CommunityCard
                key={community.community_id}
                community={community}
              />
            ))}
      </div>

      {!loading && loadingMore ? (
        <div className="grid gap-3 sm:grid-cols-2">
          {Array.from({ length: 2 }).map((_, index) => (
            <CommunityCardSkeleton key={`loading-more-${index}`} />
          ))}
        </div>
      ) : null}

      {!loading && communities.length === 0 && !error ? (
        <EmptyState
          title={keyword ? "没有找到匹配的社区" : "暂无社区"}
          description={
            keyword
              ? "尝试使用不同的关键字搜索。"
              : "社区功能正在建设中，敬请期待。"
          }
          action={
            keyword ? (
              <button
                onClick={() => {
                  router.push("/communities");
                }}
                className="inline-flex h-10 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-white transition hover:bg-primary-dark"
              >
                清除搜索
              </button>
            ) : undefined
          }
        />
      ) : null}

      {!loading && communities.length > 0 ? (
        <div
          ref={loadMoreRef}
          className="flex justify-center py-3 text-sm text-muted-strong"
        >
          {hasMore
            ? "继续向下滚动加载更多"
            : total > PAGE_SIZE || page > 1
            ? "已经到底了"
            : null}
        </div>
      ) : null}
    </PageShell>
  );
}

export default function CommunitiesPage() {
  return (
    <Suspense
      fallback={
        <PageShell>
          <section className="rounded-lg border border-border bg-surface p-5">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Globe className="h-6 w-6 text-primary" />
                <div>
                  <h1 className="text-2xl font-bold text-foreground">社区列表</h1>
                  <p className="mt-1 text-sm text-muted-strong">正在加载...</p>
                </div>
              </div>
            </div>
          </section>
          <div className="grid gap-3 sm:grid-cols-2">
            {Array.from({ length: 6 }).map((_, index) => (
              <div key={index} className="rounded-xl border border-border bg-surface p-4">
                <div className="flex items-start gap-3">
                  <div className="h-10 w-10 animate-pulse rounded-lg bg-muted/30" />
                  <div className="min-w-0 flex-1 space-y-2">
                    <div className="h-5 w-24 animate-pulse rounded bg-muted/30" />
                    <div className="h-4 w-full animate-pulse rounded bg-muted/30" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </PageShell>
      }
    >
      <CommunitiesContent />
    </Suspense>
  );
}
