import { Code2, MessagesSquare, ShieldCheck } from "lucide-react";
import PageShell from "@/components/ui/PageShell";

export default function AboutPage() {
  return (
    <PageShell className="max-w-3xl">
      <section className="rounded-lg border border-border bg-surface p-6 shadow-sm">
        <p className="text-sm font-medium text-primary">关于 SayIt</p>
        <h1 className="mt-2 text-3xl font-bold text-foreground">一个用来认真聊天的中文社区</h1>
        <p className="mt-4 text-base leading-8 text-muted-strong">
          SayIt 是一个仿 Reddit 的社区论坛平台，支持用户注册登录、浏览社区、发布帖子、投票和参与讨论。
          它的目标不是制造噪音，而是给观点、问题和经验一个清楚可读的地方。
        </p>
      </section>

      <section className="grid gap-4 md:grid-cols-3">
        <div className="rounded-lg border border-border bg-surface p-5 shadow-sm">
          <MessagesSquare className="h-5 w-5 text-primary" />
          <h2 className="mt-4 font-semibold text-foreground">社区讨论</h2>
          <p className="mt-2 text-sm leading-6 text-muted">按社区组织内容，让不同主题有自己的场域。</p>
        </div>
        <div className="rounded-lg border border-border bg-surface p-5 shadow-sm">
          <ShieldCheck className="h-5 w-5 text-primary" />
          <h2 className="mt-4 font-semibold text-foreground">账号与投票</h2>
          <p className="mt-2 text-sm leading-6 text-muted">登录后可以发帖、关注社区，并用投票表达态度。</p>
        </div>
        <div className="rounded-lg border border-border bg-surface p-5 shadow-sm">
          <Code2 className="h-5 w-5 text-primary" />
          <h2 className="mt-4 font-semibold text-foreground">技术栈</h2>
          <p className="mt-2 text-sm leading-6 text-muted">后端 Go + Gin + MySQL + Redis，前端 Next.js + TypeScript。</p>
        </div>
      </section>

      <section className="rounded-lg border border-border bg-surface p-6 shadow-sm">
        <h2 className="text-xl font-semibold text-foreground">核心功能</h2>
        <ul className="mt-4 grid gap-3 text-sm text-muted-strong sm:grid-cols-2">
          <li>用户注册与登录</li>
          <li>社区浏览与关注</li>
          <li>帖子发布与详情阅读</li>
          <li>按时间或热度排序</li>
          <li>帖子投票机制</li>
          <li>响应式页面布局</li>
        </ul>
      </section>
    </PageShell>
  );
}
