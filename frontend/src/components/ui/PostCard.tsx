"use client";

import Link from "next/link";
import { ArrowDown, ArrowUp, MessageSquare, Share2 } from "lucide-react";
import type { PostListItem } from "@/lib/api";
import { formatCount, formatDateTime } from "@/lib/format";

interface PostCardProps {
  post: PostListItem;
}

export default function PostCard({ post }: PostCardProps) {
  return (
    <article className="group rounded-lg border border-border bg-surface p-4 shadow-sm transition hover:border-border-strong hover:shadow-md">
      <div className="flex gap-4">
        <div className="hidden w-11 shrink-0 flex-col items-center rounded-lg bg-surface-soft py-2 text-muted-strong sm:flex">
          <button className="rounded-md p-1 transition hover:bg-surface hover:text-primary" aria-label="赞同">
            <ArrowUp className="h-4 w-4" />
          </button>
          <span className="my-1 text-sm font-bold text-foreground">{formatCount(post.like_count)}</span>
          <button className="rounded-md p-1 transition hover:bg-surface hover:text-danger" aria-label="反对">
            <ArrowDown className="h-4 w-4" />
          </button>
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted">
            <Link
              href={`/community/${post.community_id}`}
              className="font-semibold text-primary transition hover:text-primary-dark"
            >
              {post.community_name || "未命名社区"}
            </Link>
            <span>由 {post.user_name || "匿名用户"} 发布</span>
            <span>{formatDateTime(post.create_time)}</span>
          </div>

          <Link href={`/post/${post.post_id}`} className="mt-2 block">
            <h2 className="line-clamp-2 text-lg font-semibold leading-7 text-foreground transition group-hover:text-primary">
              {post.title || "无标题帖子"}
            </h2>
          </Link>

          {post.summary ? (
            <p className="mt-2 line-clamp-3 text-sm leading-6 text-muted-strong">{post.summary}</p>
          ) : null}

          <div className="mt-4 flex flex-wrap items-center gap-2 text-sm text-muted">
            <span className="inline-flex h-8 items-center gap-1.5 rounded-lg bg-surface-soft px-2.5 font-medium sm:hidden">
              <ArrowUp className="h-4 w-4 text-primary" />
              {formatCount(post.like_count)}
            </span>
            <Link
              href={`/post/${post.post_id}`}
              className="inline-flex h-8 items-center gap-1.5 rounded-lg px-2.5 font-medium transition hover:bg-surface-soft hover:text-foreground"
            >
              <MessageSquare className="h-4 w-4" />
              {formatCount(post.comment_count)} 条评论
            </Link>
            <button className="inline-flex h-8 items-center gap-1.5 rounded-lg px-2.5 font-medium transition hover:bg-surface-soft hover:text-foreground">
              <Share2 className="h-4 w-4" />
              分享
            </button>
          </div>
        </div>
      </div>
    </article>
  );
}
