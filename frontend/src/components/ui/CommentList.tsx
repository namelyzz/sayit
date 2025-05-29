"use client";

import { useState, useEffect, useCallback } from "react";
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

  const pageSize = 20;

  const fetchComments = useCallback(async (pageNum: number, append = false) => {
    if (pageNum === 1) {
      setLoading(true);
    } else {
      setLoadingMore(true);
    }
    setError("");

    try {
      const response = await apiClient.getCommentList(postId, pageNum, pageSize, order);
      const data = response.data;

      if (append) {
        setComments((prev) => [...prev, ...data.list]);
      } else {
        setComments(data.list);
      }

      setTotal(data.total);
      setHasMore(data.list.length === pageSize);
    } catch (err) {
      setError(getErrorMessage(err, "加载评论失败"));
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, [postId, order]);

  useEffect(() => {
    setPage(1);
    fetchComments(1);
  }, [fetchComments]);

  const handleLoadMore = () => {
    if (loadingMore || !hasMore) return;
    const nextPage = page + 1;
    setPage(nextPage);
    fetchComments(nextPage, true);
  };

  const handleCreateComment = async (content: string) => {
    await apiClient.createComment(postId, content);
    // 刷新评论列表
    setPage(1);
    await fetchComments(1);
  };

  const handleReply = async (parentId: string, content: string) => {
    await apiClient.createComment(postId, content, parentId);
    // 刷新评论列表
    setPage(1);
    await fetchComments(1);
  };

  const handleDelete = async (commentId: string) => {
    await apiClient.deleteComment(commentId);
    // 刷新评论列表
    await fetchComments(1);
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
                onReply={handleReply}
                onDelete={handleDelete}
                onLike={handleLike}
                onUnlike={handleUnlike}
              />
            ))}
          </div>

          {/* 加载更多 */}
          {hasMore && (
            <div className="mt-6 text-center">
              <button
                onClick={handleLoadMore}
                disabled={loadingMore}
                className="rounded-lg border border-border bg-surface px-6 py-2 text-sm font-medium text-muted-strong transition hover:bg-surface-soft hover:text-foreground disabled:opacity-50"
              >
                {loadingMore ? "加载中..." : "加载更多评论"}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
