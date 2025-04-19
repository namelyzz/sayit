"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { apiClient } from "@/lib/api";

interface Community {
  community_id: string;
  name: string;
}

export default function SubmitPage() {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [communityId, setCommunityId] = useState("");
  const [communities, setCommunities] = useState<Community[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();
  const { user } = useAuth();

  useEffect(() => {
    if (!user) {
      router.push("/login");
      return;
    }

    const fetchCommunities = async () => {
      try {
        const response = await apiClient.getCommunities();
        setCommunities(response.data);
      } catch (error) {
        console.error("Failed to fetch communities:", error);
        setError("加载社区列表失败，请检查后端服务是否已启动");
      }
    };

    fetchCommunities();
  }, [user, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    if (!title.trim() || !content.trim() || !communityId) {
      setError("请填写所有必填字段");
      setLoading(false);
      return;
    }

    try {
      await apiClient.createPost(title, content, communityId);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "发帖失败，请稍后再试");
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return null;
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">发布新帖子</h1>
      
      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
            <span className="block sm:inline">{error}</span>
          </div>
        )}

        <div>
          <label htmlFor="community" className="block text-sm font-medium text-gray-700 mb-2">
            选择社区
          </label>
          <select
            id="community"
            value={communityId}
            onChange={(e) => setCommunityId(e.target.value)}
            required
            disabled={communities.length === 0}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
          >
            <option value="">请选择社区</option>
            {communities.map((community) => (
              <option key={community.community_id} value={community.community_id}>
                {community.name}
              </option>
            ))}
          </select>
          {communities.length === 0 && <p className="text-sm text-gray-500 mt-1">暂无可用社区</p>}
        </div>

        <div>
          <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
            标题
          </label>
          <input
            id="title"
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            maxLength={128}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
            placeholder="请输入标题（最多128字）"
          />
        </div>

        <div>
          <label htmlFor="content" className="block text-sm font-medium text-gray-700 mb-2">
            内容
          </label>
          <textarea
            id="content"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            required
            rows={10}
            maxLength={8192}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
            placeholder="请输入内容（最多8192字）"
          ></textarea>
        </div>

        <div className="flex justify-end space-x-4">
          <button
            type="button"
            onClick={() => router.back()}
            className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
          >
            取消
          </button>
          <button
            type="submit"
            disabled={loading}
            className="px-6 py-2 bg-primary hover:bg-primary-light text-white rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? "发布中..." : "发布"}
          </button>
        </div>
      </form>
    </div>
  );
}
