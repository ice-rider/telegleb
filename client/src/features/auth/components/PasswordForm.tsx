import { Show, createSignal } from "solid-js";
import { Button } from "../../../shared/components";
import { Input } from "../../../shared/components";
import { Loader } from "../../../shared/components";
import { useAuth } from "../store";
import "./PasswordForm.css";

export function PasswordForm() {
  const { state, verifyPassword } = useAuth();
  const [password, setPassword] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    if (password().trim()) {
      verifyPassword(password().trim());
    }
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
        error={state().error ?? undefined}
      />
      <Button
        type="submit"
        fullWidth
        disabled={state().isLoading || !password()}
      >
        <Show when={!state().isLoading} fallback={<Loader size="sm" />}>
          Войти
        </Show>
      </Button>
    </form>
  );
}
