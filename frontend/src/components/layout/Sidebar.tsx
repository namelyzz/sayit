"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { Home, TrendingUp, Shuffle, Heart } from "lucide-react";
import { apiClient } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";

interface HotCommunity {
  community_id: string;
  community_name: string;
}

interface RandomCommunity {
  community_id: string;
  name: string;
}

interface FollowedCommunity {
  community_id: string;
  name: string;
}

const FOLLOW_REFRESH_EVENT = "follow-refresh";

export default function Sidebar() {
  const { user, loading: authLoading } = useAuth();

  const [hotCommunities, setHotCommunities] = useState<HotCommunity[]>([]);
  const [randomCommunities, setRandomCommunities] = useState<RandomCommunity[]>([]);
  const [followedCommunities, setFollowedCommunities] = useState<FollowedCommunity[]>([]);
  const [hotLoading, setHotLoading] = useState(true);
  const [randomLoading, setRandomLoading] = useState(true);
  const [followedLoading, setFollowedLoading] = useState(true);

  useEffect(() => {
    const fetchHotCommunities = async () => {
      try {
        const response = await apiClient.getHotCommunities(5);
        setHotCommunities(response.data ?? []);
      } catch (error) {
        console.error("Failed to fetch hot communities:", error);
        setHotCommunities([]);
      } finally {
        setHotLoading(false);
      }
    };

    fetchHotCommunities();
  }, []);

  const fetchRandomCommunities = async () => {
    setRandomLoading(true);
    try {
      const response = await apiClient.getRandomCommunities(5);
      setRandomCommunities(response.data ?? []);
    } catch (error) {
      console.error("Failed to fetch random communities:", error);
      setRandomCommunities([]);
    } finally {
      setRandomLoading(false);
    }
  };

  useEffect(() => {
    fetchRandomCommunities();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const fetchFollowedCommunities = async () => {
    if (!user) {
      setFollowedCommunities([]);
      setFollowedLoading(false);
      return;
    }

    setFollowedLoading(true);
    try {
      const response = await apiClient.getFollowedCommunities();
      setFollowedCommunities(response.data ?? []);
    } catch (error: any) {
      console.error("Failed to fetch followed communities:", error);
      // apiClient 会自动清除无效 token，这里只清空列表
      setFollowedCommunities([]);
    } finally {
      setFollowedLoading(false);
    }
  };

  useEffect(() => {
    if (!user) {
      setFollowedCommunities([]);
      setFollowedLoading(false);
      return;
    }

    fetchFollowedCommunities();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  // 监听关注刷新事件
  useEffect(() => {
    const handler = () => {
      fetchFollowedCommunities();
    };
    window.addEventListener(FOLLOW_REFRESH_EVENT, handler);
    return () => {
      window.removeEventListener(FOLLOW_REFRESH_EVENT, handler);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  return (
    <aside className="w-64 bg-white border-r border-gray-200 h-[calc(100vh-4rem)] sticky top-16 overflow-y-auto hidden lg:block">
      <div className="p-4">
        {/* Navigation Links */}
        <nav className="space-y-1 mb-6">
          <Link
            href="/"
            className="flex items-center space-x-3 px-3 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <Home className="h-5 w-5" />
            <span>首页</span>
          </Link>
        </nav>

        {/* Divider */}
        <div className="border-t border-gray-200 my-4"></div>

        {/* Followed Communities Section (only for logged-in users) */}
        {user && (
          <div>
            <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
              已关注社区
            </h3>
            {followedLoading ? (
              <div className="space-y-2">
                {[...Array(3)].map((_, i) => (
                  <div
                    key={i}
                    className="h-10 bg-gray-100 rounded-lg animate-pulse"
                  ></div>
                ))}
              </div>
            ) : followedCommunities.length > 0 ? (
              <div className="space-y-1">
                {followedCommunities.map((community) => (
                  <Link
                    key={community.community_id}
                    href={`/community/${community.community_id}`}
                    className="flex items-center space-x-3 px-3 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                  >
                    <Heart className="h-4 w-4 text-primary fill-current" />
                    <span className="text-sm font-medium">
                      {community.name}
                    </span>
                  </Link>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-400 py-2">尚未关注任何社区</p>
            )}
            <div className="border-t border-gray-200 my-4"></div>
          </div>
        )}

        {/* Hot Communities Section */}
        <div>
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
            热门社区
          </h3>
          {hotLoading ? (
            <div className="space-y-2">
              {[...Array(3)].map((_, i) => (
                <div
                  key={i}
                  className="h-10 bg-gray-100 rounded-lg animate-pulse"
                ></div>
              ))}
            </div>
          ) : hotCommunities.length > 0 ? (
            <div className="space-y-1">
              {hotCommunities.map((community) => (
                <Link
                  key={community.community_id}
                  href={`/community/${community.community_id}`}
                  className="flex items-center space-x-3 px-3 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <TrendingUp className="h-4 w-4 text-accent" />
                  <span className="text-sm font-medium">
                    {community.community_name}
                  </span>
                </Link>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-400 py-2">暂无热门社区</p>
          )}
        </div>

        {/* Random Recommended Communities */}
        <div className="border-t border-gray-200 my-4 pt-4">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">
              随机推荐
            </h3>
            <button
              onClick={fetchRandomCommunities}
              disabled={randomLoading}
              className="text-xs text-primary hover:text-primary/80 disabled:opacity-50 flex items-center gap-1 transition-colors"
              title="刷新推荐"
            >
              <Shuffle className={`h-3 w-3 ${randomLoading ? "animate-spin" : ""}`} />
              刷新
            </button>
          </div>
          {randomLoading ? (
            <div className="space-y-2">
              {[...Array(3)].map((_, i) => (
                <div
                  key={i}
                  className="h-10 bg-gray-100 rounded-lg animate-pulse"
                ></div>
              ))}
            </div>
          ) : randomCommunities.length > 0 ? (
            <div className="space-y-1">
              {randomCommunities.map((community) => (
                <Link
                  key={community.community_id}
                  href={`/community/${community.community_id}`}
                  className="flex items-center space-x-3 px-3 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <Shuffle className="h-4 w-4 text-gray-400" />
                  <span className="text-sm font-medium">
                    {community.name}
                  </span>
                </Link>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-400 py-2">暂无推荐社区</p>
          )}
        </div>
      </div>
    </aside>
  );
}
