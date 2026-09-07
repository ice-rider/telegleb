import { Show, createSignal } from "solid-js";
import { Button, Icon, Input, Loader } from "~/shared/components";
import { useAuth } from "../store";
import "./AuthForm.css";

export function PasswordForm() {
  const { isLoading, error, verifyPassword } = useAuth();
  const [password, setPassword] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const value = password().trim();
    if (value) verifyPassword(value);
  }

  return (
    <form class="auth-form" onSubmit={handleSubmit}>
      <div class="auth-form__icon"><Icon name="lock" size={22} /></div>
      <h2 class="auth-form__title">Двухфакторная аутентификация</h2>
      <p class="auth-form__subtitle">
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
