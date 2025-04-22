"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { Home, TrendingUp, Shuffle } from "lucide-react";
import { apiClient } from "@/lib/api";

interface HotCommunity {
  community_id: string;
  community_name: string;
}

export default function Sidebar() {
  const [hotCommunities, setHotCommunities] = useState<HotCommunity[]>([]);
  const [hotLoading, setHotLoading] = useState(true);

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

        {/* Random Recommended Communities (Placeholder) */}
        <div className="border-t border-gray-200 my-4 pt-4">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
            随机推荐
          </h3>
          <div className="space-y-1">
            <Link
              href="#"
              className="flex items-center space-x-3 px-3 py-2 text-gray-400 hover:bg-gray-50 rounded-lg transition-colors cursor-not-allowed"
            >
              <Shuffle className="h-4 w-4 text-gray-300" />
              <span className="text-sm">敬请期待</span>
            </Link>
          </div>
        </div>
      </div>
    </aside>
  );
}
