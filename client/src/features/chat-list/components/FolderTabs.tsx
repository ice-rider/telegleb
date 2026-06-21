import { Show, For } from "solid-js";
import type { Folder } from "../../../types";
import "./FolderTabs.css";

interface FolderTabsProps {
  folders: Folder[];
  selectedId: number | null;
  onSelect: (id: number | null) => void;
}

export function FolderTabs(props: FolderTabsProps) {
  return (
    <Show when={props.folders.length > 0}>
      <div class="folder-tabs">
        <button
          class={`folder-tab ${props.selectedId === null ? "folder-tab--active" : ""}`}
          onClick={() => props.onSelect(null)}
        >
          Все
        </button>
        <For each={props.folders}>
          {(folder) => (
            <button
              class={`folder-tab ${props.selectedId === folder.id ? "folder-tab--active" : ""}`}
              onClick={() => props.onSelect(folder.id)}
            >
              {folder.title}
            </button>
          )}
        </For>
      </div>
    </Show>
  );
}
