import { Show, createResource, onCleanup } from "solid-js";
import { api } from "~/core/api";
import type { Media } from "~/types";

interface MessageMediaProps {
  media: Media;
}

const VISUAL_KINDS = new Set(["photo", "sticker", "gif"]);

export function MessageMedia(props: MessageMediaProps) {
  // Вложение тянется через fetch с заголовком Authorization: тег <img> свои
  // заголовки отправить не может, а токен в query-строке утекал бы в логи.
  const [objectUrl] = createResource(
    () => (VISUAL_KINDS.has(props.media.kind) ? props.media.ref : undefined),
    (ref) => api.mediaObjectUrl(ref),
  );

  onCleanup(() => {
    const url = objectUrl.latest;
    if (url) URL.revokeObjectURL(url);
  });

  return (
    <Show
      when={VISUAL_KINDS.has(props.media.kind)}
      fallback={
        <div class="message__attachment">
          📎 {props.media.fileName || props.media.kind}
          <Show when={props.media.size}>
            {(size) => <span class="message__attachment-size">{formatSize(size())}</span>}
          </Show>
        </div>
      }
    >
      <Show when={objectUrl()} fallback={<div class="message__media-placeholder" />}>
        {(url) => <img class="message__media" src={url()} alt={props.media.fileName || ""} />}
      </Show>
    </Show>
  );
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} Б`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} КБ`;
  return `${(bytes / 1024 / 1024).toFixed(1)} МБ`;
}
