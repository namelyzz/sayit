"use client";

import { useState } from "react";
import Link from "next/link";
import { Heart, MessageSquare, Trash2, ChevronDown, ChevronUp } from "lucide-react";
import { apiClient, type CommentDetail } from "@/lib/api";
import { formatDateTime, formatCount } from "@/lib/format";
import { cn } from "@/lib/utils";
import CommentForm from "./CommentForm";

const childPageSize = 5;

interface CommentItemProps {
  comment: CommentDetail;
  currentUserId?: string;
  postAuthorId: string;
  order: "desc" | "asc";
  onReply: (parentId: string, content: string) => Promise<void>;
  onDelete: (commentId: string) => Promise<void>;
  onLike: (commentId: string) => Promise<void>;
  onUnlike: (commentId: string) => Promise<void>;
  depth?: number;
}

export default function CommentItem({
  comment,
  currentUserId,
  postAuthorId,
  order,
  onReply,
  onDelete,
  onLike,
  onUnlike,
  depth = 0,
}: CommentItemProps) {
  const [showReplyForm, setShowReplyForm] = useState(false);
  const [liked, setLiked] = useState(comment.is_liked);
  const [likeCount, setLikeCount] = useState(comment.like_count);
  const [isDeleted, setIsDeleted] = useState(comment.status === 2);

  // 子评论懒加载状态
  const [showChildren, setShowChildren] = useState(false);
  const [children, setChildren] = useState<CommentDetail[]>([]);
  const [childrenPage, setChildrenPage] = useState(0);
  const [childrenHasMore, setChildrenHasMore] = useState(false);
  const [loadingChildren, setLoadingChildren] = useState(false);

  const isAuthor = currentUserId === comment.author_id;
  const isPostAuthor = currentUserId === postAuthorId;
  const canDelete = isAuthor || isPostAuthor;
  const isDeletedContent = comment.status === 2;

  const handleLike = async () => {
    if (!currentUserId) return;
    try {
      if (liked) {
        await onUnlike(comment.comment_id);
        setLiked(false);
        setLikeCount((prev) => Math.max(0, prev - 1));
      } else {
        await onLike(comment.comment_id);
        setLiked(true);
        setLikeCount((prev) => prev + 1);
      }
    } catch {
      // 静默处理
    }
  };

  const handleDelete = async () => {
    if (!confirm("确定要删除这条评论吗？")) return;
    try {
      await onDelete(comment.comment_id);
      setIsDeleted(true);
    } catch {
      // 静默处理
    }
  };

  const handleReply = async (content: string) => {
    await onReply(comment.comment_id, content);
    setShowReplyForm(false);
  };

  const loadChildren = async (page: number) => {
    setLoadingChildren(true);
    try {
      const response = await apiClient.getCommentChildren(
        comment.comment_id,
        page,
        childPageSize,
        order
      );
      const data = response.data;
      if (page === 1) {
        setChildren(data.list);
      } else {
        setChildren((prev) => [...prev, ...data.list]);
      }
      setChildrenPage(page);
      setChildrenHasMore(data.has_more);
    } catch {
      // 静默处理
    } finally {
      setLoadingChildren(false);
    }
  };

  const handleToggleChildren = async () => {
    if (showChildren) {
      setShowChildren(false);
    } else {
      setShowChildren(true);
      if (children.length === 0 && comment.child_count > 0) {
        await loadChildren(1);
      }
    }
  };

  const handleLoadMoreChildren = async () => {
    if (loadingChildren || !childrenHasMore) return;
    await loadChildren(childrenPage + 1);
  };

  if (isDeleted && comment.child_count === 0) {
    return null;
  }

  return (
    <div className={cn("group", depth > 0 && "ml-6 border-l-2 border-border pl-4")}>
      <div className="py-3">
        {/* 头部信息 */}
        <div className="flex items-center gap-2 text-sm">
          <Link
            href={`/user/${comment.author_id}`}
            className="font-semibold text-muted-strong hover:text-primary"
          >
            {comment.author_name || "匿名用户"}
          </Link>
          <span className="text-muted">·</span>
          <span className="text-muted">{formatDateTime(comment.create_time)}</span>
          {isDeletedContent && (
            <>
              <span className="text-muted">·</span>
              <span className="text-xs text-danger">已删除</span>
            </>
          )}
        </div>

        {/* 内容 */}
        <div className="mt-2 text-sm leading-relaxed text-foreground">
          {isDeletedContent ? (
            <span className="italic text-muted">[已删除]</span>
          ) : (
            <p className="whitespace-pre-wrap">{comment.content}</p>
          )}
        </div>

        {/* 操作栏 */}
        {!isDeletedContent && (
          <div className="mt-2 flex items-center gap-1">
            {/* 点赞 */}
            <button
              onClick={handleLike}
              className={cn(
                "inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs transition",
                liked
                  ? "text-danger"
                  : "text-muted hover:bg-surface-soft hover:text-foreground"
              )}
              disabled={!currentUserId}
            >
              <Heart className={cn("h-3.5 w-3.5", liked && "fill-current")} />
              {likeCount > 0 && formatCount(likeCount)}
            </button>

            {/* 回复 */}
            <button
              onClick={() => setShowReplyForm(!showReplyForm)}
              className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs text-muted transition hover:bg-surface-soft hover:text-foreground"
            >
              <MessageSquare className="h-3.5 w-3.5" />
              回复
            </button>

            {/* 删除 */}
            {canDelete && (
              <button
                onClick={handleDelete}
                className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs text-muted transition hover:bg-surface-soft hover:text-danger"
              >
                <Trash2 className="h-3.5 w-3.5" />
                删除
              </button>
            )}
          </div>
        )}

        {/* 回复表单 */}
        {showReplyForm && (
          <div className="mt-3">
            <CommentForm
              onSubmit={handleReply}
              placeholder={`回复 ${comment.author_name}...`}
              autoFocus
              compact
            />
          </div>
        )}
      </div>

      {/* 子评论 */}
      {comment.child_count > 0 && (
        <div>
          <button
            onClick={handleToggleChildren}
            className="mb-1 inline-flex items-center gap-1 text-xs text-primary hover:underline"
          >
            {showChildren ? (
              <>
                <ChevronUp className="h-3 w-3" />
                收起回复
              </>
            ) : (
              <>
                <ChevronDown className="h-3 w-3" />
                展开 {comment.child_count} 条回复
              </>
            )}
          </button>
          {showChildren && (
            <div>
              {children.map((child) => (
                <CommentItem
                  key={child.comment_id}
                  comment={child}
                  currentUserId={currentUserId}
                  postAuthorId={postAuthorId}
                  order={order}
                  onReply={onReply}
                  onDelete={onDelete}
                  onLike={onLike}
                  onUnlike={onUnlike}
                  depth={depth + 1}
                />
              ))}
              {childrenHasMore && (
                <button
                  onClick={handleLoadMoreChildren}
                  disabled={loadingChildren}
                  className="ml-6 mb-2 text-xs text-primary hover:underline disabled:opacity-50"
                >
                  {loadingChildren ? "加载中..." : "加载更多回复"}
                </button>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
