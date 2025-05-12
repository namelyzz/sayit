"use client";

import Link from "next/link";
import { useState } from "react";
import { ArrowDown, ArrowUp, MessageSquare, Share2 } from "lucide-react";
import { apiClient, type PostListItem } from "@/lib/api";
import { formatCount, formatDateTime } from "@/lib/format";
import CommunityBadge from "@/components/ui/CommunityBadge";
import { cn, getErrorMessage } from "@/lib/utils";

interface PostCardProps {
  post: PostListItem;
}

export default function PostCard({ post }: PostCardProps) {
  const [vote, setVote] = useState(post.current_user_vote || 0);
  const [voteCount, setVoteCount] = useState(post.vote_count || 0);
  const [error, setError] = useState("");

  const handleVote = async (direction: number) => {
    const nextVote = vote === direction ? 0 : direction;
    setError("");

    try {
      await apiClient.vote(post.post_id, nextVote);
      setVoteCount((currentVoteCount) => currentVoteCount + nextVote - vote);
      setVote(nextVote);
    } catch (err) {
      setError(getErrorMessage(err, "投票失败，请稍后重试。"));
    }
  };

  return (
    <article className="group rounded-lg border border-border bg-surface p-4 shadow-[0_6px_18px_rgba(69,123,157,0.10)] transition hover:border-border-strong hover:shadow-[0_10px_24px_rgba(69,123,157,0.14)]">
      <div className="flex gap-4">
        <div className="hidden w-11 shrink-0 flex-col items-center rounded-lg bg-surface-soft py-2 text-muted-strong sm:flex">
          <button
            onClick={() => handleVote(1)}
            className={cn(
              "rounded-md p-1 transition hover:bg-surface hover:text-primary",
              vote === 1 && "bg-primary text-white hover:bg-primary hover:text-white"
            )}
            aria-label="赞同"
          >
            <ArrowUp className="h-4 w-4" />
          </button>
          <span className="my-1 text-sm font-bold text-foreground">{formatCount(voteCount)}</span>
          <button
            onClick={() => handleVote(-1)}
            className={cn(
              "rounded-md p-1 transition hover:bg-surface hover:text-danger",
              vote === -1 && "bg-danger text-white hover:bg-danger hover:text-white"
            )}
            aria-label="反对"
          >
            <ArrowDown className="h-4 w-4" />
          </button>
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-x-2 gap-y-2 text-xs text-muted">
            <CommunityBadge
              id={post.community_id}
              name={post.community_name}
              href={`/community/${post.community_id}`}
            />
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
              {formatCount(voteCount)}
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

          {error ? <p className="mt-3 text-sm text-danger">{error}</p> : null}
        </div>
      </div>
    </article>
  );
}
