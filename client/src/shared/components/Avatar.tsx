interface AvatarProps {
  name: string;
  size?: "sm" | "md" | "lg";
  url?: string;
}

const COLORS = [
  "#e17076",
  "#7bc862",
  "#e5ca77",
  "#65aadd",
  "#a695e7",
  "#ee7aae",
  "#6ec9cb",
  "#faa774",
];

function getColor(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  return COLORS[Math.abs(hash) % COLORS.length];
}

function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return name.slice(0, 2).toUpperCase();
}

export function Avatar(props: AvatarProps) {
  const sizeClass = () => `avatar--${props.size ?? "md"}`;

  return (
    <div class={`avatar ${sizeClass()}`} style={{ "background-color": getColor(props.name) }}>
      {props.url ? (
        <img class="avatar__img" src={props.url} alt={props.name} />
      ) : (
        <span class="avatar__initials">{getInitials(props.name)}</span>
      )}
    </div>
  );
}
