import { Show, createSignal } from "solid-js";
import { Button, Input, Loader } from "~/shared/components";
import { useAuth } from "../store";
import "./CodeForm.css";

export function CodeForm() {
  const { isLoading, error, codeHint, verifyCode } = useAuth();
  const [code, setCode] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const value = code().trim();
    if (value) verifyCode(value);
  }

  return (
    <form class="code-form" onSubmit={handleSubmit}>
      <div class="code-form__icon">🔑</div>
      <h2 class="code-form__title">Код подтверждения</h2>
      <p class="code-form__subtitle">{codeHint()}</p>
      <Input
        label="Код"
        type="text"
        placeholder="12345"
        value={code()}
        onInput={(e) => setCode(e.currentTarget.value)}
        error={error() ?? undefined}
        autocomplete="one-time-code"
      />
      <Button type="submit" fullWidth disabled={isLoading() || !code()}>
        <Show when={!isLoading()} fallback={<Loader size="sm" />}>
          Подтвердить
        </Show>
      </Button>
    </form>
  );
}
