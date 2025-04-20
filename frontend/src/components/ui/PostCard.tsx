"use client";

import Link from "next/link";
import { ArrowBigUp, ArrowBigDown, MessageSquare, Share2 } from "lucide-react";

interface PostCardProps {
  post: {
    post_id: string;
    title: string;
    summary: string;
    user_name: string;
    community_id: string;
    community_name: string;
    create_time: string;
    like_count: number;
    comment_count: number;
  };
}

export default function PostCard({ post }: PostCardProps) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
      <div className="flex">
        {/* Vote Buttons */}
        <div className="flex flex-col items-center p-3 bg-gray-50 rounded-l-lg">
          <button className="p-1 hover:bg-gray-200 rounded transition-colors">
            <ArrowBigUp className="h-6 w-6 text-gray-500 hover:text-primary" />
          </button>
          <span className="text-sm font-bold text-gray-700 my-1">
            {post.like_count}
          </span>
          <button className="p-1 hover:bg-gray-200 rounded transition-colors">
            <ArrowBigDown className="h-6 w-6 text-gray-500 hover:text-red-500" />
          </button>
        </div>

        {/* Post Content */}
        <div className="flex-1 p-4">
          {/* Meta Info */}
          <div className="flex items-center text-sm text-gray-500 mb-2">
            <Link
              href={`/community/${post.community_id}`}
              className="font-medium text-primary hover:underline"
            >
              {post.community_name}
            </Link>
            <span className="mx-2">•</span>
            <span>发布者: {post.user_name}</span>
            <span className="mx-2">•</span>
            <span>{post.create_time}</span>
          </div>

          {/* Title */}
          <Link href={`/post/${post.post_id}`}>
            <h2 className="text-lg font-semibold text-gray-900 hover:text-primary transition-colors mb-2">
              {post.title}
            </h2>
          </Link>

          {/* Summary */}
          <p className="text-gray-600 text-sm mb-4 line-clamp-3">
            {post.summary}
          </p>

          {/* Actions */}
          <div className="flex items-center space-x-4 text-sm text-gray-500">
            <Link
              href={`/post/${post.post_id}`}
              className="flex items-center space-x-1 hover:bg-gray-100 px-2 py-1 rounded transition-colors"
            >
              <MessageSquare className="h-4 w-4" />
              <span>{post.comment_count} 评论</span>
            </Link>
            <button className="flex items-center space-x-1 hover:bg-gray-100 px-2 py-1 rounded transition-colors">
              <Share2 className="h-4 w-4" />
              <span>分享</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
