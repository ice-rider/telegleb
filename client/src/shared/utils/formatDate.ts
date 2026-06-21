export function formatTime(date: string | Date): string {
  const d = new Date(date);
  return d.toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatDate(date: string | Date): string {
  const d = new Date(date);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  const dayMs = 86400000;

  if (diff < dayMs && now.getDate() === d.getDate()) {
    return formatTime(d);
  }
  if (diff < dayMs * 2) {
    return "Вчера";
  }
  if (diff < dayMs * 7) {
    return d.toLocaleDateString("ru-RU", { weekday: "short" });
  }
  return d.toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "short",
  });
}

export function formatLastMessage(text: string, maxLen = 40): string {
  if (text.length <= maxLen) return text;
  return text.slice(0, maxLen) + "...";
}
