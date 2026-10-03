export function Esqueleto({ filas = 3 }: { filas?: number }) {
  return (
    <div className="skel-wrap" aria-hidden="true">
      {Array.from({ length: filas }, (_, i) => (
        <div className="skel" key={i} />
      ))}
    </div>
  )
}
