---
name: VerdePUCP
description: Gestión de Áreas Verdes. Mesa de guardia del campus PUCP Pando.
colors:
  tinta: "#102033"
  papel: "#e7eeeb"
  hoja: "#f7faf8"
  linea: "#c5d4ce"
  verde: "#308046"
  marino: "#083465"
  yucca: "#308046"
  yucca-deep: "#083465"
  yucca-ink: "#f4faf6"
  alerta: "#8c3832"
  pendiente: "#8a6410"
  muted: "#3e5148"
  cerrado: "#5c6a62"
typography:
  display:
    fontFamily: "Newsreader, Georgia, serif"
    fontSize: "56px"
    fontWeight: 500
    lineHeight: 1.02
    letterSpacing: "-0.03em"
  headline:
    fontFamily: "Familjen Grotesk, sans-serif"
    fontSize: "22px"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "-0.02em"
  title:
    fontFamily: "Familjen Grotesk, sans-serif"
    fontSize: "16px"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.01em"
  body:
    fontFamily: "Familjen Grotesk, sans-serif"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.45
    letterSpacing: "0"
  label:
    fontFamily: "Familjen Grotesk, sans-serif"
    fontSize: "12px"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "0.01em"
rounded:
  control: "6px"
  sheet: "10px"
  mark: "4px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "12px"
  lg: "16px"
  xl: "24px"
  xxl: "40px"
components:
  button-primary:
    backgroundColor: "{colors.yucca}"
    textColor: "{colors.yucca-ink}"
    rounded: "{rounded.control}"
    padding: "10px 16px"
  button-primary-hover:
    backgroundColor: "{colors.yucca-deep}"
    textColor: "{colors.yucca-ink}"
    rounded: "{rounded.control}"
    padding: "10px 16px"
  button-quiet:
    backgroundColor: "{colors.hoja}"
    textColor: "{colors.tinta}"
    rounded: "{rounded.control}"
    padding: "10px 14px"
  field:
    backgroundColor: "{colors.hoja}"
    textColor: "{colors.tinta}"
    rounded: "{rounded.control}"
    padding: "10px 12px"
---

# Design System: VerdePUCP

## Overview

**Creative North Star: "La mesa de guardia"**

El sistema vigente (riel negro de 76 px, Source Sans 3, formulario de acceso en la esquina, todo en esquina viva y sin sombra) queda como anti-referencia. Esta ficha es el mundo de reemplazo, modo Operate: una herramienta de turno, no una portada.

La mesa es papel verde grisáceo, tinta casi negra con sesgo verde, y un solo acento: yucca, el verde de una hoja gruesa, no el esmeralda de un tablero SaaS. Newsreader aparece una sola vez, en el nombre del producto. El resto es Familjen Grotesk, con cifras tabulares. La firma es la franja de guardia del acceso: un plano de yucca a la izquierda, el formulario centrado en el papel a la derecha. Dentro de la app, el mapa ocupa el escenario y el panel es una hoja con cabeza fija.

**Key Characteristics:**

- Un acento. Yucca para la acción principal, la selección y el estado en proceso. Alerta y pendiente son semántica, no decoración.
- Densidad de oficina: listas compactas, aire entre bloques, nunca entre la etiqueta y su campo.
- El mapa no se enmarca como tarjeta. Los controles flotan sobre él con la misma hoja del panel.
- Sin ceja en mayúsculas, sin gradiente de texto, sin tarjeta genérica de icono + título + párrafo.

## Colors

Papel frío con tinta verde. El acento se gasta poco.

### Primary

- **Yucca** (#1a5c44): botón primario, módulo activo, labor en proceso, marca del tipo cuando no hay un estado más urgente.
- **Yucca profundo** (#0e3b2c): hover y presión del primario. También el plano de la franja de acceso.
- **Tinta de yucca** (#f3faf6): texto sobre yucca.

### Neutral

- **Tinta** (#1a2420): texto.
- **Papel** (#e4ebe4): fondo de la aplicación.
- **Hoja** (#f7faf6): paneles, barra, campos.
- **Línea** (#c5d2c8): cortes y bordes de 1 px.
- **Apagado** (#3e5148): texto secundario. Nunca un gris neutro sobre el papel.

### Semantic

- **Alerta** (#8c3832): error, labor bloqueada, acción destructiva.
- **Pendiente** (#8a6410): labor pendiente.
- **Cerrado** (#5c6a62): labor cerrada o cancelada.

**The One Voice Rule.** Yucca pinta la acción y la selección. No pinta fondos de sección ni iconos decorativos.

## Typography

**Display Font:** Newsreader (Georgia)
**Body Font:** Familjen Grotesk (sans-serif del sistema)
**Label/Mono:** la misma grotesca, con `font-variant-numeric: tabular-nums` en conteos, coordenadas y fechas.

**Character:** el nombre del producto es un título de plano; la interfaz es una nota de campo. No hay una tercera familia.

### Hierarchy

- **Display** (500, 56 px, 1.02): solo el nombre en el acceso. En la barra de la app baja a 22 px.
- **Headline** (600, 22 px, 1.2): título del panel.
- **Title** (600, 16 px, 1.3): nombre de labor o de área.
- **Body** (400, 15 px, 1.45): texto, tope ~65 ch en párrafos.
- **Label** (600, 12 px, 1.3): etiquetas de campo, en oración, no en versalitas.

**The No Eyebrow Rule.** No hay una línea pequeña en mayúsculas encima del título. El título habla solo.

## Layout

Escritorio: barra de 56 px, luego una grilla `188 px | 400 px | mapa`. El panel tiene cabeza fija y cuerpo con scroll. En Catastro el cuerpo se parte: la lista ocupa como máximo el 46 % y la ficha queda visible debajo, con su propio scroll. En el teléfono (≤820 px) la navegación es una fila horizontal, el mapa es el escenario y el panel entra como hoja desde la izquierda, con la identidad y Salir visibles en la barra.

El acceso no usa esa grilla. Es `minmax(280px, 42%) | 1fr` a 1440 px, y en el teléfono la franja de yucca pasa a ser una banda de 168 px arriba del formulario.

## Elevation & Depth

Las hojas de trabajo son planas: las separa el tono (papel contra hoja) y una línea de 1 px. La sombra aparece solo cuando algo flota sobre el mapa (barra de modo, popup, hoja móvil) o en la franja de acceso, donde el formulario no lleva tarjeta: el plano de yucca ya es el contraste.

### Shadow Vocabulary

- **Flotante** (`box-shadow: 0 10px 28px -16px rgba(26, 36, 32, 0.45)`): controles sobre el mapa y la hoja móvil.
- Nada más. Sin halo de color a offset cero. Sin sombra dura.

**The Flat Desk Rule.** Una lista en reposo no tiene sombra. La elevación significa “esto está sobre el mapa”.

## Shapes

Controles a 6 px, hojas flotantes a 10 px, la marca de estado (la letra del tipo) a 4 px. Nada es píldora salvo el punto de un pin. La franja de acceso es un rectángulo a sangre, sin radio.

## Components

- **Botón primario:** yucca, texto claro, altura mínima 44 px. Hover a yucca profundo. Presión `scale(0.98)`. Deshabilitado baja la opacidad a 0.45 y no responde.
- **Botón quieto:** hoja, línea de 1 px, tinta. El peligro usa texto alerta y línea alerta, sin relleno rojo.
- **Campo:** hoja, línea, radio 6 px, etiqueta encima en 12 px. El foco es un anillo de 2 px yucca con 2 px de separación.
- **Fecha:** texto `dd/mm/aaaa`, nunca `input type="date"`. El valor que viaja a la API sigue siendo `AAAA-MM-DD`.
- **Lista de labor:** fila de una línea, marca de 4 px con la letra, título y meta. La fila activa lleva fondo yucca al 8 %, no una barra lateral.
- **Vacío:** una frase que dice qué falta y la acción que lo resuelve, dentro del mismo panel, sin ilustración de stock.
- **Carga:** esqueleto de la lista (tres filas) en lugar de un texto suelto “cargando” en medio del mapa. El mapa puede decir su estado en la barra.

## Do's and Don'ts

- Sí: español del Perú, fechas `dd/mm/aaaa` con `es-PE` y `America/Lima`, códigos de área cuando no hay nombre.
- Sí: un solo acento, cifras tabulares, foco visible, `prefers-reduced-motion` que anula el desplazamiento y deja el cambio de color.
- No: Inter, Outfit, Geist, gradiente púrpura, crema con terracota, riel negro de disfraz, formulario de acceso en la esquina, “Sin nombre” como título de una lista entera.
- No: ceja en mayúsculas, texto con gradiente, tarjeta anidada, emoji como icono, `mm/dd/yyyy`.
