import { Show, createSignal } from "solid-js";
import { Button, Icon, Input, Loader } from "~/shared/components";
import { useAuth } from "../store";
import "./AuthForm.css";

export function PhoneForm() {
  const { isLoading, error, requestCode } = useAuth();
  const [phone, setPhone] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const value = phone().trim();
    if (value) requestCode(value);
  }

  return (
    <form class="auth-form" onSubmit={handleSubmit}>
      <div class="auth-form__icon"><Icon name="phone" size={22} /></div>
      <h2 class="auth-form__title">Вход в Telegleb</h2>
      <p class="auth-form__subtitle">
        Введите номер телефона для входа в аккаунт Telegram
      </p>
      <Input
        label="Номер телефона"
        type="tel"
        placeholder="+79001234567"
        value={phone()}
        onInput={(e) => setPhone(e.currentTarget.value)}
        error={error() ?? undefined}
      />
      <Button type="submit" fullWidth disabled={isLoading() || !phone()}>
        <Show when={!isLoading()} fallback={<Loader size="sm" />}>
          Получить код
        </Show>
      </Button>
    </form>
  );
}
