import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { registerSW } from "virtual:pwa-register"
import "@fontsource/familjen-grotesk/400.css"
import "@fontsource/familjen-grotesk/500.css"
import "@fontsource/familjen-grotesk/600.css"
import "@fontsource/newsreader/500.css"
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
