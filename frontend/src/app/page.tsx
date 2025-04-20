"use client";

import { useState, useEffect } from "react";
import PostCard from "@/components/ui/PostCard";
import { apiClient } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import Link from "next/link";

interface Post {
  post_id: string;
  title: string;
  summary: string;
  user_name: string;
  community_id: string;
  community_name: string;
  create_time: string;
  like_count: number;
  comment_count: number;
}

export default function Home() {
  const { user, loading: authLoading } = useAuth();
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [sortBy, setSortBy] = useState<"create_time" | "score">("create_time");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState("");
  const [needLogin, setNeedLogin] = useState(false);

  useEffect(() => {
    if (authLoading) return;

    const fetchPosts = async () => {
      setLoading(true);
      try {
        const response = await apiClient.getPosts({
          page,
          size: 10,
          sort_by: sortBy,
          order: "desc",
        });
        const postsData = Array.isArray(response.data) ? response.data : (response.data?.list ?? []);
        const totalData = Array.isArray(response.data) ? response.data.length : (response.data?.total ?? 0);
        setPosts(postsData);
        setTotal(totalData);
        setNeedLogin(false);
      } catch (error: any) {
        console.error("Failed to fetch posts:", error);
        if (!user) {
          setNeedLogin(true);
          setPosts([]);
        } else {
          setError("加载帖子失败，请检查后端服务是否已启动");
          setPosts([]);
        }
      } finally {
        setLoading(false);
      }
    };

    fetchPosts();
  }, [sortBy, page, user, authLoading]);

  return (
    <div>
      {/* Sort Controls */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
        <div className="flex items-center space-x-4">
          <span className="text-sm font-medium text-gray-700">排序：</span>
          <button
            onClick={() => { setSortBy("create_time"); setPage(1); }}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              sortBy === "create_time"
                ? "bg-primary text-white"
                : "bg-gray-100 text-gray-700 hover:bg-gray-200"
            }`}
          >
            最新
          </button>
          <button
            onClick={() => { setSortBy("score"); setPage(1); }}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              sortBy === "score"
                ? "bg-primary text-white"
                : "bg-gray-100 text-gray-700 hover:bg-gray-200"
            }`}
          >
            热门
          </button>
        </div>
      </div>

      {/* Posts List */}
      <div className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
            <span className="block sm:inline">{error}</span>
          </div>
        )}
        {needLogin && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
            <p className="text-gray-600 mb-4">登录后查看帖子内容</p>
            <Link
              href="/login"
              className="inline-block px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors"
            >
              去登录
            </Link>
          </div>
        )}
        {loading ? (
          [...Array(5)].map((_, i) => (
            <div
              key={i}
              className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 animate-pulse"
            >
              <div className="flex">
                <div className="w-10 bg-gray-200 rounded mr-4"></div>
                <div className="flex-1">
                  <div className="h-4 bg-gray-200 rounded w-1/4 mb-2"></div>
                  <div className="h-6 bg-gray-200 rounded w-3/4 mb-2"></div>
                  <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
                  <div className="h-4 bg-gray-200 rounded w-2/3"></div>
                </div>
              </div>
            </div>
          ))
        ) : posts && posts.length > 0 ? (
          posts.map((post) => <PostCard key={post.post_id} post={post} />)
        ) : !needLogin && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
            <p className="text-gray-500">暂无帖子</p>
          </div>
        )}
      </div>

      {/* Pagination */}
      {total > 10 && (
        <div className="mt-6 flex justify-center space-x-2">
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
            className="px-4 py-2 bg-white border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            上一页
          </button>
          <span className="px-4 py-2 text-sm text-gray-700">
            第 {page} 页
          </span>
          <button
            onClick={() => setPage(page + 1)}
            disabled={!posts || posts.length < 10}
            className="px-4 py-2 bg-white border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            下一页
          </button>
        </div>
      )}
    </div>
  );
}
