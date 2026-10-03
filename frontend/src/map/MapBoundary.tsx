import { Component, type ReactNode } from "react"

type Props = { children: ReactNode }
type State = { message: string | null }

export class MapBoundary extends Component<Props, State> {
  state: State = { message: null }

  static getDerivedStateFromError(error: unknown): State {
    const message = error instanceof Error ? error.message : "No se pudo dibujar el mapa"
    return { message }
  }

  render() {
    if (this.state.message) {
      return (
        <div className="map-fallback">
          <p>El mapa no arrancó en este navegador.</p>
          <p>{this.state.message}</p>
        </div>
      )
    }
    return this.props.children
  }
}
