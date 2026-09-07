export function formatTime(date: string | Date): string {
  return new Date(date).toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatDate(date: string | Date): string {
  const d = new Date(date);
  const now = new Date();

  const startOfDay = (x: Date) => new Date(x.getFullYear(), x.getMonth(), x.getDate()).getTime();
  const days = Math.round((startOfDay(now) - startOfDay(d)) / 86400000);

  if (days <= 0) return formatTime(d);
  if (days === 1) return "Вчера";
  if (days < 7) return d.toLocaleDateString("ru-RU", { weekday: "short" });
  if (d.getFullYear() === now.getFullYear()) {
    return d.toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
  }
  return d.toLocaleDateString("ru-RU", { day: "numeric", month: "short", year: "numeric" });
}

export function formatDayLabel(date: string | Date): string {
  return new Date(date).toLocaleDateString("ru-RU", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

export function formatLastMessage(text: string, maxLen = 40): string {
  const oneLine = text.replace(/\s+/g, " ").trim();
  return oneLine.length <= maxLen ? oneLine : oneLine.slice(0, maxLen) + "…";
}
