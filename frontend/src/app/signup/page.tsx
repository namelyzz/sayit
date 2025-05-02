"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { UserPlus } from "lucide-react";
import Button from "@/components/ui/Button";
import { FieldHint, FieldLabel, Input } from "@/components/ui/Field";
import { useAuth } from "@/context/AuthContext";
import { getErrorMessage } from "@/lib/utils";

export default function SignupPage() {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [rePassword, setRePassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();
  const { signup } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    if (password !== rePassword) {
      setError("两次输入的密码不一致。");
      setLoading(false);
      return;
    }

    try {
      await signup(username, password, rePassword);
      router.push("/login");
    } catch (err) {
      setError(getErrorMessage(err, "注册失败，请稍后再试。"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto flex min-h-[calc(100vh-10rem)] w-full max-w-5xl items-center justify-center">
      <div className="grid w-full overflow-hidden rounded-lg border border-border bg-surface shadow-sm md:grid-cols-[1.1fr_0.9fr]">
        <section className="p-6 sm:p-8">
          <p className="text-sm font-medium text-primary">注册 SayIt</p>
          <h1 className="mt-1 text-2xl font-bold text-foreground">创建你的社区身份</h1>
          <p className="mt-2 text-sm text-muted">
            已有账号？{" "}
            <Link href="/login" className="font-medium text-primary hover:text-primary-dark">
              立即登录
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
                placeholder="给自己取一个好记的名字"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>

            <div>
              <FieldLabel htmlFor="email">邮箱</FieldLabel>
              <Input
                id="email"
                name="email"
                type="email"
                placeholder="可选"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
              <FieldHint>当前后端未使用邮箱字段，这里仅保留为未来资料扩展。</FieldHint>
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

            <div>
              <FieldLabel htmlFor="re-password">确认密码</FieldLabel>
              <Input
                id="re-password"
                name="re-password"
                type="password"
                required
                placeholder="再输入一次密码"
                value={rePassword}
                onChange={(e) => setRePassword(e.target.value)}
              />
            </div>

            <Button type="submit" disabled={loading} className="w-full">
              <UserPlus className="h-4 w-4" />
              {loading ? "注册中..." : "注册"}
            </Button>
          </form>
        </section>

        <section className="hidden bg-surface-soft p-8 md:block">
          <div className="flex h-full flex-col justify-between">
            <div>
              <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-primary text-lg font-bold text-white">
                S
              </div>
              <h2 className="mt-8 text-3xl font-bold leading-tight text-foreground">找到社区，也让社区找到你。</h2>
              <p className="mt-4 text-sm leading-6 text-muted-strong">
                用一个稳定的账号保存关注、发布和互动记录。
              </p>
            </div>
            <p className="text-sm text-muted">清爽社区，从第一条认真发言开始。</p>
          </div>
        </section>
      </div>
    </div>
  );
}
