import { Show, createSignal } from "solid-js";
import { Button } from "../../../shared/components";
import { Input } from "../../../shared/components";
import { Loader } from "../../../shared/components";
import { useAuth } from "../store";
import "./PhoneForm.css";

export function PhoneForm() {
  const { state, requestLogin } = useAuth();
  const [phone, setPhone] = createSignal("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const p = phone().trim();
    if (p && p.startsWith("+")) {
      requestLogin(p);
    }
  }

  return (
    <form class="phone-form" onSubmit={handleSubmit}>
      <div class="phone-form__icon">📱</div>
      <h2 class="phone-form__title">Вход в Telegleb</h2>
      <p class="phone-form__subtitle">
        Введите номер телефона для входа в аккаунт Telegram
      </p>
      <Input
        label="Номер телефона"
        type="tel"
        placeholder="+7 900 123 45 67"
        value={phone()}
        onInput={(e) => setPhone(e.currentTarget.value)}
        error={state().error ?? undefined}
      />
      <Button type="submit" fullWidth disabled={state().isLoading || !phone()}>
        <Show when={!state().isLoading} fallback={<Loader size="sm" />}>
          Получить код
        </Show>
      </Button>
    </form>
  );
}
