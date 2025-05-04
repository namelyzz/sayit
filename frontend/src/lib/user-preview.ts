export interface PreviewPerson {
  id: string;
  name: string;
  note: string;
}

export interface PreviewComment {
  id: string;
  postTitle: string;
  excerpt: string;
  createdAt: string;
}

export const DEFAULT_SIGNATURE = "把有意思的想法慢慢晒出来。";
export const PREVIEW_REGISTER_DATE = "2024-08-18T10:00:00+08:00";

export const previewFollowers: PreviewPerson[] = [
  { id: "f-101", name: "南风", note: "喜欢你写的社区观察。" },
  { id: "f-102", name: "阿宁", note: "经常来看看你的新帖子。" },
  { id: "f-103", name: "清和", note: "关注了你的创作分享。" },
];

export const previewFollowing: PreviewPerson[] = [
  { id: "g-201", name: "北岛", note: "经常更新技术随笔。" },
  { id: "g-202", name: "晚舟", note: "擅长写很有温度的生活记录。" },
];

export const previewComments: PreviewComment[] = [
  {
    id: "c-301",
    postTitle: "如果只保留一个日常习惯，你会留下什么？",
    excerpt: "我大概率会留下散步。它不像效率工具，更像帮人把脑子里的噪音慢慢放掉。",
    createdAt: "2026-05-10T20:15:00+08:00",
  },
  {
    id: "c-302",
    postTitle: "中文社区产品应该更像论坛还是轻社交？",
    excerpt: "我更倾向于先把讨论质量做稳，再慢慢叠加关系链，不然首页很容易被节奏带偏。",
    createdAt: "2026-05-08T09:40:00+08:00",
  },
];
