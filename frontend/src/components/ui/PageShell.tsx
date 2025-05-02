import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

interface PageShellProps {
  children: ReactNode;
  className?: string;
}

export default function PageShell({ children, className }: PageShellProps) {
  return <div className={cn("mx-auto w-full max-w-3xl space-y-5", className)}>{children}</div>;
}
