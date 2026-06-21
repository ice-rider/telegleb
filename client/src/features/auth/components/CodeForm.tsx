import { Show, createSignal } from "solid-js";
import { Button } from "../../../shared/components";
import { Input } from "../../../shared/components";
import { Loader } from "../../../shared/components";
import { useAuth } from "../store";
import "./CodeForm.css";

export function CodeForm() {
  const { state, verifyCode } = useAuth();
  const [code, setCode] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    if (code().trim()) {
      verifyCode(code().trim());
    }
  }

  return (
    <form class="code-form" onSubmit={handleSubmit}>
      <div class="code-form__icon">🔑</div>
      <h2 class="code-form__title">Код подтверждения</h2>
      <p class="code-form__subtitle">
        Мы отправили код в Telegram. Введите его ниже.
      </p>
      <Input
        label="Код"
        type="text"
        placeholder="12345"
        value={code()}
        onInput={(e) => setCode(e.currentTarget.value)}
        error={state().error ?? undefined}
        autocomplete="one-time-code"
      />
      <Button type="submit" fullWidth disabled={state().isLoading || !code()}>
        <Show when={!state().isLoading} fallback={<Loader size="sm" />}>
          Подтвердить
        </Show>
      </Button>
    </form>
  );
}
