import { useEffect, useState } from "react"
import { etiquetaRol, fetchCuentas } from "../producto"

export function AdminPanel() {
  const [data, setData] = useState<Awaited<ReturnType<typeof fetchCuentas>> | null>(null)
  const [error, setError] = useState("")

  useEffect(() => {
    void fetchCuentas()
      .then(setData)
      .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo leer las cuentas"))
  }, [])

  return (
    <section className="block">
      <h2>Admin</h2>
      <p className="lede">Cuentas de VerdePUCP. Las administra la jefatura de sección.</p>
      {error && <p className="status error">{error}</p>}
      {!data && !error && (
        <div className="skel-wrap" aria-hidden="true">
          <div className="skel" />
          <div className="skel" />
          <div className="skel" />
        </div>
      )}
      <ul className="labor-list">
        {(data?.usuarios ?? []).map((user) => (
          <li key={user.id} className="agenda">
            <strong>{user.nombre}</strong>
            <small>
              {user.usuario} · {etiquetaRol(user.rol, user.rol_nombre)}
              {user.capataz_id ? ` · ${user.capataz_id}` : ""}
            </small>
          </li>
        ))}
      </ul>
      <h3>Permisos semilla</h3>
      <ul className="labor-list">
        {(data?.permisos ?? []).map((item) => (
          <li key={`${item.rol}-${item.accion}`} className="agenda">
            <small>
              {etiquetaRol(item.rol)} · {item.accion}
            </small>
          </li>
        ))}
      </ul>
    </section>
  )
}
