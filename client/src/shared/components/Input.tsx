import { Show, splitProps, type JSX } from "solid-js";

interface InputProps extends JSX.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export function Input(props: InputProps) {
  const [local, others] = splitProps(props, ["label", "error", "class"]);

  return (
    <div class={["input", local.class].filter(Boolean).join(" ")}>
      <Show when={local.label}>{(label) => <label class="input__label">{label()}</label>}</Show>
      <input
        class={`input__field ${local.error ? "input__field--error" : ""}`}
        {...others}
      />
      <Show when={local.error}>{(err) => <span class="input__error">{err()}</span>}</Show>
    </div>
  );
}
