"use client";

import Link from "next/link";
import { useState } from "react";
import { Search, Bell, User, Menu, X, LogOut } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

export default function Navbar() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const { user, logout } = useAuth();

  return (
    <nav className="bg-primary text-white shadow-lg sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo and Brand */}
          <div className="flex items-center">
            <Link href="/" className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-accent rounded-full flex items-center justify-center">
                <span className="font-bold text-lg">S</span>
              </div>
              <span className="text-xl font-bold">SayIt</span>
            </Link>
          </div>

          {/* Search Bar - Desktop */}
          <div className="hidden md:flex flex-1 max-w-md mx-8">
            <div className="relative w-full">
              <input
                type="text"
                placeholder="搜索帖子或社区..."
                className="w-full px-4 py-2 pl-10 bg-primary-light text-white placeholder-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-accent"
              />
              <Search className="absolute left-3 top-2.5 h-5 w-5 text-gray-300" />
            </div>
          </div>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center space-x-4">
            <Link
              href="/submit"
              className="bg-accent hover:bg-yellow-600 text-white px-4 py-2 rounded-lg font-medium transition-colors"
            >
              发帖
            </Link>
            <button className="p-2 hover:bg-primary-light rounded-lg transition-colors">
              <Bell className="h-5 w-5" />
            </button>
            
            {user ? (
              <div className="flex items-center space-x-4">
                <Link
                  href={`/user/${user.user_id}`}
                  className="flex items-center space-x-2 hover:bg-primary-light px-3 py-2 rounded-lg transition-colors"
                >
                  <User className="h-5 w-5" />
                  <span>{user.user_name}</span>
                </Link>
                <button
                  onClick={logout}
                  className="flex items-center space-x-2 hover:bg-primary-light px-3 py-2 rounded-lg transition-colors"
                >
                  <LogOut className="h-5 w-5" />
                  <span>退出</span>
                </button>
              </div>
            ) : (
              <Link
                href="/login"
                className="flex items-center space-x-2 hover:bg-primary-light px-3 py-2 rounded-lg transition-colors"
              >
                <User className="h-5 w-5" />
                <span>登录</span>
              </Link>
            )}
          </div>

          {/* Mobile Menu Button */}
          <div className="md:hidden">
            <button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="p-2 hover:bg-primary-light rounded-lg transition-colors"
            >
              {isMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
            </button>
          </div>
        </div>

        {/* Mobile Search Bar */}
        {isMenuOpen && (
          <div className="md:hidden pb-4">
            <div className="relative">
              <input
                type="text"
                placeholder="搜索帖子或社区..."
                className="w-full px-4 py-2 pl-10 bg-primary-light text-white placeholder-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-accent"
              />
              <Search className="absolute left-3 top-2.5 h-5 w-5 text-gray-300" />
            </div>
            <div className="mt-4 space-y-2">
              <Link
                href="/submit"
                className="block bg-accent hover:bg-yellow-600 text-white px-4 py-2 rounded-lg font-medium transition-colors text-center"
              >
                发帖
              </Link>
              
              {user ? (
                <>
                  <Link
                    href={`/user/${user.user_id}`}
                    className="block hover:bg-primary-light px-4 py-2 rounded-lg transition-colors text-center"
                  >
                    {user.user_name}
                  </Link>
                  <button
                    onClick={logout}
                    className="block w-full hover:bg-primary-light px-4 py-2 rounded-lg transition-colors text-center"
                  >
                    退出
                  </button>
                </>
              ) : (
                <Link
                  href="/login"
                  className="block hover:bg-primary-light px-4 py-2 rounded-lg transition-colors text-center"
                >
                  登录
                </Link>
              )}
            </div>
          </div>
        )}
      </div>
    </nav>
  );
}
