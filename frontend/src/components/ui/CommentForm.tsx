"use client";

import { useState } from "react";
import { MessageSquare } from "lucide-react";
import { getErrorMessage } from "@/lib/utils";
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
  const [error, setError] = useState("");

  const handleSubmit = async () => {
    const trimmed = content.trim();
    if (!trimmed || submitting) return;

    setSubmitting(true);
    setError("");
    try {
      await onSubmit(trimmed);
      setContent("");
    } catch (err) {
      setError(getErrorMessage(err, "评论失败"));
    } finally {
      setSubmitting(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setContent(e.target.value);
    if (error) setError("");
  };

  return (
    <div>
      <Textarea
        rows={compact ? 2 : 3}
        placeholder={placeholder}
        value={content}
        onChange={handleChange}
        autoFocus={autoFocus}
        disabled={submitting}
      />
      {error && (
        <div className="mt-2 text-sm text-danger">{error}</div>
      )}
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
