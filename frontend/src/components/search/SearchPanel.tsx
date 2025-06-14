"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Search, X } from "lucide-react";
import { apiClient, SearchSuggestItem } from "@/lib/api";
import { cn } from "@/lib/utils";

interface SearchPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function SearchPanel({ isOpen, onClose }: SearchPanelProps) {
  const router = useRouter();
  const panelRef = useRef<HTMLDivElement>(null);
  const communityInputRef = useRef<HTMLInputElement>(null);
  const userInputRef = useRef<HTMLInputElement>(null);

  const [keyword, setKeyword] = useState("");
  const [communityName, setCommunityName] = useState("");
  const [userName, setUserName] = useState("");

  const [communitySuggestions, setCommunitySuggestions] = useState<SearchSuggestItem[]>([]);
  const [userSuggestions, setUserSuggestions] = useState<SearchSuggestItem[]>([]);
  const [showCommunitySuggestions, setShowCommunitySuggestions] = useState(false);
  const [showUserSuggestions, setShowUserSuggestions] = useState(false);

  const communityTimerRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const userTimerRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  const fetchCommunitySuggestions = useCallback(async (q: string) => {
    if (q.length < 1) {
      setCommunitySuggestions([]);
      return;
    }
    try {
      const res = await apiClient.getSearchSuggest(q, "community", 8);
      setCommunitySuggestions(res.data);
    } catch {
      setCommunitySuggestions([]);
    }
  }, []);

  const fetchUserSuggestions = useCallback(async (q: string) => {
    if (q.length < 1) {
      setUserSuggestions([]);
      return;
    }
    try {
      const res = await apiClient.getSearchSuggest(q, "user", 8);
      setUserSuggestions(res.data);
    } catch {
      setUserSuggestions([]);
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
    clearTimeout(userTimerRef.current);
    if (userName && showUserSuggestions) {
      userTimerRef.current = setTimeout(() => {
        fetchUserSuggestions(userName);
      }, 300);
    } else {
      setUserSuggestions([]);
    }
    return () => clearTimeout(userTimerRef.current);
  }, [userName, showUserSuggestions, fetchUserSuggestions]);

  useEffect(() => {
    if (!isOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        setShowCommunitySuggestions(false);
        setShowUserSuggestions(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) {
      setKeyword("");
      setCommunityName("");
      setUserName("");
      setCommunitySuggestions([]);
      setUserSuggestions([]);
      setShowCommunitySuggestions(false);
      setShowUserSuggestions(false);
    }
  }, [isOpen]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
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
    userInputRef.current?.focus();
  };

  const selectUser = (item: SearchSuggestItem) => {
    setUserName(item.name);
    setShowUserSuggestions(false);
  };

  if (!isOpen) return null;

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

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              帖子关键字
            </label>
            <input
              type="text"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              placeholder="搜索帖子标题..."
              className="h-10 w-full rounded-lg border border-slate-200 px-3 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>

          <div className="relative">
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              社区名称
            </label>
            <input
              ref={communityInputRef}
              type="text"
              value={communityName}
              onChange={(e) => setCommunityName(e.target.value)}
              onFocus={() => setShowCommunitySuggestions(true)}
              placeholder="模糊匹配社区名..."
              className="h-10 w-full rounded-lg border border-slate-200 px-3 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
            {showCommunitySuggestions && communitySuggestions.length > 0 && (
              <div className="absolute left-0 right-0 top-full z-20 mt-1 max-h-48 overflow-y-auto rounded-lg border border-slate-200 bg-white py-1 shadow-lg">
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
            )}
          </div>

          <div className="relative">
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              用户名
            </label>
            <input
              ref={userInputRef}
              type="text"
              value={userName}
              onChange={(e) => setUserName(e.target.value)}
              onFocus={() => setShowUserSuggestions(true)}
              placeholder="模糊匹配用户名..."
              className="h-10 w-full rounded-lg border border-slate-200 px-3 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
            {showUserSuggestions && userSuggestions.length > 0 && (
              <div className="absolute left-0 right-0 top-full z-20 mt-1 max-h-48 overflow-y-auto rounded-lg border border-slate-200 bg-white py-1 shadow-lg">
                {userSuggestions.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => selectUser(item)}
                    className="w-full px-3 py-2 text-left text-sm hover:bg-slate-50"
                  >
                    {item.name}
                  </button>
                ))}
              </div>
            )}
          </div>

          <button
            type="submit"
            className="flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-primary text-sm font-medium text-white transition hover:bg-primary-dark"
          >
            <Search className="h-4 w-4" />
            搜索
          </button>
        </form>
      </div>
    </div>
  );
}
