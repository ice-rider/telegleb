import { splitProps, type JSX } from "solid-js";

interface ButtonProps extends JSX.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "ghost";
  size?: "sm" | "md" | "lg";
  fullWidth?: boolean;
}

export function Button(props: ButtonProps) {
  const [local, others] = splitProps(props, [
    "variant",
    "size",
    "fullWidth",
    "class",
    "children",
  ]);

  const classes = () =>
    [
      "btn",
      `btn--${local.variant ?? "primary"}`,
      `btn--${local.size ?? "md"}`,
      local.fullWidth && "btn--full",
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
