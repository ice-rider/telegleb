import "./SearchBar.css";

interface SearchBarProps {
  value: string;
  onInput: (value: string) => void;
  placeholder?: string;
}

export function SearchBar(props: SearchBarProps) {
  return (
    <div class="search-bar">
      <svg class="search-bar__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8" />
        <path d="M21 21l-4.35-4.35" />
      </svg>
      <input
        class="search-bar__input"
        type="text"
        placeholder={props.placeholder ?? "Поиск чатов..."}
        value={props.value}
        onInput={(e) => props.onInput(e.currentTarget.value)}
      />
    </div>
  );
}
