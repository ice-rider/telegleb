import { Icon } from "~/shared/components";
import "./SearchBar.css";

interface SearchBarProps {
  value: string;
  onInput: (value: string) => void;
  placeholder?: string;
}

export function SearchBar(props: SearchBarProps) {
  return (
    <div class="search-bar">
      <Icon name="search" size={17} class="search-bar__icon" />
      <input
        class="search-bar__input"
        type="text"
        placeholder={props.placeholder ?? "Поиск по чатам"}
        value={props.value}
        onInput={(e) => props.onInput(e.currentTarget.value)}
      />
    </div>
  );
}
