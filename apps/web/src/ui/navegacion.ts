/** En el teléfono caben cinco casillas: cuatro módulos y «Más» con el resto. */
export function repartirModulos<T>(modulos: readonly T[], casillas = 5): { barra: T[]; mas: T[] } {
  if (modulos.length <= casillas) return { barra: [...modulos], mas: [] }
  return { barra: modulos.slice(0, casillas - 1), mas: modulos.slice(casillas - 1) }
}
