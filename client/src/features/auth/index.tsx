import { Switch, Match } from "solid-js";
import { useAuth } from "./store";
import { PhoneForm } from "./components/PhoneForm";
import { CodeForm } from "./components/CodeForm";
import { PasswordForm } from "./components/PasswordForm";

export { useAuth } from "./store";

export function AuthFeature() {
  const { state } = useAuth();

  return (
    <div class="auth-page">
      <Switch>
        <Match when={state().status === "AWAITING_PHONE"}>
          <PhoneForm />
        </Match>
        <Match when={state().status === "AWAITING_CODE"}>
          <CodeForm />
        </Match>
        <Match when={state().status === "AWAITING_PASSWORD"}>
          <PasswordForm />
        </Match>
      </Switch>
    </div>
  );
}
