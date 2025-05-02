export function cn(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(" ");
}

export function getErrorMessage(error: unknown, fallback = "请求失败，请稍后再试") {
  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }
  return fallback;
}
