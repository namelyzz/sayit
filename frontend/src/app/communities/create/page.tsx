"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Globe, Send } from "lucide-react";
import Button from "@/components/ui/Button";
import PageShell from "@/components/ui/PageShell";
import { FieldHint, FieldLabel, Input, Textarea } from "@/components/ui/Field";
import { apiClient } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { getErrorMessage } from "@/lib/utils";

export default function CreateCommunityPage() {
  const [name, setName] = useState("");
  const [introduction, setIntroduction] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();

  useEffect(() => {
    if (authLoading) return;

    if (!user) {
      router.push("/login");
    }
  }, [user, authLoading, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    if (!name.trim() || !introduction.trim()) {
      setError("请填写社区名称和简介。");
      setLoading(false);
      return;
    }

    if (name.trim().length < 2) {
      setError("社区名称至少需要2个字符。");
      setLoading(false);
      return;
    }

    if (introduction.trim().length < 2) {
      setError("社区简介至少需要2个字符。");
      setLoading(false);
      return;
    }

    try {
      const response = await apiClient.createCommunity(name.trim(), introduction.trim());
      setSuccess(true);
      setTimeout(() => {
        router.push(`/community/${response.data.community_id}`);
      }, 1500);
    } catch (err) {
      setError(getErrorMessage(err, "创建失败，请稍后再试。"));
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
            <Globe className="h-5 w-5" />
          </div>
          <div>
            <p className="text-sm font-medium text-primary">申请创建社区</p>
            <h1 className="text-2xl font-bold text-foreground">创建一个属于你的社区</h1>
          </div>
        </div>
      </section>

      <form onSubmit={handleSubmit} className="rounded-lg border border-border bg-surface p-5 shadow-sm">
        {error ? (
          <div className="mb-5 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">{error}</div>
        ) : null}

        {success ? (
          <div className="mb-5 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
            社区创建成功！正在跳转到社区页面...
          </div>
        ) : null}

        <div className="space-y-5">
          <div>
            <FieldLabel htmlFor="name">社区名称</FieldLabel>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              minLength={2}
              maxLength={128}
              placeholder="给你的社区起一个独特的名字"
            />
            <FieldHint>{name.length}/128</FieldHint>
          </div>

          <div>
            <FieldLabel htmlFor="introduction">社区简介</FieldLabel>
            <Textarea
              id="introduction"
              value={introduction}
              onChange={(e) => setIntroduction(e.target.value)}
              required
              rows={4}
              maxLength={256}
              placeholder="简要介绍社区的主题和讨论方向..."
              className="resize-y"
            />
            <FieldHint>{introduction.length}/256</FieldHint>
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <Button variant="outline" onClick={() => router.back()}>
            取消
          </Button>
          <Button type="submit" disabled={loading || success}>
            <Send className="h-4 w-4" />
            {loading ? "创建中..." : success ? "创建成功" : "创建社区"}
          </Button>
        </div>
      </form>
    </PageShell>
  );
}
