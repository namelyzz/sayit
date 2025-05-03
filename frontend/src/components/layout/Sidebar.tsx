"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { Flame, Heart, Home, RefreshCw, Sparkles } from "lucide-react";
import { apiClient, type CommunitySummary, type HotCommunity } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { cn, getErrorMessage } from "@/lib/utils";

const FOLLOW_REFRESH_EVENT = "follow-refresh";

function CommunityLink({
  href,
  label,
  icon,
  tone = "neutral",
}: {
  href: string;
  label: string;
  icon: React.ReactNode;
  tone?: "neutral" | "accent" | "primary";
}) {
  return (
    <Link
      href={href}
      className="group flex min-h-10 items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-slate-900/85 transition hover:bg-white/45 hover:text-slate-950"
    >
      <span
        className={cn(
          "flex h-7 w-7 shrink-0 items-center justify-center rounded-lg shadow-sm",
          tone === "primary" && "bg-white/80 text-primary",
          tone === "accent" && "bg-white/70 text-accent",
          tone === "neutral" && "bg-white/65 text-slate-700"
        )}
      >
        {icon}
      </span>
      <span className="truncate">{label}</span>
    </Link>
  );
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return <h3 className="px-3 text-xs font-semibold uppercase tracking-wide text-slate-900/70">{children}</h3>;
}

function SectionBlock({ children }: { children: React.ReactNode }) {
  return <section className="rounded-xl bg-white/24 p-3">{children}</section>;
}

export default function Sidebar() {
  const { user } = useAuth();
  const [hotCommunities, setHotCommunities] = useState<HotCommunity[]>([]);
  const [randomCommunities, setRandomCommunities] = useState<CommunitySummary[]>([]);
  const [followedCommunities, setFollowedCommunities] = useState<CommunitySummary[]>([]);
  const [hotLoading, setHotLoading] = useState(true);
  const [randomLoading, setRandomLoading] = useState(true);
  const [followedLoading, setFollowedLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchHotCommunities = async () => {
      try {
        const response = await apiClient.getHotCommunities(6);
        setHotCommunities(response.data ?? []);
      } catch (err) {
        setError(getErrorMessage(err, "社区列表暂时不可用"));
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
    } catch (err) {
      setError(getErrorMessage(err, "推荐社区暂时不可用"));
      setRandomCommunities([]);
    } finally {
      setRandomLoading(false);
    }
  };

  useEffect(() => {
    fetchRandomCommunities();
  }, []);

  const fetchFollowedCommunities = useCallback(async () => {
    if (!user) {
      setFollowedCommunities([]);
      setFollowedLoading(false);
      return;
    }

    setFollowedLoading(true);
    try {
      const response = await apiClient.getFollowedCommunities();
      setFollowedCommunities(response.data ?? []);
    } catch {
      setFollowedCommunities([]);
    } finally {
      setFollowedLoading(false);
    }
  }, [user]);

  useEffect(() => {
    fetchFollowedCommunities();
  }, [fetchFollowedCommunities]);

  useEffect(() => {
    const handler = () => {
      fetchFollowedCommunities();
    };
    window.addEventListener(FOLLOW_REFRESH_EVENT, handler);
    return () => window.removeEventListener(FOLLOW_REFRESH_EVENT, handler);
  }, [fetchFollowedCommunities]);

  return (
    <aside className="hidden min-h-[calc(100vh-4rem)] overflow-y-auto px-4 pb-4 pt-2 lg:block scrollbar-thin">
      <div className="sticky top-2 space-y-3">
        <SectionBlock>
          <nav className="space-y-1">
            <CommunityLink href="/" label="首页" icon={<Home className="h-4 w-4" />} tone="primary" />
          </nav>
        </SectionBlock>

        {user ? (
          <SectionBlock>
            <div className="space-y-3">
              <SectionTitle>我的社区</SectionTitle>
              {followedLoading ? (
                <div className="space-y-2 px-3">
                  <div className="h-9 animate-pulse rounded-lg bg-white/40" />
                  <div className="h-9 animate-pulse rounded-lg bg-white/40" />
                </div>
              ) : followedCommunities.length > 0 ? (
                <div className="space-y-1">
                  {followedCommunities.map((community) => (
                    <CommunityLink
                      key={community.community_id}
                      href={`/community/${community.community_id}`}
                      label={community.name}
                      icon={<Heart className="h-4 w-4 fill-current" />}
                      tone="primary"
                    />
                  ))}
                </div>
              ) : (
                <p className="px-3 text-sm leading-6 text-slate-900/70">关注社区后，会在这里快速进入。</p>
              )}
            </div>
          </SectionBlock>
        ) : null}

        <SectionBlock>
          <div className="space-y-3">
            <SectionTitle>热门社区</SectionTitle>
            {hotLoading ? (
              <div className="space-y-2 px-3">
                <div className="h-9 animate-pulse rounded-lg bg-white/40" />
                <div className="h-9 animate-pulse rounded-lg bg-white/40" />
                <div className="h-9 animate-pulse rounded-lg bg-white/40" />
              </div>
            ) : hotCommunities.length > 0 ? (
              <div className="space-y-1">
                {hotCommunities.map((community) => (
                  <CommunityLink
                    key={community.community_id}
                    href={`/community/${community.community_id}`}
                    label={community.community_name}
                    icon={<Flame className="h-4 w-4" />}
                    tone="accent"
                  />
                ))}
              </div>
            ) : (
              <p className="px-3 text-sm leading-6 text-slate-900/70">暂无热门社区。</p>
            )}
          </div>
        </SectionBlock>

        <SectionBlock>
          <div className="space-y-3">
            <div className="flex items-center justify-between px-3">
              <SectionTitle>随便逛逛</SectionTitle>
              <button
                onClick={fetchRandomCommunities}
                disabled={randomLoading}
                className="inline-flex h-7 items-center gap-1 rounded-md px-2 text-xs font-medium text-primary transition hover:bg-white/40 disabled:opacity-50"
              >
                <RefreshCw className={cn("h-3.5 w-3.5", randomLoading && "animate-spin")} />
                换一批
              </button>
            </div>
            {randomCommunities.length > 0 ? (
              <div className="space-y-1">
                {randomCommunities.map((community) => (
                  <CommunityLink
                    key={community.community_id}
                    href={`/community/${community.community_id}`}
                    label={community.name}
                    icon={<Sparkles className="h-4 w-4" />}
                  />
                ))}
              </div>
            ) : (
              <p className="px-3 text-sm leading-6 text-slate-900/70">
                {randomLoading ? "正在寻找社区..." : "暂无推荐社区。"}
              </p>
            )}
          </div>
        </SectionBlock>

        {error ? <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-danger">{error}</p> : null}
      </div>
    </aside>
  );
}
