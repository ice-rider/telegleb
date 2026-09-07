import { Show, createSignal } from "solid-js";
import { Icon, Loader } from "~/shared/components";
import "./MessageInput.css";

interface MessageInputProps {
  onSend: (text: string) => void;
  isSending: boolean;
}

export function MessageInput(props: MessageInputProps) {
  const [text, setText] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const value = text().trim();
    if (value && !props.isSending) {
      props.onSend(value);
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
      <Icon name="attach" size={19} class="message-input__attach" />
      <textarea
        class="message-input__field"
        placeholder="Написать сообщение"
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
          <Icon name="send" size={18} filled />
        </Show>
      </button>
    </form>
  );
}
