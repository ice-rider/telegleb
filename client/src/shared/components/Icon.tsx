import { Show } from "solid-js";
import type { JSX } from "solid-js";

/**
 * Набор иконок приложения.
 *
 * Рисуются как inline SVG, а не эмодзи: эмодзи по-разному выглядят в разных
 * системах, не перекрашиваются под тему и в тёмном интерфейсе смотрятся
 * инородно. Все контуры на сетке 24 со штрихом 1.8 — так они одинаково
 * читаются в любом размере.
 */
export type IconName =
  | "search"
  | "send"
  | "attach"
  | "back"
  | "plus"
  | "menu"
  | "more"
  | "logout"
  | "star"
  | "pin"
  | "muted"
  | "group"
  | "channel"
  | "archive"
  | "forum"
  | "lock"
  | "eye"
  | "phone"
  | "key"
  | "chat"
  | "photo"
  | "video"
  | "mic"
  | "music"
  | "file"
  | "check";

interface IconProps {
  name: IconName;
  size?: number;
  /** Заливка вместо контура — для звезды приоритета и кнопки отправки. */
  filled?: boolean;
  class?: string;
  style?: JSX.CSSProperties;
  title?: string;
}

const STAR = "M12 3l2.6 5.6 6.1.8-4.5 4.2 1.2 6-5.4-3-5.4 3 1.2-6L3.3 9.4l6.1-.8z";
const SEND = "M3 20.5l18-8.5L3 3.5V10l12 2-12 2z";

const PATHS: Record<IconName, JSX.Element> = {
  search: (
    <>
      <circle cx="11" cy="11" r="7" />
      <path d="M20 20l-3.6-3.6" />
    </>
  ),
  send: <path d={SEND} />,
  attach: <path d="M21 12.5l-8.4 8.4a5 5 0 0 1-7-7l8.9-8.9a3.3 3.3 0 0 1 4.7 4.7l-8.9 8.9a1.7 1.7 0 0 1-2.3-2.3l8.2-8.2" />,
  back: <path d="M15 5l-7 7 7 7" />,
  plus: <path d="M12 5v14M5 12h14" />,
  menu: <path d="M4 7h16M4 12h16M4 17h10" />,
  more: (
    <>
      <circle cx="12" cy="5" r="1.4" />
      <circle cx="12" cy="12" r="1.4" />
      <circle cx="12" cy="19" r="1.4" />
    </>
  ),
  logout: (
    <>
      <path d="M9 20H5V4h4" />
      <path d="M16 16l4-4-4-4M20 12H10" />
    </>
  ),
  star: <path d={STAR} />,
  pin: (
    <>
      <path d="M9 4h6l-1 6 4 3v2H6v-2l4-3z" />
      <path d="M12 15v5" />
    </>
  ),
  muted: (
    <>
      <path d="M6 8.5a6 6 0 0 1 12 0c0 4 1.6 5.5 1.6 5.5H4.4S6 12.5 6 8.5z" />
      <path d="M10 18a2 2 0 0 0 4 0" />
      <path d="M4 4l16 16" />
    </>
  ),
  group: (
    <>
      <circle cx="9" cy="9" r="3.2" />
      <path d="M3 19c0-3 2.7-4.6 6-4.6s6 1.6 6 4.6" />
      <path d="M16 6.5a3 3 0 0 1 0 5.6M17.5 19c0-2 .6-3.4-1-4.4" />
    </>
  ),
  channel: (
    <>
      <path d="M4 10v4h3l6 4V6l-6 4H4z" />
      <path d="M17 9.5a4 4 0 0 1 0 5" />
    </>
  ),
  archive: (
    <>
      <path d="M4 8h16v11H4z" />
      <path d="M2 5h20v3H2zM10 12h4" />
    </>
  ),
  forum: (
    <>
      <path d="M4 5h16v11H9l-5 4z" />
      <path d="M8 9h8M8 12.5h5" />
    </>
  ),
  lock: (
    <>
      <rect x="5" y="11" width="14" height="9" rx="2" />
      <path d="M8.5 11V8a3.5 3.5 0 0 1 7 0v3" />
    </>
  ),
  eye: (
    <>
      <path d="M2 12s3.6-6.5 10-6.5S22 12 22 12s-3.6 6.5-10 6.5S2 12 2 12z" />
      <circle cx="12" cy="12" r="2.6" />
    </>
  ),
  phone: (
    <>
      <rect x="6" y="2.5" width="12" height="19" rx="3" />
      <path d="M10.5 18.5h3" />
    </>
  ),
  key: (
    <>
      <circle cx="8" cy="12" r="4" />
      <path d="M12 12h9M17.5 12v3.5M20.5 12v2.5" />
    </>
  ),
  chat: <path d="M4 5h16v11H9l-5 4z" />,
  photo: (
    <>
      <rect x="3" y="5" width="18" height="14" rx="2.5" />
      <circle cx="8.5" cy="10" r="1.6" />
      <path d="M4 17l5-4.5 4 3.5 3-2.5 4 3.5" />
    </>
  ),
  video: (
    <>
      <rect x="3" y="6" width="12" height="12" rx="2.5" />
      <path d="M15 10.5l6-3v9l-6-3z" />
    </>
  ),
  mic: (
    <>
      <rect x="9" y="3" width="6" height="10" rx="3" />
      <path d="M5.5 11a6.5 6.5 0 0 0 13 0M12 17.5V21" />
    </>
  ),
  music: (
    <>
      <path d="M9 18V6l10-2v12" />
      <circle cx="6.5" cy="18" r="2.5" />
      <circle cx="16.5" cy="16" r="2.5" />
    </>
  ),
  file: (
    <>
      <path d="M13 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V9z" />
      <path d="M13 3v6h6" />
    </>
  ),
  check: <path d="M4 12.5l5 5L20 6.5" />,
};

export function Icon(props: IconProps) {
  const size = () => props.size ?? 18;

  return (
    <svg
      class={props.class}
      style={props.style}
      width={size()}
      height={size()}
      viewBox="0 0 24 24"
      fill={props.filled ? "currentColor" : "none"}
      stroke={props.filled ? "none" : "currentColor"}
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden={props.title ? undefined : "true"}
      role={props.title ? "img" : undefined}
    >
      <Show when={props.title}>{(t) => <title>{t()}</title>}</Show>
      {PATHS[props.name]}
    </svg>
  );
}
