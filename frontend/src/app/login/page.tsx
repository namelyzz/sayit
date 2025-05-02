"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { LogIn } from "lucide-react";
import Button from "@/components/ui/Button";
import { FieldLabel, Input } from "@/components/ui/Field";
import { useAuth } from "@/context/AuthContext";
import { getErrorMessage } from "@/lib/utils";

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [rememberMe, setRememberMe] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();
  const { login } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await login(username, password, rememberMe);
      router.push("/");
    } catch (err) {
      setError(getErrorMessage(err, "登录失败，请检查用户名和密码。"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto flex min-h-[calc(100vh-10rem)] w-full max-w-5xl items-center justify-center">
      <div className="grid w-full overflow-hidden rounded-lg border border-border bg-surface shadow-sm md:grid-cols-[0.9fr_1.1fr]">
        <section className="hidden bg-primary p-8 text-white md:block">
          <div className="flex h-full flex-col justify-between">
            <div>
              <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-white/15 text-lg font-bold">S</div>
              <h1 className="mt-8 text-3xl font-bold leading-tight">欢迎回来，继续把想法晒出来。</h1>
              <p className="mt-4 text-sm leading-6 text-teal-50">
                登录后可以发布帖子、关注社区，并参与投票和评论。
              </p>
            </div>
            <p className="text-sm text-teal-50">SayIt · 晒一个有意思的灵魂</p>
          </div>
        </section>

        <section className="p-6 sm:p-8">
          <p className="text-sm font-medium text-primary">登录 SayIt</p>
          <h2 className="mt-1 text-2xl font-bold text-foreground">进入你的社区首页</h2>
          <p className="mt-2 text-sm text-muted">
            还没有账号？{" "}
            <Link href="/signup" className="font-medium text-primary hover:text-primary-dark">
              立即注册
            </Link>
          </p>

          <form className="mt-8 space-y-5" onSubmit={handleSubmit}>
            {error ? (
              <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-danger">{error}</div>
            ) : null}

            <div>
              <FieldLabel htmlFor="username">用户名</FieldLabel>
              <Input
                id="username"
                name="username"
                type="text"
                required
                placeholder="输入用户名"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>

            <div>
              <FieldLabel htmlFor="password">密码</FieldLabel>
              <Input
                id="password"
                name="password"
                type="password"
                required
                placeholder="输入密码"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>

            <div className="flex items-center justify-between gap-4">
              <label htmlFor="remember-me" className="flex items-center gap-2 text-sm text-muted-strong">
                <input
                  id="remember-me"
                  name="remember-me"
                  type="checkbox"
                  checked={rememberMe}
                  onChange={(e) => setRememberMe(e.target.checked)}
                  className="h-4 w-4 rounded border-border text-primary focus:ring-primary"
                />
                记住登录状态
              </label>
              <a href="#" className="text-sm font-medium text-primary hover:text-primary-dark">
                忘记密码？
              </a>
            </div>

            <Button type="submit" disabled={loading} className="w-full">
              <LogIn className="h-4 w-4" />
              {loading ? "登录中..." : "登录"}
            </Button>
          </form>
        </section>
      </div>
    </div>
  );
}
