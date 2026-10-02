import { useState, type FormEvent } from "react"
import { entrar, type Usuario } from "../producto"

export function Login(props: { onIn: (usuario: Usuario) => void }) {
  const [usuario, setUsuario] = useState("coordinacion")
  const [clave, setClave] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  async function submit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      props.onIn(await entrar(usuario.trim(), clave))
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo entrar")
    } finally {
      setPending(false)
    }
  }

  return (
    <main className="gate">
      <section className="gate-brand">
        <img className="gate-logo" src="/logo-verdepucp.png" alt="VerdePUCP. Gestión de Áreas Verdes" />
        <p>Inicia sesión con tu cuenta de VerdePUCP.</p>
      </section>
      <section className="gate-form">
        <div className="gate-card">
          <h2>Entrar al turno</h2>
          <form onSubmit={(event) => void submit(event)}>
            <label className="field">
              Usuario
              <input value={usuario} onChange={(event) => setUsuario(event.target.value)} autoComplete="username" required />
            </label>
            <label className="field">
              Clave
              <input type="password" value={clave} onChange={(event) => setClave(event.target.value)} autoComplete="current-password" required />
            </label>
            {error && <p className="status error">{error}</p>}
            <button type="submit" className="primary" disabled={pending}>
              {pending ? "Entrando…" : "Entrar"}
            </button>
          </form>
          <p className="hint">
            Cuentas de esta instalación: norte, sur, riego, coordinacion, jefatura, admin. La clave es la de cada ambiente; pídasela al administrador.
          </p>
        </div>
      </section>
    </main>
  )
}
