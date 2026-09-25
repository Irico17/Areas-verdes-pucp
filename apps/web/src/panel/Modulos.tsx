import { useEffect, useState, type FormEvent } from "react"
import { FechaCampo } from "../FechaCampo"
import { formatFecha, hoyISO } from "../fecha"
import { ApiError, ESTADOS, etiquetaEstado, etiquetaTipo } from "../operacion"
import {
  crearAreaSinGeom,
  crearCatalogo,
  crearOrden,
  crearRiego,
  crearSolicitud,
  desactivarCatalogo,
  entrar,
  fetchCatalogo,
  fetchCuentas,
  fetchFichas,
  fetchOrdenes,
  fetchReporte,
  fetchRiego,
  fetchSolicitudes,
  guardarFicha,
  reporteHref,
  type CatalogoItem,
  type Ficha,
  type Orden,
  type Usuario,
} from "../producto"

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
        <p>La sesión de esta instalación es local: el SSO de la universidad no está conectado.</p>
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
            Cuentas de esta instalación: norte, sur, riego, coordinacion, jefatura, admin. Clave local: pando-local.
          </p>
        </div>
      </section>
    </main>
  )
}

function tituloFicha(row: Ficha): string {
  const nombre = row.nombre.trim()
  return nombre || row.feature_id
}

export function CatastroPanel() {
  const [q, setQ] = useState("")
  const [rows, setRows] = useState<Ficha[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [sel, setSel] = useState<Ficha | null>(null)
  const [nombreNuevo, setNombreNuevo] = useState("")
  const [usoNuevo, setUsoNuevo] = useState("")
  const [aviso, setAviso] = useState("")

  async function load(query = q) {
    setLoading(true)
    try {
      setRows(await fetchFichas(query))
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo leer el catastro")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let cancelled = false
    fetchFichas("")
      .then((rows) => {
        if (cancelled) return
        setRows(rows)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo leer el catastro")
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="split">
      <div className="split-list">
        <h2>Catastro</h2>
        <p className="lede">Las áreas con nombre van primero. Si el catastro no trae nombre, se muestra el código.</p>
        <form
          className="form"
          onSubmit={(event) => {
            event.preventDefault()
            void load(q)
          }}
        >
          <label className="field">
            Buscar
            <input value={q} onChange={(event) => setQ(event.target.value)} placeholder="Nombre, código o uso" />
          </label>
          <button type="submit">Buscar</button>
        </form>
        {error && <p className="status error">{error}</p>}
        {loading && (
          <div className="skel-wrap" aria-hidden="true">
            <div className="skel" />
            <div className="skel" />
            <div className="skel" />
          </div>
        )}
        {!loading && rows.length === 0 && !error && <p className="empty">Ningún área coincide con esa búsqueda.</p>}
        <ul className="labor-list">
          {rows.map((row) => (
            <li key={row.feature_id}>
              <button
                type="button"
                className={sel?.feature_id === row.feature_id ? "labor on" : "labor"}
                onClick={() => setSel(row)}
              >
                <span>
                  <strong>{tituloFicha(row)}</strong>
                  <small>
                    {row.nombre.trim() ? row.feature_id : "Sin nombre en el catastro"}
                    {row.uso ? ` · ${row.uso}` : ""}
                    {row.con_geometria ? "" : " · sin geometría"}
                  </small>
                </span>
              </button>
            </li>
          ))}
        </ul>
      </div>
      <div className="split-detail">
        {!sel && <p className="empty">Elija un área. La ficha queda en este panel, sin bajar por la lista.</p>}
        {sel && (
          <form
            className="form"
            onSubmit={(event) => {
              event.preventDefault()
              void guardarFicha(sel)
                .then(() => {
                  setAviso("Ficha guardada.")
                  setRows((current) => current.map((row) => (row.feature_id === sel.feature_id ? sel : row)))
                })
                .catch((err: unknown) => setAviso(err instanceof Error ? err.message : "No se pudo guardar"))
            }}
          >
            <h3>{tituloFicha(sel)}</h3>
            <label className="field">
              Nombre
              <input value={sel.nombre} onChange={(event) => setSel({ ...sel, nombre: event.target.value })} />
            </label>
            <label className="field">
              Uso
              <input value={sel.uso} onChange={(event) => setSel({ ...sel, uso: event.target.value })} />
            </label>
            <label className="field">
              Riego actual
              <input value={sel.riego_act} onChange={(event) => setSel({ ...sel, riego_act: event.target.value })} />
            </label>
            <label className="field">
              Referencia
              <input value={sel.referencia} onChange={(event) => setSel({ ...sel, referencia: event.target.value })} />
            </label>
            <button type="submit" className="primary">
              Guardar ficha
            </button>
          </form>
        )}
        <h3>Área sin GPS</h3>
        <form
          className="form"
          onSubmit={(event) => {
            event.preventDefault()
            void crearAreaSinGeom(nombreNuevo, usoNuevo)
              .then(() => {
                setNombreNuevo("")
                setUsoNuevo("")
                setAviso("Área creada sin geometría.")
                return load(q)
              })
              .catch((err: unknown) => setAviso(err instanceof Error ? err.message : "No se pudo crear"))
          }}
        >
          <label className="field">
            Nombre
            <input value={nombreNuevo} onChange={(event) => setNombreNuevo(event.target.value)} required />
          </label>
          <label className="field">
            Uso
            <input value={usoNuevo} onChange={(event) => setUsoNuevo(event.target.value)} />
          </label>
          <button type="submit">Registrar sin geometría</button>
        </form>
        {aviso && <p className={aviso.toLowerCase().includes("no se") ? "status error" : "banner"}>{aviso}</p>}
      </div>
    </section>
  )
}

export function SolicitudesPanel(props: { actividadId: string }) {
  const [rows, setRows] = useState<Awaited<ReturnType<typeof fetchSolicitudes>>>([])
  const [ordenes, setOrdenes] = useState<Orden[]>([])
  const [error, setError] = useState("")
  const [titulo, setTitulo] = useState("")
  const [fuente, setFuente] = useState("osg")
  const [codigo, setCodigo] = useState("")
  const [prioridad, setPrioridad] = useState("media")
  const [lugar, setLugar] = useState("")
  const [empresa, setEmpresa] = useState("")
  const [referencia, setReferencia] = useState("")

  async function load() {
    try {
      const [sol, ord] = await Promise.all([fetchSolicitudes(), fetchOrdenes()])
      setRows(sol)
      setOrdenes(ord)
      setError("")
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "No se pudieron leer las solicitudes")
    }
  }

  useEffect(() => {
    let cancelled = false
    Promise.all([fetchSolicitudes(), fetchOrdenes()])
      .then(([sol, ord]) => {
        if (cancelled) return
        setRows(sol)
        setOrdenes(ord)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : "No se pudieron leer las solicitudes")
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="block">
      <h2>Solicitudes</h2>
      <p className="lede">Captura manual. El código externo se conserva si viene de Centuria u OSG; el sistema no lo inventa.</p>
      {error && <p className="status error">{error}</p>}
      {rows.length === 0 && !error && <p className="empty">No hay solicitudes registradas.</p>}
      <ul className="labor-list">
        {rows.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.titulo}</strong>
            <small>
              {row.fuente} · {row.estado} · {row.prioridad}
              {row.codigo_externo ? ` · ${row.codigo_externo}` : " · sin código externo"}
            </small>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          void crearSolicitud({
            id: crypto.randomUUID(),
            titulo,
            fuente,
            codigo_externo: codigo,
            prioridad,
            lugar,
            detalle: "",
          })
            .then(() => {
              setTitulo("")
              setCodigo("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo crear"))
        }}
      >
        <label className="field">
          Título
          <input value={titulo} onChange={(event) => setTitulo(event.target.value)} required />
        </label>
        <label className="field">
          Fuente
          <select value={fuente} onChange={(event) => setFuente(event.target.value)}>
            <option value="centuria">Centuria</option>
            <option value="osg">Matriz OSG</option>
            <option value="correo">Correo</option>
            <option value="interna">Interna</option>
          </select>
        </label>
        <label className="field">
          Código externo
          <input value={codigo} onChange={(event) => setCodigo(event.target.value)} placeholder="Opcional" />
        </label>
        <label className="field">
          Prioridad
          <select value={prioridad} onChange={(event) => setPrioridad(event.target.value)}>
            <option value="baja">Baja</option>
            <option value="media">Media</option>
            <option value="alta">Alta</option>
          </select>
        </label>
        <label className="field">
          Lugar
          <input value={lugar} onChange={(event) => setLugar(event.target.value)} />
        </label>
        <button type="submit" className="primary">
          Registrar solicitud
        </button>
      </form>
      <h3>Órdenes de servicio</h3>
      <p className="lede">Solo se vinculan a una labor marcada como tercerizada. Sin orden, esa labor no se cierra.</p>
      {ordenes.length === 0 && <p className="empty">Todavía no hay órdenes. Una labor tercerizada no se cierra hasta que exista una.</p>}
      <ul className="labor-list">
        {ordenes.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.empresa}</strong>
            <small>
              {row.referencia} · {row.estado}
            </small>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          if (!props.actividadId) {
            setError("Elija una labor en el mapa antes de crear la orden.")
            return
          }
          void crearOrden({
            id: crypto.randomUUID(),
            actividad_id: props.actividadId,
            empresa,
            referencia,
            frecuencia: "",
          })
            .then(() => {
              setEmpresa("")
              setReferencia("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo crear la orden"))
        }}
      >
        <p className="hint">{props.actividadId ? `Labor seleccionada ${props.actividadId.slice(0, 8)}` : "Ninguna labor seleccionada."}</p>
        <label className="field">
          Empresa
          <input value={empresa} onChange={(event) => setEmpresa(event.target.value)} required />
        </label>
        <label className="field">
          Referencia de contratación
          <input value={referencia} onChange={(event) => setReferencia(event.target.value)} required />
        </label>
        <button type="submit">Registrar orden</button>
      </form>
    </section>
  )
}

export function ReportesPanel() {
  const [estado, setEstado] = useState("")
  const [desde, setDesde] = useState("")
  const [hasta, setHasta] = useState("")
  const [zona, setZona] = useState("")
  const [cuadrilla, setCuadrilla] = useState("")
  const [origen, setOrigen] = useState("")
  const [data, setData] = useState<Awaited<ReturnType<typeof fetchReporte>> | null>(null)
  const [error, setError] = useState("")
  const filtro = { estado, desde, hasta, zona, cuadrilla, origen }

  async function load() {
    try {
      setData(await fetchReporte(filtro))
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo armar el reporte")
    }
  }

  useEffect(() => {
    let cancelled = false
    fetchReporte({})
      .then((report) => {
        if (cancelled) return
        setData(report)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo armar el reporte")
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="block">
      <h2>Reportes</h2>
      <p className="lede">Reporte básico de labores, filtrable por zona, cuadrilla, origen y fechas.</p>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          void load()
        }}
      >
        <label className="field">
          Estado
          <select value={estado} onChange={(event) => setEstado(event.target.value)}>
            <option value="">Todos</option>
            {ESTADOS.map((item) => (
              <option key={item.id} value={item.id}>
                {item.label}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          Zona
          <input value={zona} onChange={(event) => setZona(event.target.value)} placeholder="Z1" />
        </label>
        <label className="field">
          Cuadrilla
          <input value={cuadrilla} onChange={(event) => setCuadrilla(event.target.value)} placeholder="Nombre ficticio" />
        </label>
        <label className="field">
          Origen
          <input value={origen} onChange={(event) => setOrigen(event.target.value)} placeholder="monitoreo" />
        </label>
        <label className="field">
          Desde
          <FechaCampo value={desde} onChange={setDesde} />
        </label>
        <label className="field">
          Hasta
          <FechaCampo value={hasta} onChange={setHasta} />
        </label>
        <button type="submit" className="primary">
          Actualizar
        </button>
      </form>
      {error && <p className="status error">{error}</p>}
      {data && (
        <>
          <p className="hint">{data.aviso}</p>
          <ul className="counts">
            {data.por_estado.map((row) => (
              <li key={row.estado}>
                <strong>{row.n}</strong>
                <span>{etiquetaEstado(row.estado)}</span>
              </li>
            ))}
          </ul>
          <p className="row-actions">
            <a href={reporteHref("csv", filtro)}>Descargar CSV</a>
            <a href={reporteHref("xls", filtro)}>Descargar Excel</a>
          </p>
          {data.filas.length === 0 && <p className="empty">No hay labores en ese rango.</p>}
          <ul className="labor-list">
            {data.filas.slice(0, 40).map((row) => (
              <li key={row.id} className="agenda">
                <strong>{row.titulo}</strong>
                <small>
                  {row.clase || etiquetaTipo(row.tipo)} · {etiquetaEstado(row.estado)}
                  {row.lugar ? ` · ${row.lugar}` : ""}
                  {row.cuadrilla ? ` · ${row.cuadrilla}` : ""}
                  {row.fecha_solicitud ? ` · ${row.fecha_solicitud}` : ""}
                </small>
              </li>
            ))}
          </ul>
        </>
      )}
    </section>
  )
}

const CLASES = [
  ["tipo_actividad", "Tipos de labor"],
  ["estado", "Estados"],
  ["prioridad", "Prioridades"],
  ["lugar", "Lugares"],
  ["especie", "Especies"],
  ["motivo_archivo", "Motivos de archivo"],
  ["turno", "Turnos"],
  ["fuente", "Fuentes"],
] as const

export function CatalogosPanel(props: { editable: boolean }) {
  const [clase, setClase] = useState<string>("tipo_actividad")
  const [items, setItems] = useState<CatalogoItem[]>([])
  const [error, setError] = useState("")
  const [codigo, setCodigo] = useState("")
  const [nombre, setNombre] = useState("")

  async function load(next = clase) {
    try {
      setItems(await fetchCatalogo(next, false))
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo leer el catálogo")
    }
  }

  useEffect(() => {
    let cancelled = false
    fetchCatalogo(clase, false)
      .then((next) => {
        if (cancelled) return
        setItems(next)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo leer el catálogo")
      })
    return () => {
      cancelled = true
    }
  }, [clase])

  return (
    <section className="block">
      <h2>Catálogos</h2>
      <p className="lede">Valores que usa la operación. Desactivar no borra el historial.</p>
      <label className="field">
        Clase
        <select value={clase} onChange={(event) => setClase(event.target.value)}>
          {CLASES.map(([id, label]) => (
            <option key={id} value={id}>
              {label}
            </option>
          ))}
        </select>
      </label>
      {error && <p className="status error">{error}</p>}
      {items.length === 0 && !error && <p className="empty">Esta clase no tiene ítems.</p>}
      <ul className="labor-list">
        {items.map((item) => (
          <li key={item.id} className="agenda">
            <strong>
              {item.nombre}
              {item.activo ? "" : " · inactivo"}
            </strong>
            <small>{item.codigo}</small>
            {props.editable && item.activo && (
              <button type="button" className="link" onClick={() => void desactivarCatalogo(item.id).then(() => load())}>
                Desactivar
              </button>
            )}
          </li>
        ))}
      </ul>
      {props.editable && (
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          void crearCatalogo(clase, codigo, nombre)
            .then(() => {
              setCodigo("")
              setNombre("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo guardar"))
        }}
      >
        <label className="field">
          Código
          <input value={codigo} onChange={(event) => setCodigo(event.target.value)} placeholder="minusculas_sin_espacio" required />
        </label>
        <label className="field">
          Nombre
          <input value={nombre} onChange={(event) => setNombre(event.target.value)} required />
        </label>
        <button type="submit" className="primary">
          Agregar
        </button>
      </form>
      )}
      {!props.editable && <p className="hint">Puede consultar el catálogo. Solo administración lo modifica.</p>}
    </section>
  )
}

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
      <p className="lede">{data?.aviso || "Cuentas locales. El SSO institucional no forma parte de este piloto."}</p>
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
              {user.usuario} · {user.rol}
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
              {item.rol} · {item.accion}
            </small>
          </li>
        ))}
      </ul>
    </section>
  )
}

export function RiegoPanel(props: { capatazId: string }) {
  const [aviso, setAviso] = useState("")
  const [rows, setRows] = useState<Awaited<ReturnType<typeof fetchRiego>>["registros"]>([])
  const [error, setError] = useState("")
  const [sector, setSector] = useState("Eje central")
  const [zona, setZona] = useState("Z1")
  const [turno, setTurno] = useState("manana")
  const [fecha, setFecha] = useState(hoyISO)
  const [nota, setNota] = useState("")

  async function load() {
    try {
      const body = await fetchRiego()
      setAviso(body.aviso)
      setRows(body.registros)
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo leer el riego")
    }
  }

  useEffect(() => {
    let cancelled = false
    fetchRiego()
      .then((body) => {
        if (cancelled) return
        setAviso(body.aviso)
        setRows(body.registros)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo leer el riego")
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="block">
      <h2>Riego</h2>
      <p className="lede">{aviso || "Sector, turno y equipo. Sin porcentaje oficial de cobertura."}</p>
      {error && <p className="status error">{error}</p>}
      {rows.length === 0 && !error && <p className="empty">Todavía no hay turnos registrados.</p>}
      <ul className="labor-list">
        {rows.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.sector}</strong>
            <small>
              {formatFecha(row.fecha)} · {row.turno === "manana" ? "mañana" : row.turno} · {row.equipo || "sin equipo"}
            </small>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          void crearRiego({
            id: crypto.randomUUID(),
            sector,
            turno,
            capataz_id: props.capatazId,
            fecha,
            nota,
            zona_supervision_id: zona,
          })
            .then(() => {
              setNota("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo registrar"))
        }}
      >
        <label className="field">
          Zona de supervisión
          <select value={zona} onChange={(event) => setZona(event.target.value)}>
            <option value="Z1">Z1</option>
            <option value="Z2">Z2</option>
            <option value="Z3">Z3</option>
            <option value="Z4">Z4</option>
          </select>
        </label>
        <label className="field">
          Sector
          <input value={sector} onChange={(event) => setSector(event.target.value)} required />
        </label>
        <label className="field">
          Turno
          <select value={turno} onChange={(event) => setTurno(event.target.value)}>
            <option value="manana">Mañana</option>
            <option value="tarde">Tarde</option>
          </select>
        </label>
        <label className="field">
          Fecha
          <FechaCampo value={fecha} onChange={setFecha} required />
        </label>
        <label className="field">
          Nota
          <input value={nota} onChange={(event) => setNota(event.target.value)} />
        </label>
        <button type="submit" className="primary">
          Registrar turno
        </button>
      </form>
    </section>
  )
}
