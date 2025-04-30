"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import PostCard from "@/components/ui/PostCard";
import { apiClient } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { Users, Calendar, Heart } from "lucide-react";

interface Community {
  community_id: string;
  name: string;
  introduction: string;
  create_time: string;
}

interface Post {
  post_id: string;
  title: string;
  summary: string;
  user_name: string;
  community_name: string;
  create_time: string;
  like_count: number;
  comment_count: number;
}

export default function CommunityPage() {
  const params = useParams();
  const communityId = params.id as string;
  const { user, loading: authLoading } = useAuth();

  const [community, setCommunity] = useState<Community | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [isFollowed, setIsFollowed] = useState(false);
  const [loading, setLoading] = useState(true);
  const [followLoading, setFollowLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (authLoading) return;

    const fetchData = async () => {
      setLoading(true);
      try {
        const [communityResponse, postsResponse] = await Promise.all([
          apiClient.getCommunityDetail(communityId),
          apiClient.getPosts({ community_id: communityId, size: 20 }),
        ]);

        setCommunity(communityResponse.data);
        const postsData = Array.isArray(postsResponse.data) ? postsResponse.data : (postsResponse.data?.list ?? []);
        setPosts(postsData);

        // 检查关注状态（仅登录用户）
        if (user) {
          try {
            const followResponse = await apiClient.isFollowedCommunity(communityId);
            setIsFollowed(followResponse.data?.is_followed ?? false);
          } catch {
            setIsFollowed(false);
          }
        }
      } catch (error: any) {
        console.error("Failed to fetch community data:", error);
        const msg = error?.message || "";
        if (msg.includes("token") || msg.includes("登录")) {
          setError("登录后查看完整内容");
        } else {
          setError("加载社区数据失败，请检查后端服务是否已启动");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [communityId, authLoading]);

  const handleFollowToggle = async () => {
    if (!user) {
      window.location.href = "/login";
      return;
    }

    setFollowLoading(true);
    try {
      if (isFollowed) {
        await apiClient.unfollowCommunity(communityId);
        setIsFollowed(false);
      } else {
        await apiClient.followCommunity(communityId);
        setIsFollowed(true);
      }
      // 通知左侧栏刷新已关注社区列表
      window.dispatchEvent(new CustomEvent("follow-refresh"));
    } catch (error: any) {
      console.error("Follow toggle failed:", error);
      // Token 无效时 apiClient 会自动清除，这里提示用户
      if (error?.message?.includes("token") || error?.message?.includes("登录")) {
        window.location.href = "/login";
      } else {
        alert("操作失败，请重试");
      }
    } finally {
      setFollowLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 animate-pulse mb-6">
          <div className="h-8 bg-gray-200 rounded w-1/3 mb-4"></div>
          <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-2/3"></div>
        </div>
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 animate-pulse">
              <div className="h-6 bg-gray-200 rounded w-3/4 mb-2"></div>
              <div className="h-4 bg-gray-200 rounded w-full"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error && !community) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
          <p className="text-gray-500">{error}</p>
          {!user && (
            <Link href="/login" className="text-primary hover:underline mt-4 inline-block">
              去登录
            </Link>
          )}
          <Link href="/" className="text-primary hover:underline mt-4 inline-block ml-4">
            返回首页
          </Link>
        </div>
      </div>
    );
  }

  if (!community) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
          <p className="text-gray-500">社区不存在</p>
          <Link href="/" className="text-primary hover:underline mt-4 inline-block">
            返回首页
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      {/* Community Header */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">
              {community.name}
            </h1>
            <p className="text-gray-600 mb-4">{community.introduction}</p>
            <div className="flex items-center space-x-4 text-sm text-gray-500">
              <span className="flex items-center">
                <Users className="h-4 w-4 mr-1" />
                社区成员
              </span>
              <span className="flex items-center">
                <Calendar className="h-4 w-4 mr-1" />
                创建于 {community.create_time}
              </span>
            </div>
          </div>
          <div className="flex flex-col items-end space-y-2">
            {/* Follow Button */}
            {user && (
              <button
                onClick={handleFollowToggle}
                disabled={followLoading}
                className={`flex items-center gap-1.5 px-4 py-2 rounded-lg font-medium transition-all ${
                  isFollowed
                    ? "bg-primary text-white"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                } disabled:opacity-60`}
              >
                <Heart
                  className={`h-4 w-4 ${
                    isFollowed ? "fill-current" : ""
                  }`}
                />
                <span>{isFollowed ? "已关注" : "关注"}</span>
              </button>
            )}
            {/* Create Post Button */}
            {user ? (
              <Link
                href="/submit"
                className="bg-primary hover:bg-primary-light text-white px-4 py-2 rounded-lg font-medium transition-colors"
              >
                发帖
              </Link>
            ) : (
              <Link
                href="/login"
                className="bg-gray-100 hover:bg-gray-200 text-gray-600 px-4 py-2 rounded-lg font-medium transition-colors"
              >
                登录后发帖
              </Link>
            )}
          </div>
        </div>
      </div>

      {/* Posts List */}
      <div className="space-y-4">
        {posts.length > 0 ? (
          posts.map((post) => <PostCard key={post.post_id} post={post} />)
        ) : (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
            <p className="text-gray-500">暂无帖子</p>
          </div>
        )}
      </div>
    </div>
  );
}
