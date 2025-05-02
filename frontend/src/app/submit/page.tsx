"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { FileText, Send } from "lucide-react";
import Button from "@/components/ui/Button";
import PageShell from "@/components/ui/PageShell";
import { FieldHint, FieldLabel, Input, Select, Textarea } from "@/components/ui/Field";
import { apiClient, type CommunitySummary } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { getErrorMessage } from "@/lib/utils";

export default function SubmitPage() {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [communityId, setCommunityId] = useState("");
  const [communities, setCommunities] = useState<CommunitySummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();

  useEffect(() => {
    if (authLoading) return;

    if (!user) {
      router.push("/login");
      return;
    }

    const fetchCommunities = async () => {
      try {
        const response = await apiClient.getCommunities();
        setCommunities(response.data ?? []);
      } catch (err) {
        setError(getErrorMessage(err, "社区列表加载失败，请确认后端服务已启动。"));
      }
    };

    fetchCommunities();
  }, [user, authLoading, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    if (!title.trim() || !content.trim() || !communityId) {
      setError("请填写标题、内容并选择社区。");
      setLoading(false);
      return;
    }

    try {
      await apiClient.createPost(title.trim(), content.trim(), communityId);
      router.push("/");
    } catch (err) {
      setError(getErrorMessage(err, "发布失败，请稍后再试。"));
    } finally {
      setLoading(false);
    }
  };

  if (authLoading || !user) {
    return null;
  }

  return (
    <PageShell className="max-w-3xl">
      <section className="rounded-lg border border-border bg-surface p-5 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-teal-50 text-primary">
            <FileText className="h-5 w-5" />
          </div>
          <div>
            <p className="text-sm font-medium text-primary">发布新帖子</p>
            <h1 className="text-2xl font-bold text-foreground">把想法写清楚，也写得好读</h1>
          </div>
        </div>
      </section>

      <form onSubmit={handleSubmit} className="rounded-lg border border-border bg-surface p-5 shadow-sm">
        {error ? (
          <div className="mb-5 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">{error}</div>
        ) : null}

        <div className="space-y-5">
          <div>
            <FieldLabel htmlFor="community">发布到</FieldLabel>
            <Select
              id="community"
              value={communityId}
              onChange={(e) => setCommunityId(e.target.value)}
              required
              disabled={communities.length === 0}
            >
              <option value="">选择一个社区</option>
              {communities.map((community) => (
                <option key={community.community_id} value={community.community_id}>
                  {community.name}
                </option>
              ))}
            </Select>
            <FieldHint>{communities.length === 0 ? "暂无可用社区。" : "选择最贴近主题的社区，讨论会更容易被看见。"}</FieldHint>
          </div>

          <div>
            <FieldLabel htmlFor="title">标题</FieldLabel>
            <Input
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
              maxLength={128}
              placeholder="用一句话说清楚你想讨论什么"
            />
            <FieldHint>{title.length}/128</FieldHint>
          </div>

          <div>
            <FieldLabel htmlFor="content">正文</FieldLabel>
            <Textarea
              id="content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              required
              rows={14}
              maxLength={8192}
              placeholder="写下背景、观点、问题或你希望大家回应的方向..."
              className="resize-y"
            />
            <FieldHint>{content.length}/8192</FieldHint>
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <Button variant="outline" onClick={() => router.back()}>
            取消
          </Button>
          <Button type="submit" disabled={loading}>
            <Send className="h-4 w-4" />
            {loading ? "发布中..." : "发布"}
          </Button>
        </div>
      </form>
    </PageShell>
  );
}
