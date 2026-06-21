interface LoaderProps {
  size?: "sm" | "md" | "lg";
}

export function Loader(props: LoaderProps) {
  return <div class={`loader loader--${props.size ?? "md"}`} />;
}
