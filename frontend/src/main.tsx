import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { registerSW } from "virtual:pwa-register"
import "@fontsource/familjen-grotesk/latin-400"
import "@fontsource/familjen-grotesk/latin-500"
import "@fontsource/familjen-grotesk/latin-600"
import "@fontsource/newsreader/latin-500"
import App from "./App"
import "./styles.css"

if (import.meta.env.PROD) {
  registerSW({ immediate: true })
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
