import type {
  InputHTMLAttributes,
  LabelHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from "react";
import { cn } from "@/lib/utils";

export function FieldLabel({
  className,
  children,
  ...props
}: LabelHTMLAttributes<HTMLLabelElement>) {
  return (
    <label className={cn("mb-2 block text-sm font-medium text-muted-strong", className)} {...props}>
      {children}
    </label>
  );
}

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "h-11 w-full rounded-lg border border-border bg-surface px-3 text-sm text-foreground shadow-sm transition placeholder:text-muted focus:border-primary focus:outline-none focus:ring-4 focus:ring-teal-100 disabled:bg-surface-soft",
        className
      )}
      {...props}
    />
  );
}

export function Select({ className, children, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={cn(
        "h-11 w-full rounded-lg border border-border bg-surface px-3 text-sm text-foreground shadow-sm transition focus:border-primary focus:outline-none focus:ring-4 focus:ring-teal-100 disabled:bg-surface-soft",
        className
      )}
      {...props}
    >
      {children}
    </select>
  );
}

export function Textarea({ className, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(
        "w-full rounded-lg border border-border bg-surface px-3 py-3 text-sm text-foreground shadow-sm transition placeholder:text-muted focus:border-primary focus:outline-none focus:ring-4 focus:ring-teal-100 disabled:bg-surface-soft",
        className
      )}
      {...props}
    />
  );
}

export function FieldHint({ children }: { children: ReactNode }) {
  return <p className="mt-2 text-xs text-muted">{children}</p>;
}
