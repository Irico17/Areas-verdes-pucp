---
name: frontend-design
description: Diseña y revisa la interfaz de Campus Verde (PUCP Pando), una PWA operativa de campo y oficina con el mapa como superficie principal. Úsala al crear o cambiar pantallas, chrome, tokens, tipografía o estados vacíos. Rechaza el aspecto genérico de plantilla.
---

# Diseño de interfaz — Campus Verde

Producto: supervisión de áreas verdes del campus PUCP Pando. Quien lo usa está de guardia o en el campo, no en un landing. El mapa es la superficie. El resto es instrumento.

## Qué no hacer

- Paleta por defecto crema + terracota / ocre (`#f3f0e8` + `#a8843d` y equivalentes). Ese par ya se usó y se retira.
- Rejillas de tarjetas idénticas, hero vacío, gradiente púrpura, kit Inter + violeta.
- Rótulos en versalitas con tracking ancho en cada sección (eyebrow spam). Una marca, no un sistema de cejas.
- Iconos de emoji, ilustraciones de stock, sombras suaves de marketing.
- Un segundo estilo para “verse moderno”. Densidad de herramienta, no de brochure.

## Qué sí hacer

- Mapa a sangre (MapLibre, OSM). El chrome es un riel estrecho y un panel del módulo activo.
- Una sola firma: el **riel de guardia** (barra vertical oscura con los módulos en castellano, sentencia, no mayúsculas).
- Tokens con nombre, pocos, alto contraste. El verde es dosel vivo, no decoración.
- Tipografía de lectura en campo: un sans de trabajo y, como mucho, una serif solo en el nombre del producto.
- Controles de 40 px en acciones primarias. Texto de interfaz desde 13 px. Estados vacío, carga y error en español, concretos.
- Color de estado solo en el dato (pendiente, en proceso, bloqueada), nunca como fondo de página.

## Al cerrar un cambio de UI

1. ¿El mapa sigue siendo lo primero que se ve en Mapa?
2. ¿Hay crema, ocre de marca, Inter o púrpura? Si sí, corregir.
3. ¿Los títulos de sección van en mayúsculas con tracking? Pasarlos a sentencia.
4. ¿Un módulo nuevo es otra tarjeta igual? Meterlo en el riel, no en un grid.
5. Contraste de texto sobre campo y sobre tinta. Probar el panel a ~390 px.
