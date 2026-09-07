import { splitProps, type JSX } from "solid-js";

interface ButtonProps extends JSX.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "ghost";
  fullWidth?: boolean;
}

export function Button(props: ButtonProps) {
  const [local, others] = splitProps(props, ["variant", "fullWidth", "class", "children"]);

  const classes = () =>
    [
      "button",
      local.variant === "ghost" && "button--ghost",
      local.fullWidth && "button--full",
      local.class,
    ]
      .filter(Boolean)
      .join(" ");

  return (
    <button class={classes()} {...others}>
      {local.children}
    </button>
  );
}
