import { createSignal, Show } from "solid-js";
import { Loader } from "../../../shared/components";
import "./MessageInput.css";

interface MessageInputProps {
  onSend: (text: string) => void;
  isSending: boolean;
}

export function MessageInput(props: MessageInputProps) {
  const [text, setText] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const t = text().trim();
    if (t && !props.isSending) {
      props.onSend(t);
      setText("");
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  }

  return (
    <form class="message-input" onSubmit={handleSubmit}>
      <textarea
        class="message-input__field"
        placeholder="Напишите сообщение..."
        rows={1}
        value={text()}
        onInput={(e) => setText(e.currentTarget.value)}
        onKeyDown={handleKeyDown}
      />
      <button
        type="submit"
        class="message-input__send"
        disabled={!text().trim() || props.isSending}
      >
        <Show when={!props.isSending} fallback={<Loader size="sm" />}>
          <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
            <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
          </svg>
        </Show>
      </button>
    </form>
  );
}
