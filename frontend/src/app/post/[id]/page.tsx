"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowBigUp, ArrowBigDown, MessageSquare, Share2, ArrowLeft } from "lucide-react";
import { apiClient } from "@/lib/api";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type PostDetailData = any;

export default function PostDetailPage() {
  const params = useParams();
  const postId = params.id as string;

  const [post, setPost] = useState<PostDetailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [vote, setVote] = useState(0);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchPost = async () => {
      setLoading(true);
      try {
        const response = await apiClient.getPostDetail(postId);
        setPost(response.data);
      } catch (err) {
        console.error("Failed to fetch post detail:", err);
        setError("帖子不存在或加载失败");
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
      setVote(newVote);
    } catch (err) {
      console.error("Vote failed:", err);
      alert("投票失败，请重试");
    }
  };

  const formatTime = (timeStr: string) => {
    try {
      const date = new Date(timeStr);
      return date.toLocaleString("zh-CN");
    } catch {
      return timeStr;
    }
  };

  if (loading) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-3/4 mb-4"></div>
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-8"></div>
          <div className="space-y-3">
            <div className="h-4 bg-gray-200 rounded w-full"></div>
            <div className="h-4 bg-gray-200 rounded w-full"></div>
            <div className="h-4 bg-gray-200 rounded w-2/3"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !post) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
          <p className="text-gray-500">{error || "帖子不存在"}</p>
          <Link href="/" className="text-primary hover:underline mt-4 inline-block">
            返回首页
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <Link
        href="/"
        className="inline-flex items-center text-gray-600 hover:text-primary mb-4 transition-colors"
      >
        <ArrowLeft className="h-4 w-4 mr-1" />
        返回
      </Link>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="flex">
          <div className="flex flex-col items-center p-4 bg-gray-50 rounded-l-lg">
            <button
              onClick={() => handleVote(1)}
              className={`p-1 rounded transition-colors ${
                vote === 1
                  ? "bg-primary text-white"
                  : "hover:bg-gray-200 text-gray-500 hover:text-primary"
              }`}
            >
              <ArrowBigUp className="h-8 w-8" />
            </button>
            <span className="text-lg font-bold text-gray-700 my-2">
              {(post.like_count || 0) + vote}
            </span>
            <button
              onClick={() => handleVote(-1)}
              className={`p-1 rounded transition-colors ${
                vote === -1
                  ? "bg-red-500 text-white"
                  : "hover:bg-gray-200 text-gray-500 hover:text-red-500"
              }`}
            >
              <ArrowBigDown className="h-8 w-8" />
            </button>
          </div>

          <div className="flex-1 p-6">
            <div className="flex items-center text-sm text-gray-500 mb-4">
              {post.community && (
                <>
                  <Link
                    href={`/community/${post.community.community_id}`}
                    className="font-medium text-primary hover:underline"
                  >
                    {post.community.name}
                  </Link>
                  <span className="mx-2">•</span>
                </>
              )}
              <span>发布者: {post.author_name}</span>
              <span className="mx-2">•</span>
              <span>{formatTime(post.create_time)}</span>
            </div>

            <h1 className="text-2xl font-bold text-gray-900 mb-6">
              {post.title}
            </h1>

            <div className="prose prose-lg max-w-none text-gray-700 leading-relaxed whitespace-pre-wrap">
              {post.content}
            </div>

            <div className="flex items-center space-x-4 mt-8 pt-6 border-t border-gray-200 text-sm text-gray-500">
              <span className="flex items-center space-x-1">
                <MessageSquare className="h-5 w-5" />
                <span>{post.comment_count || 0} 评论</span>
              </span>
              <button className="flex items-center space-x-1 hover:bg-gray-100 px-3 py-2 rounded transition-colors">
                <Share2 className="h-5 w-5" />
                <span>分享</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Comments Section */}
      <div className="mt-6 bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-bold text-gray-900 mb-4">评论</h3>

        <div className="mb-6">
          <textarea
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
            rows={3}
            placeholder="写下你的评论..."
          ></textarea>
          <div className="mt-2 flex justify-end">
            <button className="bg-primary hover:bg-primary-light text-white px-4 py-2 rounded-lg font-medium transition-colors">
              发表评论
            </button>
          </div>
        </div>

        <div className="text-center text-gray-400 py-8">
          <p>暂无评论</p>
        </div>
      </div>
    </div>
  );
}
