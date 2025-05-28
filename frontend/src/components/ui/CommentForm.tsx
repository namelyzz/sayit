"use client";

import { useState } from "react";
import { MessageSquare } from "lucide-react";
import Button from "./Button";
import { Textarea } from "./Field";

interface CommentFormProps {
  onSubmit: (content: string) => Promise<void>;
  placeholder?: string;
  autoFocus?: boolean;
  compact?: boolean;
}

export default function CommentForm({
  onSubmit,
  placeholder = "写下你的评论...",
  autoFocus = false,
  compact = false,
}: CommentFormProps) {
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
    const trimmed = content.trim();
    if (!trimmed || submitting) return;

    setSubmitting(true);
    try {
      await onSubmit(trimmed);
      setContent("");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <Textarea
        rows={compact ? 2 : 3}
        placeholder={placeholder}
        value={content}
        onChange={(e) => setContent(e.target.value)}
        autoFocus={autoFocus}
        disabled={submitting}
      />
      <div className="mt-2 flex justify-end">
        <Button
          onClick={handleSubmit}
          disabled={!content.trim() || submitting}
          size={compact ? "sm" : "md"}
        >
          <MessageSquare className="h-4 w-4" />
          {submitting ? "发送中..." : "发表评论"}
        </Button>
      </div>
    </div>
  );
}
