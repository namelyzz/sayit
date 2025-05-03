import Link from "next/link";
import { getCommunityPalette } from "@/lib/community-colors";

interface CommunityBadgeProps {
  id?: string;
  name?: string;
  href?: string;
}

export default function CommunityBadge({ id, name, href }: CommunityBadgeProps) {
  const label = name || "未命名社区";
  const palette = getCommunityPalette(id || label);
  const style = {
    backgroundColor: palette.bg,
    borderColor: palette.border,
    color: palette.text,
  };
  const content = (
    <>
      <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: palette.dot }} />
      <span className="truncate">{label}</span>
    </>
  );

  if (href) {
    return (
      <Link
        href={href}
        className="inline-flex max-w-full items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs font-semibold transition hover:bg-[#9caba9]/20"
        style={style}
      >
        {content}
      </Link>
    );
  }

  return (
    <span
      className="inline-flex max-w-full items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs font-semibold"
      style={style}
    >
      {content}
    </span>
  );
}
