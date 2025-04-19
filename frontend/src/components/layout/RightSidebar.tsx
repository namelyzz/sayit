"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { User, Award } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

function CountdownTimer() {
  const [timeLeft, setTimeLeft] = useState("");

  useEffect(() => {
    const target = new Date();
    target.setDate(target.getDate() + 3);
    target.setHours(23, 59, 59, 0);

    const update = () => {
      const now = new Date();
      const diff = target.getTime() - now.getTime();
      if (diff <= 0) {
        setTimeLeft("已结束");
        return;
      }
      const days = Math.floor(diff / (1000 * 60 * 60 * 24));
      const hours = Math.floor((diff / (1000 * 60 * 60)) % 24);
      const mins = Math.floor((diff / (1000 * 60)) % 60);
      const secs = Math.floor((diff / 1000) % 60);
      setTimeLeft(`还剩 ${days} 天 ${String(hours).padStart(2, "0")}:${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`);
    };

    update();
    const timer = setInterval(update, 1000);
    return () => clearInterval(timer);
  }, []);

  return <span className="text-xs text-gray-500">{timeLeft}</span>;
}

export default function RightSidebar() {
  const { user } = useAuth();

  return (
    <aside className="w-80 bg-white border-l border-gray-200 h-[calc(100vh-4rem)] sticky top-16 overflow-y-auto hidden xl:block">
      <div className="p-4">
        {user ? (
          /* Logged In User Card */
          <div className="bg-gradient-to-br from-primary to-primary-dark rounded-lg p-6 text-white mb-6">
            <div className="flex items-center space-x-4 mb-4">
              <div className="w-16 h-16 bg-accent rounded-full flex items-center justify-center">
                <User className="h-8 w-8" />
              </div>
              <div>
                <h3 className="text-lg font-bold">{user.user_name}</h3>
                <p className="text-sm text-gray-200">
                  ID: {user.user_id}
                </p>
              </div>
            </div>
            
            <div className="grid grid-cols-2 gap-4 mt-4">
              <div className="bg-white/10 rounded-lg p-3 text-center">
                <div className="text-2xl font-bold">0</div>
                <div className="text-xs text-gray-200">帖子</div>
              </div>
              <div className="bg-white/10 rounded-lg p-3 text-center">
                <div className="text-2xl font-bold">0</div>
                <div className="text-xs text-gray-200">粉丝</div>
              </div>
            </div>
            
            <Link
              href="/submit"
              className="block mt-4 bg-accent hover:bg-yellow-600 text-white text-center py-2 rounded-lg font-medium transition-colors"
            >
              发帖
            </Link>
          </div>
        ) : (
          /* Login/Signup Card */
          <div className="bg-gray-50 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-bold text-gray-900 mb-2">
              欢迎来到 SayIt
            </h3>
            <p className="text-sm text-gray-600 mb-4">
              登录以参与讨论、发帖和投票
            </p>
            <div className="space-y-3">
              <Link
                href="/login"
                className="block w-full bg-primary hover:bg-primary-light text-white text-center py-2 rounded-lg font-medium transition-colors"
              >
                登录
              </Link>
              <Link
                href="/signup"
                className="block w-full bg-white hover:bg-gray-50 text-primary border border-primary text-center py-2 rounded-lg font-medium transition-colors"
              >
                注册
              </Link>
            </div>
          </div>
        )}

        {/* Community Bulletin Board */}
        <div className="bg-gray-50 rounded-lg overflow-hidden mb-6">
          <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200">
            <h3 className="text-sm font-bold text-gray-900 flex items-center">
              <span className="mr-2">📢</span>
              社区活动
            </h3>
            <span className="text-xs text-primary cursor-pointer hover:underline">更多</span>
          </div>

          <div className="divide-y divide-gray-200">
            {/* Ongoing */}
            <div className="px-4 py-3">
              <div className="flex items-center space-x-2 mb-1">
                <span className="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>
                <span className="text-xs font-medium text-red-600">进行中</span>
              </div>
              <p className="text-sm font-medium text-gray-900">2025夏季摄影大赛</p>
              <CountdownTimer />
            </div>

            {/* Upcoming */}
            <div className="px-4 py-3">
              <div className="flex items-center space-x-2 mb-1">
                <span className="w-2 h-2 bg-yellow-500 rounded-full"></span>
                <span className="text-xs font-medium text-yellow-600">即将开始</span>
              </div>
              <p className="text-sm font-medium text-gray-900">AMA：独立开发者圆桌</p>
              <span className="text-xs text-gray-500">明天 20:00 开始</span>
            </div>

            {/* Preview */}
            <div className="px-4 py-3">
              <div className="flex items-center space-x-2 mb-1">
                <span className="w-2 h-2 bg-gray-400 rounded-full"></span>
                <span className="text-xs font-medium text-gray-500">预告</span>
              </div>
              <p className="text-sm font-medium text-gray-900">晒意一周年庆生</p>
              <span className="text-xs text-gray-500">5月20日</span>
            </div>
          </div>
        </div>

        {/* Rules */}
        <div className="bg-gray-50 rounded-lg p-6">
          <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center">
            <Award className="h-5 w-5 mr-2 text-primary" />
            社区规则
          </h3>
          <ul className="space-y-2 text-sm text-gray-600">
            <li className="flex items-start">
              <span className="text-primary mr-2">1.</span>
              尊重他人，友好交流
            </li>
            <li className="flex items-start">
              <span className="text-primary mr-2">2.</span>
              禁止发布垃圾信息
            </li>
            <li className="flex items-start">
              <span className="text-primary mr-2">3.</span>
              保护个人隐私
            </li>
            <li className="flex items-start">
              <span className="text-primary mr-2">4.</span>
              遵守法律法规
            </li>
          </ul>
        </div>

        {/* Footer */}
        <div className="mt-6 text-center">
          <p className="text-base font-bold rainbow-text">晒一个有意思的灵魂</p>
          <div className="mt-2 space-x-4 text-xs text-gray-500">
            <Link href="/about" className="hover:text-primary">关于</Link>
            <Link href="/help" className="hover:text-primary">帮助</Link>
          </div>
        </div>
      </div>
    </aside>
  );
}
