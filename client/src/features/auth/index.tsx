import { Match, Switch } from "solid-js";
import { CodeForm } from "./components/CodeForm";
import { PasswordForm } from "./components/PasswordForm";
import { PhoneForm } from "./components/PhoneForm";
import { useAuth } from "./store";

export { useAuth } from "./store";

export function AuthFeature() {
  const { status } = useAuth();

  return (
    <div class="auth-page">
      <Switch fallback={<PhoneForm />}>
        <Match when={status() === "awaitingCode"}>
          <CodeForm />
        </Match>
        <Match when={status() === "awaitingPassword"}>
          <PasswordForm />
        </Match>
      </Switch>
    </div>
  );
}
