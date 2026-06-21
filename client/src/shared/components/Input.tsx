import { splitProps, type JSX } from "solid-js";

interface InputProps extends JSX.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export function Input(props: InputProps) {
  const [local, others] = splitProps(props, [
    "label",
    "error",
    "class",
  ]);

  const classes = () =>
    ["input-wrapper", local.error && "input-wrapper--error", local.class]
      .filter(Boolean)
      .join(" ");

  return (
    <div class={classes()}>
      {local.label && <label class="input-label">{local.label}</label>}
      <input class="input" {...others} />
      {local.error && <span class="input-error">{local.error}</span>}
    </div>
  );
}
