import { For, Show } from "solid-js";
import type { MessageEntity } from "~/types";

interface MessageTextProps {
  text: string;
  entities: MessageEntity[];
}

interface Segment {
  text: string;
  entity?: MessageEntity;
}

/**
 * Смещения Telegram считаются в кодовых единицах UTF-16, что совпадает с
 * индексацией строк в JavaScript. Вложенность не поддерживается: при
 * пересечении выигрывает первая по порядку сущность.
 */
function toSegments(text: string, entities: MessageEntity[]): Segment[] {
  const sorted = [...entities].sort((a, b) => a.offset - b.offset);
  const segments: Segment[] = [];
  let cursor = 0;

  for (const entity of sorted) {
    if (entity.offset < cursor) continue;
    const end = entity.offset + entity.length;
    if (entity.offset > text.length) break;

    if (entity.offset > cursor) {
      segments.push({ text: text.slice(cursor, entity.offset) });
    }
    segments.push({ text: text.slice(entity.offset, end), entity });
    cursor = end;
  }

  if (cursor < text.length) {
    segments.push({ text: text.slice(cursor) });
  }
  return segments;
}

export function MessageText(props: MessageTextProps) {
  const segments = () => toSegments(props.text, props.entities ?? []);

  return (
    <p class="message__text">
      <For each={segments()}>{(segment) => <Piece segment={segment} />}</For>
    </p>
  );
}

function Piece(props: { segment: Segment }) {
  const entity = () => props.segment.entity;
  const text = () => props.segment.text;

  return (
    <Show when={entity()} fallback={<>{text()}</>}>
      {(e) => {
        switch (e().type) {
          case "bold":
            return <strong>{text()}</strong>;
          case "italic":
            return <em>{text()}</em>;
          case "underline":
            return <u>{text()}</u>;
          case "strike":
            return <s>{text()}</s>;
          case "code":
            return <code class="message__code">{text()}</code>;
          case "pre":
            return <pre class="message__pre">{text()}</pre>;
          case "blockquote":
            return <blockquote class="message__quote">{text()}</blockquote>;
          case "spoiler":
            return <span class="message__spoiler">{text()}</span>;
          case "url":
            return (
              <a class="message__link" href={text()} target="_blank" rel="noreferrer noopener">
                {text()}
              </a>
            );
          case "text_url":
            return (
              <a class="message__link" href={e().url} target="_blank" rel="noreferrer noopener">
                {text()}
              </a>
            );
          case "mention":
          case "mention_name":
          case "hashtag":
            return <span class="message__mention">{text()}</span>;
          default:
            return <>{text()}</>;
        }
      }}
    </Show>
  );
}
