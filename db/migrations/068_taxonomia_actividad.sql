-- Clase (primer nivel, ya cargada) y tipo (segundo nivel) con padre.
-- Riesgo: solo bajo y alto. «Medio» no se siembra hasta que el cliente lo confirme.

ALTER TABLE catalogos ADD COLUMN IF NOT EXISTS padre_codigo TEXT;

COMMENT ON COLUMN catalogos.padre_codigo IS
  'Código de la clase de actividad cuando el ítem es un tipo de segundo nivel. Nulo en el resto.';

INSERT INTO catalogos (clase, codigo, nombre, orden, activo, provisional) VALUES
  ('nivel_riesgo', 'bajo', 'Bajo', 1, TRUE, FALSE),
  ('nivel_riesgo', 'alto', 'Alto', 2, TRUE, FALSE)
ON CONFLICT (clase, codigo) DO NOTHING;

INSERT INTO catalogos (clase, codigo, nombre, orden, activo, provisional, padre_codigo) VALUES
  ('subtipo_actividad', 'preparacion_terreno', 'Preparación del terreno', 1, TRUE, FALSE, 'habilitacion'),
  ('subtipo_actividad', 'incorporacion_sustrato', 'Incorporación de sustrato', 2, TRUE, FALSE, 'habilitacion'),
  ('subtipo_actividad', 'instalacion_plantas', 'Instalación de plantas', 3, TRUE, FALSE, 'habilitacion'),
  ('subtipo_actividad', 'instalacion_cesped', 'Instalación de césped', 4, TRUE, FALSE, 'habilitacion'),
  ('subtipo_actividad', 'cobertura_ornamental', 'Colocación de cobertura ornamental', 5, TRUE, FALSE, 'habilitacion'),
  ('subtipo_actividad', 'instalacion_tutores', 'Instalación de tutores', 6, TRUE, FALSE, 'habilitacion'),
  ('subtipo_actividad', 'reposicion_plantas', 'Reposición de plantas', 1, TRUE, FALSE, 'rehabilitacion'),
  ('subtipo_actividad', 'recuperacion_areas', 'Recuperación de áreas verdes', 2, TRUE, FALSE, 'rehabilitacion'),
  ('subtipo_actividad', 'renovacion_jardineras', 'Renovación de jardineras', 3, TRUE, FALSE, 'rehabilitacion'),
  ('subtipo_actividad', 'mejoramiento_suelo', 'Mejoramiento del suelo', 4, TRUE, FALSE, 'rehabilitacion'),
  ('subtipo_actividad', 'renovacion_cobertura', 'Renovación de cobertura ornamental', 5, TRUE, FALSE, 'rehabilitacion'),
  ('subtipo_actividad', 'resiembra', 'Resiembra', 6, TRUE, FALSE, 'rehabilitacion'),
  ('subtipo_actividad', 'reubicacion_macetas', 'Reubicación de macetas', 7, TRUE, FALSE, 'rehabilitacion'),
  ('subtipo_actividad', 'canteo', 'Canteo', 1, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'deshierbo', 'Deshierbo', 2, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'escarda', 'Escarda', 3, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'limpieza_hojarasca', 'Limpieza de hojarasca', 4, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'limpieza_plantas', 'Limpieza de plantas', 5, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'aireacion_suelo', 'Aireación del suelo', 6, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'limpieza_jardineras', 'Limpieza integral de jardineras', 7, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'fertilizacion', 'Fertilización', 8, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'aplicacion_enmiendas', 'Aplicación de enmiendas', 9, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'traslado_macetas', 'Traslado de macetas', 10, TRUE, FALSE, 'mantenimiento'),
  ('subtipo_actividad', 'poda_mantenimiento', 'Poda de mantenimiento', 1, TRUE, FALSE, 'poda'),
  ('subtipo_actividad', 'poda_formacion', 'Poda de formación', 2, TRUE, FALSE, 'poda'),
  ('subtipo_actividad', 'poda_sanitaria', 'Poda sanitaria', 3, TRUE, FALSE, 'poda'),
  ('subtipo_actividad', 'poda_despeje', 'Poda de despeje/reducción', 4, TRUE, FALSE, 'poda'),
  ('subtipo_actividad', 'division_matas', 'Propagación por división de matas', 1, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'esquejes', 'Propagación por esquejes', 2, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'trasplante', 'Trasplante', 3, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'siembra_plantas', 'Siembra de plantas', 4, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'plantacion_arboles', 'Plantación de árboles', 5, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'plantacion_arbustos', 'Plantación de arbustos', 6, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'plantacion_cubresuelos', 'Plantación de cubresuelos', 7, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'plantacion_macetas', 'Plantación en macetas', 8, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'trasplante_macetas', 'Trasplante a macetas', 9, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'cambio_maceta', 'Cambio de maceta', 10, TRUE, FALSE, 'propagacion'),
  ('subtipo_actividad', 'riego_manual', 'Riego manual', 1, TRUE, FALSE, 'riego'),
  ('subtipo_actividad', 'riego_nocturno', 'Riego nocturno', 2, TRUE, FALSE, 'riego'),
  ('subtipo_actividad', 'riego_establecimiento', 'Riego de establecimiento', 3, TRUE, FALSE, 'riego'),
  ('subtipo_actividad', 'verificacion_riego', 'Verificación del sistema de riego', 4, TRUE, FALSE, 'riego'),
  ('subtipo_actividad', 'recoleccion_hojarasca', 'Recolección de hojarasca', 1, TRUE, FALSE, 'residuos'),
  ('subtipo_actividad', 'recoleccion_ramas', 'Recolección de ramas', 2, TRUE, FALSE, 'residuos'),
  ('subtipo_actividad', 'triturado_residuos', 'Triturado de residuos', 3, TRUE, FALSE, 'residuos'),
  ('subtipo_actividad', 'aprovechamiento_residuos', 'Disposición o aprovechamiento de residuos', 4, TRUE, FALSE, 'residuos')
ON CONFLICT (clase, codigo) DO NOTHING;
