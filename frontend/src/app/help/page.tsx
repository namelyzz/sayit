import { Mail, MessageCircleQuestion } from "lucide-react";
import PageShell from "@/components/ui/PageShell";

const faqs = [
  {
    question: "如何注册账号？",
    answer: "点击页面右上角的注册入口，填写用户名和密码即可创建账号。",
  },
  {
    question: "如何发布帖子？",
    answer: "登录后点击发布按钮，选择社区、填写标题和正文，然后提交。",
  },
  {
    question: "如何加入社区？",
    answer: "在侧边栏或社区页面找到感兴趣的社区，进入后点击关注。",
  },
  {
    question: "投票规则是什么？",
    answer: "每个帖子支持赞同或反对，用来帮助社区识别更有价值的内容。",
  },
];

export default function HelpPage() {
  return (
    <PageShell className="max-w-3xl">
      <section className="rounded-lg border border-border bg-surface p-6 shadow-sm">
        <p className="text-sm font-medium text-primary">帮助与反馈</p>
        <h1 className="mt-2 text-3xl font-bold text-foreground">遇到问题，我们一起把它理顺</h1>
        <p className="mt-4 text-base leading-8 text-muted-strong">
          如果你在使用 SayIt 时遇到问题，或者对社区体验有建议，可以通过下面的邮箱联系。
        </p>

        <div className="mt-6 flex items-center gap-3 rounded-lg border border-border bg-surface-soft p-4">
          <Mail className="h-5 w-5 text-primary" />
          <a href="mailto:905390065@qq.com" className="font-medium text-primary hover:text-primary-dark">
            905390065@qq.com
          </a>
        </div>
      </section>

      <section className="rounded-lg border border-border bg-surface p-6 shadow-sm">
        <div className="flex items-center gap-2">
          <MessageCircleQuestion className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-semibold text-foreground">常见问题</h2>
        </div>

        <div className="mt-5 divide-y divide-border">
          {faqs.map((item) => (
            <div key={item.question} className="py-4 first:pt-0 last:pb-0">
              <h3 className="font-medium text-foreground">{item.question}</h3>
              <p className="mt-2 text-sm leading-6 text-muted">{item.answer}</p>
            </div>
          ))}
        </div>
      </section>
    </PageShell>
  );
}
