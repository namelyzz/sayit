"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { ChevronUp, Search, X } from "lucide-react";
import { apiClient, SearchSuggestItem } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";

interface SearchPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function SearchPanel({ isOpen, onClose }: SearchPanelProps) {
  const router = useRouter();
  const { user } = useAuth();
  const panelRef = useRef<HTMLDivElement>(null);
  const communityInputRef = useRef<HTMLInputElement>(null);
  const suggestionsRef = useRef<HTMLDivElement>(null);

  const [keyword, setKeyword] = useState("");
  const [communityName, setCommunityName] = useState("");
  const [userName, setUserName] = useState("");

  const [communitySuggestions, setCommunitySuggestions] = useState<SearchSuggestItem[]>([]);
  const [showCommunitySuggestions, setShowCommunitySuggestions] = useState(false);

  const communityTimerRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  const hasAnyInput = keyword.trim() || communityName.trim() || userName.trim();

  const fetchCommunitySuggestions = useCallback(async (q: string) => {
    if (q.length < 1) {
      setCommunitySuggestions([]);
      return;
    }
    try {
      const res = await apiClient.getSearchSuggest(q, "community", 5);
      setCommunitySuggestions(res.data);
    } catch {
      setCommunitySuggestions([]);
    }
  }, []);

  useEffect(() => {
    clearTimeout(communityTimerRef.current);
    if (communityName && showCommunitySuggestions) {
      communityTimerRef.current = setTimeout(() => {
        fetchCommunitySuggestions(communityName);
      }, 300);
    } else {
      setCommunitySuggestions([]);
    }
    return () => clearTimeout(communityTimerRef.current);
  }, [communityName, showCommunitySuggestions, fetchCommunitySuggestions]);

  useEffect(() => {
    if (!isOpen) {
      setKeyword("");
      setCommunityName("");
      setUserName("");
      setCommunitySuggestions([]);
      setShowCommunitySuggestions(false);
    }
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as Node;
      const clickedInsideInput = communityInputRef.current?.contains(target);
      const clickedInsideSuggestions = suggestionsRef.current?.contains(target);

      if (!clickedInsideInput && !clickedInsideSuggestions) {
        setShowCommunitySuggestions(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen]);

  const handleSubmit = () => {
    if (!hasAnyInput) return;

    if (!user) {
      router.push("/login");
      onClose();
      return;
    }

    const params = new URLSearchParams();
    if (keyword.trim()) params.set("keyword", keyword.trim());
    if (communityName.trim()) params.set("community_name", communityName.trim());
    if (userName.trim()) params.set("user_name", userName.trim());
    const query = params.toString();
    router.push(`/search${query ? `?${query}` : ""}`);
    onClose();
  };

  const selectCommunity = (item: SearchSuggestItem) => {
    setCommunityName(item.name);
    setShowCommunitySuggestions(false);
  };

  const handleCommunityKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      setShowCommunitySuggestions(false);
      if (hasAnyInput) {
        handleSubmit();
      }
    }
    if (e.key === "Escape") {
      setShowCommunitySuggestions(false);
    }
  };

  if (!isOpen) return null;

  const showSuggestions = showCommunitySuggestions && communitySuggestions.length > 0;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[10vh]">
      <div className="fixed inset-0 bg-black/40 backdrop-blur-sm" onClick={onClose} />
      <div
        ref={panelRef}
        className="relative z-10 w-full max-w-lg rounded-xl bg-white p-6 shadow-2xl"
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-900">高级搜索</h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="space-y-4">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              帖子关键字
            </label>
            <input
              type="text"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && hasAnyInput) {
                  e.preventDefault();
                  handleSubmit();
                }
              }}
              placeholder="搜索帖子标题..."
              className="h-10 w-full rounded-lg border border-slate-200 px-3 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              社区名称
            </label>
            <input
              ref={communityInputRef}
              type="text"
              value={communityName}
              onChange={(e) => {
                setCommunityName(e.target.value);
                setShowCommunitySuggestions(true);
              }}
              onFocus={() => setShowCommunitySuggestions(true)}
              onKeyDown={handleCommunityKeyDown}
              placeholder="输入社区名前缀匹配..."
              className="h-10 w-full rounded-lg border border-slate-200 px-3 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>

          {showSuggestions && (
            <div
              ref={suggestionsRef}
              className="overflow-hidden rounded-lg border border-slate-200 bg-white"
            >
              <div className="flex items-center justify-between border-b border-slate-100 px-3 py-1.5">
                <span className="text-xs text-slate-400">选择社区</span>
                <button
                  type="button"
                  onClick={() => setShowCommunitySuggestions(false)}
                  className="rounded p-0.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
                >
                  <ChevronUp className="h-3.5 w-3.5" />
                </button>
              </div>
              <div className="max-h-[200px] overflow-y-auto py-1">
                {communitySuggestions.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => selectCommunity(item)}
                    className="w-full px-3 py-2 text-left text-sm hover:bg-slate-50"
                  >
                    {item.name}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              用户名
            </label>
            <input
              type="text"
              value={userName}
              onChange={(e) => setUserName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && hasAnyInput) {
                  e.preventDefault();
                  handleSubmit();
                }
              }}
              placeholder="输入用户名..."
              className="h-10 w-full rounded-lg border border-slate-200 px-3 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>

          <button
            type="button"
            onClick={handleSubmit}
            disabled={!hasAnyInput}
            className="flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-primary text-sm font-medium text-white transition hover:bg-primary-dark disabled:cursor-not-allowed disabled:opacity-50"
          >
            <Search className="h-4 w-4" />
            搜索
          </button>
        </div>

        {!user && (
          <p className="mt-3 text-center text-xs text-slate-500">
            搜索功能需要{" "}
            <button onClick={() => { router.push("/login"); onClose(); }} className="text-primary hover:underline">
              登录
            </button>{" "}
            后使用
          </p>
        )}
      </div>
    </div>
  );
}
