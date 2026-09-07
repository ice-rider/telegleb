import { render } from "solid-js/web";
import App from "./App";
import { installGlobalErrorLogging, log } from "./core/log";
import "./index.css";

installGlobalErrorLogging();

const root = document.getElementById("root");
if (!root) {
  // Раньше приложение в этом случае просто не запускалось молча.
  log.error("boot", "элемент #root не найден, приложение не смонтировано");
} else {
  render(() => <App />, root);
}
