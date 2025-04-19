"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { Home, Users, Star } from "lucide-react";
import { apiClient } from "@/lib/api";

interface Community {
  community_id: string;
  name: string;
}

export default function Sidebar() {
  const [communities, setCommunities] = useState<Community[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchCommunities = async () => {
      try {
        const response = await apiClient.getCommunities();
        setCommunities(response.data ?? []);
      } catch (error) {
        console.error("Failed to fetch communities:", error);
        setCommunities([
          { community_id: "1", name: "GolangStudy" },
          { community_id: "2", name: "KamenRiderFaiz" },
          { community_id: "3", name: "A_Stock" },
          { community_id: "4", name: "EnglishSpeaking" },
          { community_id: "5", name: "WoodworkingDIY" },
          { community_id: "6", name: "AnimeLovers" },
          { community_id: "7", name: "HomeCook" },
          { community_id: "8", name: "FitnessBeginner" },
          { community_id: "9", name: "DigitalNomad" },
          { community_id: "10", name: "PlantParents" },
        ]);
      } finally {
        setLoading(false);
      }
    };

    fetchCommunities();
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

        {/* Communities Section */}
        <div>
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
            社区
          </h3>
          
          {loading ? (
            <div className="space-y-2">
              {[...Array(5)].map((_, i) => (
                <div
                  key={i}
                  className="h-10 bg-gray-100 rounded-lg animate-pulse"
                ></div>
              ))}
            </div>
          ) : (
            <div className="space-y-1">
              {communities.map((community) => (
                <Link
                  key={community.community_id}
                  href={`/community/${community.community_id}`}
                  className="flex items-center space-x-3 px-3 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <div className="w-8 h-8 bg-primary rounded-full flex items-center justify-center">
                    <Users className="h-4 w-4 text-white" />
                  </div>
                  <span className="text-sm font-medium">
                    {community.name}
                  </span>
                </Link>
              ))}
            </div>
          )}
        </div>

        {/* Popular Communities */}
        <div className="border-t border-gray-200 my-4 pt-4">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
            热门社区
          </h3>
          <div className="space-y-1">
            {communities.slice(0, 5).map((community) => (
              <Link
                key={community.community_id}
                href={`/community/${community.community_id}`}
                className="flex items-center space-x-3 px-3 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <Star className="h-4 w-4 text-accent" />
                <span className="text-sm">
                  {community.name}
                </span>
              </Link>
            ))}
          </div>
        </div>
      </div>
    </aside>
  );
}
