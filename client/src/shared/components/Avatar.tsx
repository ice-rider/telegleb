import { Show } from "solid-js";

interface AvatarProps {
  name: string;
  size?: "sm" | "md" | "lg";
  url?: string;
}

/**
 * Палитра подложек подобрана под тёмный фон: приглушённые, чтобы аватар не
 * перетягивал внимание с акцента, но различимые между собой.
 */
const TONES = [
  { bg: "#2f3640", fg: "#dfe3e8" },
  { bg: "#243040", fg: "#9dc0e8" },
  { bg: "#2a2f26", fg: "#c2d69a" },
  { bg: "#332a3d", fg: "#cbaee0" },
  { bg: "#3d2f2a", fg: "#e0b8a4" },
  { bg: "#25353a", fg: "#98c9d4" },
  { bg: "#38312a", fg: "#d8c3a5" },
  { bg: "#2b3340", fg: "#aebdd4" },
];

function toneOf(str: string) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  return TONES[Math.abs(hash) % TONES.length];
}

function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
  return parts[0].slice(0, 2).toUpperCase();
}

export function Avatar(props: AvatarProps) {
  const tone = () => toneOf(props.name || "?");

  return (
    <div
      class={`avatar avatar--${props.size ?? "md"}`}
      style={{ "background-color": tone().bg, color: tone().fg }}
    >
      <Show when={props.url} fallback={<span class="avatar__initials">{initials(props.name)}</span>}>
        {(url) => <img class="avatar__img" src={url()} alt="" />}
      </Show>
    </div>
  );
}
