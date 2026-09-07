import { For, Show } from "solid-js";
import { Icon } from "~/shared/components";
import type { Folder } from "~/types";
import type { Tab } from "../store";
import "./FolderTabs.css";

interface FolderTabsProps {
  folders: Folder[];
  archivedCount: number;
  selected: Tab;
  onSelect: (tab: Tab) => void;
}

function isSame(a: Tab, b: Tab): boolean {
  if (a.kind !== b.kind) return false;
  return a.kind !== "folder" || b.kind !== "folder" || a.id === b.id;
}

export function FolderTabs(props: FolderTabsProps) {
  const tabClass = (tab: Tab) =>
    `folder-tab ${isSame(props.selected, tab) ? "folder-tab--active" : ""}`;

  return (
    <div class="folder-tabs">
      <button class={tabClass({ kind: "all" })} onClick={() => props.onSelect({ kind: "all" })}>
        Все
      </button>

      <For each={props.folders}>
        {(folder) => (
          <button
            class={tabClass({ kind: "folder", id: folder.id })}
            onClick={() => props.onSelect({ kind: "folder", id: folder.id })}
          >
            <Show when={folder.emoticon}>
              <span class="folder-tab__icon">{folder.emoticon}</span>
            </Show>
            {folder.title}
          </button>
        )}
      </For>

      {/* Архив — обычная вкладка, но его чаты никогда не попадают в «Все». */}
      <Show when={props.archivedCount > 0}>
        <button
          class={tabClass({ kind: "archive" })}
          onClick={() => props.onSelect({ kind: "archive" })}
          title="Архив"
        >
          <Icon name="archive" size={14} />
          <span class="folder-tab__count mono">{props.archivedCount}</span>
        </button>
      </Show>
    </div>
  );
}
