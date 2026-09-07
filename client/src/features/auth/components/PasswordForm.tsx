import { Show, createSignal } from "solid-js";
import { Button, Input, Loader } from "~/shared/components";
import { useAuth } from "../store";
import "./PasswordForm.css";

export function PasswordForm() {
  const { isLoading, error, verifyPassword } = useAuth();
  const [password, setPassword] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const value = password().trim();
    if (value) verifyPassword(value);
  }

  return (
    <form class="password-form" onSubmit={handleSubmit}>
      <div class="password-form__icon">🔒</div>
      <h2 class="password-form__title">Двухфакторная аутентификация</h2>
      <p class="password-form__subtitle">
        Введите пароль двухфакторной аутентификации
      </p>
      <Input
        label="Пароль"
        type="password"
        placeholder="Введите пароль"
        value={password()}
        onInput={(e) => setPassword(e.currentTarget.value)}
        error={error() ?? undefined}
      />
      <Button type="submit" fullWidth disabled={isLoading() || !password()}>
        <Show when={!isLoading()} fallback={<Loader size="sm" />}>
          Войти
        </Show>
      </Button>
    </form>
  );
}
