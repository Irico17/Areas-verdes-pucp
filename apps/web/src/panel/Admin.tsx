import { useEffect, useState, type FormEvent } from "react"
import { Esqueleto } from "../ui/Esqueleto"
import {
  actualizarCuenta,
  actualizarPermiso,
  actualizarRol,
  crearCuenta,
  crearRol,
  fetchCuentas,
  type Cuenta,
  type PermisoCuenta,
  type RolCuenta,
} from "../producto"

const ETIQUETAS: { id: string; label: string }[] = [
  { id: "capataz", label: "Capataz" },
  { id: "coordinacion", label: "Ingeniería / Coordinación" },
  { id: "jefatura", label: "Jefatura de sección" },
  { id: "admin", label: "Administrador del sistema" },
]

const ACCIONES = ["consultar", "registrar", "validar", "reportes", "catalogos", "solicitudes", "evidencias", "usuarios"]

function etiqueta(codigo: string, nombre?: string): string {
  const oficial = ETIQUETAS.find((item) => item.id === codigo)
  if (oficial) return oficial.label
  if (nombre && nombre.trim() !== "") return nombre
  return codigo
}

export function AdminPanel() {
  const [data, setData] = useState<Awaited<ReturnType<typeof fetchCuentas>> | null>(null)
  const [error, setError] = useState("")
  const [aviso, setAviso] = useState("")
  const [usuario, setUsuario] = useState("")
  const [nombre, setNombre] = useState("")
  const [rol, setRol] = useState("capataz")
  const [clave, setClave] = useState("")
  const [capatazId, setCapatazId] = useState("")
  const [codigoRol, setCodigoRol] = useState("")
  const [nombreRol, setNombreRol] = useState("")
  const [guardando, setGuardando] = useState(false)

  async function load() {
    const next = await fetchCuentas()
    setData(next)
    setError("")
  }

  useEffect(() => {
    let cancelled = false
    void fetchCuentas()
      .then((next) => {
        if (!cancelled) setData(next)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo leer las cuentas")
      })
    return () => {
      cancelled = true
    }
  }, [])

  async function correr(accion: () => Promise<void>, ok: string) {
    setGuardando(true)
    setError("")
    setAviso("")
    try {
      await accion()
      await load()
      setAviso(ok)
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo guardar")
    } finally {
      setGuardando(false)
    }
  }

  async function alta(event: FormEvent) {
    event.preventDefault()
    await correr(async () => {
      await crearCuenta({
        usuario: usuario.trim(),
        nombre: nombre.trim(),
        rol,
        clave,
        capataz_id: rol === "capataz" ? capatazId.trim() : "",
      })
      setUsuario("")
      setNombre("")
      setClave("")
      setCapatazId("")
    }, "Cuenta creada. La persona debe cambiar la clave al entrar.")
  }

  const roles = data?.roles ?? []
  const opciones = roles.length > 0 ? roles.filter((item) => item.activo) : ETIQUETAS.map((item) => ({ codigo: item.id, nombre: item.label, activo: true }))
  const acciones = accionesVisibles(data?.permisos ?? [])

  return (
    <section className="block">
      <h2>Admin</h2>
      <p className="lede">{data?.aviso || "Cuentas de VerdePUCP. Las administra la jefatura de sección."}</p>
      {error && <p className="status error">{error}</p>}
      {aviso && <p className="status">{aviso}</p>}
      {!data && !error && <Esqueleto />}

      <h3>Nueva cuenta</h3>
      <form className="form" onSubmit={(event) => void alta(event)}>
        <label className="field">
          Usuario
          <input value={usuario} onChange={(event) => setUsuario(event.target.value)} placeholder="camila.herrera" autoComplete="off" required />
        </label>
        <label className="field">
          Nombre
          <input value={nombre} onChange={(event) => setNombre(event.target.value)} placeholder="Camila Herrera" required />
        </label>
        <label className="field">
          Rol
          <select value={rol} onChange={(event) => setRol(event.target.value)}>
            {opciones.map((item) => (
              <option key={item.codigo} value={item.codigo}>
                {etiqueta(item.codigo, item.nombre)}
              </option>
            ))}
          </select>
        </label>
        {rol === "capataz" && (
          <label className="field">
            Cuadrilla
            <input value={capatazId} onChange={(event) => setCapatazId(event.target.value)} placeholder="cap-norte" />
          </label>
        )}
        <label className="field">
          Clave inicial
          <input type="password" value={clave} onChange={(event) => setClave(event.target.value)} minLength={10} autoComplete="new-password" required />
        </label>
        <button type="submit" className="primary" disabled={guardando}>
          Dar de alta
        </button>
      </form>

      <h3>Cuentas</h3>
      <ul className="labor-list">
        {(data?.usuarios ?? []).map((cuenta) => (
          <CuentaFila key={cuenta.usuario} cuenta={cuenta} roles={roles} guardando={guardando} onCorrer={correr} />
        ))}
      </ul>

      <h3>Roles</h3>
      <p className="hint">Un rol inactivo no inicia sesión. El código no se renombra. Los operarios no tienen cuenta.</p>
      <ul className="labor-list">
        {roles.map((item) => (
          <li key={item.codigo} className="agenda">
            <strong>{etiqueta(item.codigo, item.nombre)}</strong>
            <small>
              {item.codigo}
              {item.activo ? "" : " · inactivo"}
            </small>
            <button
              type="button"
              className="link"
              disabled={guardando}
              onClick={() => void correr(() => actualizarRol(item.codigo, !item.activo), item.activo ? "Rol desactivado." : "Rol activado.")}
            >
              {item.activo ? "Desactivar" : "Activar"}
            </button>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          void correr(async () => {
            await crearRol(codigoRol.trim(), nombreRol.trim())
            setCodigoRol("")
            setNombreRol("")
          }, "Rol creado.")
        }}
      >
        <label className="field">
          Código
          <input value={codigoRol} onChange={(event) => setCodigoRol(event.target.value)} placeholder="apoyo_campo" required />
        </label>
        <label className="field">
          Nombre visible
          <input value={nombreRol} onChange={(event) => setNombreRol(event.target.value)} placeholder="Apoyo de campo" required />
        </label>
        <button type="submit" className="primary" disabled={guardando}>
          Agregar rol
        </button>
      </form>

      <h3>Permisos</h3>
      <p className="hint">La matriz es provisional: el libro aún no la valida con el cliente. El punto de partida es la semilla actual. El cambio queda guardado y no pide reiniciar.</p>
      {roles.map((item) => (
        <fieldset key={item.codigo} className="checks">
          <legend>{etiqueta(item.codigo, item.nombre)}</legend>
          {acciones.map((accion) => {
            const marcado = (data?.permisos ?? []).some((permiso) => permiso.rol === item.codigo && permiso.accion === accion)
            return (
              <label key={accion}>
                <input
                  type="checkbox"
                  checked={marcado}
                  disabled={guardando}
                  onChange={(event) => void correr(() => actualizarPermiso(item.codigo, accion, event.target.checked), "Permiso actualizado.")}
                />
                {accion}
              </label>
            )
          })}
        </fieldset>
      ))}
    </section>
  )
}

function accionesVisibles(permisos: PermisoCuenta[]): string[] {
  const extra = permisos.map((item) => item.accion).filter((accion) => !ACCIONES.includes(accion))
  return [...ACCIONES, ...Array.from(new Set(extra))]
}

function CuentaFila(props: {
  cuenta: Cuenta
  roles: RolCuenta[]
  guardando: boolean
  onCorrer: (accion: () => Promise<void>, ok: string) => Promise<void>
}) {
  const { cuenta } = props
  const [rol, setRol] = useState(cuenta.rol)
  const [clave, setClave] = useState("")
  const opciones = props.roles.length > 0 ? props.roles : ETIQUETAS.map((item) => ({ codigo: item.id, nombre: item.label, activo: true }))

  return (
    <li className="agenda">
      <strong>
        {cuenta.nombre}
        {cuenta.activo ? "" : " · de baja"}
      </strong>
      <small>
        {cuenta.usuario} · {etiqueta(cuenta.rol, cuenta.rol_nombre)}
        {cuenta.capataz_id ? ` · ${cuenta.capataz_id}` : ""}
        {cuenta.debe_cambiar_password ? " · debe cambiar la clave" : ""}
      </small>
      <label className="field">
        Rol
        <select value={rol} onChange={(event) => setRol(event.target.value)}>
          {opciones.map((item) => (
            <option key={item.codigo} value={item.codigo}>
              {etiqueta(item.codigo, item.nombre)}
            </option>
          ))}
        </select>
      </label>
      <div className="row-actions">
        <button
          type="button"
          disabled={props.guardando || rol === cuenta.rol}
          onClick={() => void props.onCorrer(() => actualizarCuenta(cuenta.usuario, { rol }), "Rol actualizado.")}
        >
          Guardar rol
        </button>
        <button
          type="button"
          disabled={props.guardando}
          onClick={() => void props.onCorrer(() => actualizarCuenta(cuenta.usuario, { activo: !cuenta.activo }), cuenta.activo ? "Cuenta dada de baja." : "Cuenta reactivada.")}
        >
          {cuenta.activo ? "Dar de baja" : "Reactivar"}
        </button>
      </div>
      <label className="field">
        Nueva clave
        <input type="password" value={clave} onChange={(event) => setClave(event.target.value)} autoComplete="new-password" minLength={10} />
      </label>
      <button
        type="button"
        className="link"
        disabled={props.guardando || clave.length < 10}
        onClick={() =>
          void props.onCorrer(async () => {
            await actualizarCuenta(cuenta.usuario, { clave })
            setClave("")
          }, "Clave actualizada. La persona debe cambiarla al entrar.")
        }
      >
        Cambiar clave
      </button>
    </li>
  )
}
