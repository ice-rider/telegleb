import { Show, createSignal } from "solid-js";
import { Button, Icon, Input, Loader } from "~/shared/components";
import { useAuth } from "../store";
import "./AuthForm.css";

export function CodeForm() {
  const { isLoading, error, codeHint, verifyCode } = useAuth();
  const [code, setCode] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const value = code().trim();
    if (value) verifyCode(value);
  }

  return (
    <form class="auth-form" onSubmit={handleSubmit}>
      <div class="auth-form__icon"><Icon name="key" size={22} /></div>
      <h2 class="auth-form__title">Код подтверждения</h2>
      <p class="auth-form__subtitle">{codeHint()}</p>
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
