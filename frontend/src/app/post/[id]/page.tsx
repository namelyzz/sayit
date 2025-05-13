"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowDown, ArrowLeft, ArrowUp, MessageSquare, Share2 } from "lucide-react";
import Button from "@/components/ui/Button";
import EmptyState from "@/components/ui/EmptyState";
import PageShell from "@/components/ui/PageShell";
import { Skeleton } from "@/components/ui/Skeleton";
import { Textarea } from "@/components/ui/Field";
import { apiClient, type PostDetail } from "@/lib/api";
import { formatCount, formatDateTime } from "@/lib/format";
import { cn, getErrorMessage } from "@/lib/utils";

const SCORE_PER_VOTE = 432;

export default function PostDetailPage() {
  const params = useParams();
  const postId = params.id as string;

  const [post, setPost] = useState<PostDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [vote, setVote] = useState(0);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchPost = async () => {
      setLoading(true);
      setError("");
      try {
        const response = await apiClient.getPostDetail(postId);
        setPost(response.data);
        setVote(response.data.current_user_vote || 0);
      } catch (err) {
        setError(getErrorMessage(err, "帖子不存在或加载失败。"));
      } finally {
        setLoading(false);
      }
    };

    fetchPost();
  }, [postId]);

  const handleVote = async (direction: number) => {
    const newVote = vote === direction ? 0 : direction;
    try {
      await apiClient.vote(postId, newVote);
      setPost((currentPost) =>
        currentPost
          ? {
              ...currentPost,
              like_count: (currentPost.like_count || 0) + (newVote - vote) * SCORE_PER_VOTE,
              vote_count: (currentPost.vote_count || 0) + newVote - vote,
              current_user_vote: newVote,
            }
          : currentPost
      );
      setVote(newVote);
    } catch (err) {
      setError(getErrorMessage(err, "投票失败，请稍后重试。"));
    }
  };

  if (loading) {
    return (
      <PageShell>
        <section className="rounded-lg border border-border bg-surface p-6 shadow-sm">
          <Skeleton className="h-4 w-40" />
          <Skeleton className="mt-4 h-8 w-5/6" />
          <Skeleton className="mt-8 h-4 w-full" />
          <Skeleton className="mt-3 h-4 w-full" />
          <Skeleton className="mt-3 h-4 w-2/3" />
        </section>
      </PageShell>
    );
  }

  if (error && !post) {
    return (
      <PageShell>
        <EmptyState
          title="帖子加载失败"
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

  if (!post) return null;

  return (
    <PageShell>
      <Link
        href="/"
        className="inline-flex items-center gap-1.5 text-sm font-medium text-muted-strong transition hover:text-primary"
      >
        <ArrowLeft className="h-4 w-4" />
        返回信息流
      </Link>

      {error ? <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">{error}</div> : null}

      <article className="rounded-lg border border-border bg-surface shadow-sm">
        <div className="flex">
          <div className="hidden w-16 shrink-0 border-r border-border bg-surface-soft p-3 sm:block">
            <div className="sticky top-24 flex flex-col items-center">
              <button
                onClick={() => handleVote(1)}
                className={cn(
                  "rounded-lg p-2 transition",
                  vote === 1 ? "bg-primary text-white" : "text-muted hover:bg-surface hover:text-primary"
                )}
                aria-label="赞同"
              >
                <ArrowUp className="h-5 w-5" />
              </button>
              <span className="my-2 text-base font-bold text-foreground">{formatCount(post.vote_count || 0)}</span>
              <button
                onClick={() => handleVote(-1)}
                className={cn(
                  "rounded-lg p-2 transition",
                  vote === -1 ? "bg-danger text-white" : "text-muted hover:bg-surface hover:text-danger"
                )}
                aria-label="反对"
              >
                <ArrowDown className="h-5 w-5" />
              </button>
            </div>
          </div>

          <div className="min-w-0 flex-1 p-5 sm:p-7">
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted">
              {post.community ? (
                <Link href={`/community/${post.community.community_id}`} className="font-semibold text-primary hover:text-primary-dark">
                  {post.community.name}
                </Link>
              ) : null}
              <span>
                由{" "}
                {post.author_id ? (
                  <Link
                    href={`/user/${post.author_id}`}
                    className="rounded-md px-1 py-0.5 font-semibold text-muted-strong transition hover:bg-[#457b9d]/10 hover:text-primary focus:bg-[#457b9d]/10 focus:text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    {post.author_name || "匿名用户"}
                  </Link>
                ) : (
                  post.author_name || "匿名用户"
                )}{" "}
                发布
              </span>
              <span>{formatDateTime(post.create_time)}</span>
            </div>

            <h1 className="mt-4 text-2xl font-bold leading-9 text-foreground sm:text-3xl sm:leading-10">{post.title}</h1>

            <div className="prose-lite mt-7 whitespace-pre-wrap text-base">{post.content}</div>

            <div className="mt-8 flex flex-wrap items-center gap-2 border-t border-border pt-5 text-sm text-muted">
              <span className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-surface-soft px-3 font-medium sm:hidden">
                <ArrowUp className="h-4 w-4 text-primary" />
                {formatCount(post.vote_count || 0)}
              </span>
              <span className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-surface-soft px-3 font-medium">
                <MessageSquare className="h-4 w-4" />
                {formatCount(post.comment_count)} 条评论
              </span>
              <button className="inline-flex h-9 items-center gap-1.5 rounded-lg px-3 font-medium transition hover:bg-surface-soft hover:text-foreground">
                <Share2 className="h-4 w-4" />
                分享
              </button>
            </div>
          </div>
        </div>
      </article>

      <section className="rounded-lg border border-border bg-surface p-5 shadow-sm">
        <h2 className="text-lg font-semibold text-foreground">评论</h2>
        <div className="mt-4">
          <Textarea rows={3} placeholder="写下你的评论..." />
          <div className="mt-3 flex justify-end">
            <Button>
              <MessageSquare className="h-4 w-4" />
              发表评论
            </Button>
          </div>
        </div>
        <div className="mt-6 rounded-lg bg-surface-soft px-4 py-8 text-center text-sm text-muted">
          暂无评论。等第一条有分量的回复。
        </div>
      </section>
    </PageShell>
  );
}
