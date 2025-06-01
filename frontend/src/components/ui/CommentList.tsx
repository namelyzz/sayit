"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { MessageSquare, ArrowUpDown } from "lucide-react";
import { apiClient, type CommentDetail } from "@/lib/api";
import { getErrorMessage } from "@/lib/utils";
import { Skeleton } from "./Skeleton";
import EmptyState from "./EmptyState";
import CommentItem from "./CommentItem";
import CommentForm from "./CommentForm";

interface CommentListProps {
  postId: string;
  currentUserId?: string;
  postAuthorId: string;
}

export default function CommentList({ postId, currentUserId, postAuthorId }: CommentListProps) {
  const [comments, setComments] = useState<CommentDetail[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");
  const [hasMore, setHasMore] = useState(true);
  const [order, setOrder] = useState<"desc" | "asc">("desc");

  const pageSize = 5;
  const sentinelRef = useRef<HTMLDivElement>(null);
  const fetchIdRef = useRef(0);

  const fetchComments = useCallback(async (pageNum: number, append = false) => {
    const currentFetchId = ++fetchIdRef.current;

    if (pageNum === 1) {
      setLoading(true);
    } else {
      setLoadingMore(true);
    }
    setError("");

    try {
      const response = await apiClient.getCommentList(postId, pageNum, pageSize, order);
      const data = response.data;

      // 如果期间有新的请求发起，丢弃本次结果
      if (currentFetchId !== fetchIdRef.current) return;

      if (append) {
        setComments((prev) => [...prev, ...data.list]);
      } else {
        setComments(data.list);
      }

      setTotal(data.total);
      setHasMore(data.list.length === pageSize);
    } catch (err) {
      if (currentFetchId !== fetchIdRef.current) return;
      setError(getErrorMessage(err, "加载评论失败"));
    } finally {
      if (currentFetchId === fetchIdRef.current) {
        setLoading(false);
        setLoadingMore(false);
      }
    }
  }, [postId, order]);

  useEffect(() => {
    setPage(1);
    fetchComments(1);
  }, [fetchComments]);

  const handleLoadMore = useCallback(() => {
    if (loadingMore || !hasMore) return;
    setPage((prev) => {
      const nextPage = prev + 1;
      fetchComments(nextPage, true);
      return nextPage;
    });
  }, [loadingMore, hasMore, fetchComments]);

  // IntersectionObserver 实现无限滚动
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loadingMore && !loading) {
          handleLoadMore();
        }
      },
      { threshold: 0.1 }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasMore, loadingMore, loading, handleLoadMore]);

  const handleCreateComment = async (content: string) => {
    await apiClient.createComment(postId, content);
    setPage(1);
    fetchComments(1);
  };

  const handleReply = async (parentId: string, content: string) => {
    await apiClient.createComment(postId, content, parentId);
    setPage(1);
    fetchComments(1);
  };

  const handleDelete = async (commentId: string) => {
    await apiClient.deleteComment(commentId);
    setPage(1);
    fetchComments(1);
  };

  const handleLike = async (commentId: string) => {
    await apiClient.likeComment(commentId);
  };

  const handleUnlike = async (commentId: string) => {
    await apiClient.unlikeComment(commentId);
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-20 w-full" />
      </div>
    );
  }

  return (
    <div>
      {/* 评论输入 */}
      {currentUserId && (
        <div className="mb-6">
          <CommentForm onSubmit={handleCreateComment} />
        </div>
      )}

      {/* 错误提示 */}
      {error && (
        <div className="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">
          {error}
        </div>
      )}

      {/* 评论列表 */}
      {comments.length === 0 ? (
        <EmptyState
          title="暂无评论"
          description={currentUserId ? "等第一条有分量的回复。" : "登录后可以发表评论。"}
        />
      ) : (
        <div>
          <div className="mb-4 flex items-center justify-between">
            <span className="text-sm text-muted">共 {total} 条评论</span>
            <button
              onClick={() => setOrder(order === "desc" ? "asc" : "desc")}
              className="flex items-center gap-1 rounded-lg px-3 py-1.5 text-sm text-muted-strong transition hover:bg-surface-soft hover:text-foreground"
            >
              <ArrowUpDown className="h-4 w-4" />
              {order === "desc" ? "最新" : "最早"}
            </button>
          </div>

          <div className="space-y-1">
            {comments.map((comment) => (
              <CommentItem
                key={comment.comment_id}
                comment={comment}
                currentUserId={currentUserId}
                postAuthorId={postAuthorId}
                order={order}
                onReply={handleReply}
                onDelete={handleDelete}
                onLike={handleLike}
                onUnlike={handleUnlike}
              />
            ))}
          </div>

          {/* 无限滚动哨兵元素 + 加载指示器 */}
          {hasMore && (
            <div ref={sentinelRef} className="py-4 text-center">
              {loadingMore && (
                <div className="flex items-center justify-center gap-2 text-sm text-muted">
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-muted border-t-primary" />
                  加载中...
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
