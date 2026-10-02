import { useEffect, useState, type FormEvent } from "react"
import { cambiarClavePropia, entrar, leerSesion, type Usuario } from "../producto"

export function Login(props: { onIn: (usuario: Usuario) => void }) {
  const [usuario, setUsuario] = useState("coordinacion")
  const [clave, setClave] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [pendienteClave, setPendienteClave] = useState<Usuario | null>(null)
  const [claveActual, setClaveActual] = useState("")
  const [claveNueva, setClaveNueva] = useState("")

  useEffect(() => {
    let cancelled = false
    void leerSesion()
      .then((sesion) => {
        if (cancelled || !sesion?.debe_cambiar_password) return
        setPendienteClave(sesion)
        setClaveActual("")
      })
      .catch(() => {
        /* el formulario de entrada sigue disponible */
      })
    return () => {
      cancelled = true
    }
  }, [])

  async function submit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      const sesion = await entrar(usuario.trim(), clave)
      if (sesion.debe_cambiar_password) {
        setPendienteClave(sesion)
        setClaveActual(clave)
        setClaveNueva("")
        return
      }
      props.onIn(sesion)
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo entrar")
    } finally {
      setPending(false)
    }
  }

  async function cambiar(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      const sesion = await cambiarClavePropia(claveActual, claveNueva)
      props.onIn({ ...sesion, debe_cambiar_password: false })
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo cambiar la clave")
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
          {pendienteClave ? (
            <>
              <h2>Elija su clave</h2>
              <p className="lede">La jefatura entregó una clave inicial. Cámbiela antes de entrar al mapa.</p>
              <form onSubmit={(event) => void cambiar(event)}>
                <label className="field">
                  Clave actual
                  <input
                    type="password"
                    value={claveActual}
                    onChange={(event) => setClaveActual(event.target.value)}
                    autoComplete="current-password"
                    required
                  />
                </label>
                <label className="field">
                  Clave nueva
                  <input
                    type="password"
                    value={claveNueva}
                    onChange={(event) => setClaveNueva(event.target.value)}
                    autoComplete="new-password"
                    minLength={10}
                    required
                  />
                </label>
                {error && <p className="status error">{error}</p>}
                <button type="submit" className="primary" disabled={pending}>
                  {pending ? "Guardando…" : "Guardar clave"}
                </button>
              </form>
            </>
          ) : (
            <>
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
                Cuentas de esta instalación: norte, sur, riego, coordinacion, jefatura, admin. La clave de cada persona la entrega la jefatura de sección.
              </p>
            </>
          )}
        </div>
      </section>
    </main>
  )
}
