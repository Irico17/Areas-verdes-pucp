// ==========================================================================
// 1. CONFIGURACIÓN GENERAL Y ENLACES
// ==========================================================================
var urlPuntosPUCPCSV = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTb8rgvlJwdua0hQZbAaALgJYAbK4EkFmYdmMtJhuyZ5rT7fuHrB5zfPQY1qxa0LQ/pub?gid=1330096436&single=true&output=csv";
var datosJardinesReservaGeoJSON = null;

// [CORRECCIÓN] Declaramos la función al inicio de todo para que sea visible globalmente
function transformarEnlaceDrive(url) {
    if (!url) return "";
    if (typeof url !== 'string') return "";
    if (url.includes("wikimedia.org") || url.includes("upload.wikimedia")) return url;
    
    if (url.includes("drive.google.com") || url.includes("docs.google.com")) {
        var matches = url.match(/\/d\/([a-zA-Z0-9-_]+)/) || url.match(/id=([a-zA-Z0-9-_]+)/);
        if (matches && matches[1]) {
            return "https://drive.google.com/thumbnail?id=" + matches[1] + "&sz=w500";
        }
    }
    return url;
}


// ==========================================================================
// 2. INICIALIZACIÓN INMEDIATA DEL MAPA Y CAPAS
// ==========================================================================
// var map = L.map('map').setView([-12.068, -77.08], 15);

// Capa de satélite por defecto (Google Maps)
//var googleSatelite = L.tileLayer('https://mt{s}.google.com/vt/lyrs=y&x={x}&y={y}&z={z}', {
    //subdomains: ['0', '1', '2', '3'],
    //vmaxZoom: 20,
   // attribution: '© Google Maps'
//}).addTo(map);

// Capas alternativas secundarias
//var mapaNormal = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
   // maxZoom: 19,
   // attribution: '© OpenStreetMap'
//}); 

//var esriSatelite = L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}', {
   // maxZoom: 18,
   // attribution: '© Esri'
//});

// Inicializamos las capas de forma estándar
//var capaPuntosPUCP = L.layerGroup().addTo(map); 
//var capaReservasJardines = L.layerGroup().addTo(map);//

// ==========================================================================
// 2. INICIALIZACIÓN INMEDIATA DEL MAPA Y CAPAS (CON SUPERZOOM HABILITADO
// ==========================================================================

// 1. Agregamos maxZoom: 22 al mapa base
var map = L.map('map', {
    maxZoom: 22 // Permite al usuario hacer zoom hasta nivel 22
}).setView([-12.068, -77.08], 15);

// Capa de satélite por defecto (Google Maps)
var googleSatelite = L.tileLayer('https://mt{s}.google.com/vt/lyrs=y&x={x}&y={y}&z={z}', {
    subdomains: ['0', '1', '2', '3'],
    maxNativeZoom: 20, // Nivel máximo de imágenes reales que tiene Google
    maxZoom: 22,       // Forzamos a estirar digitalmente la imagen más allá del límite
    attribution: '© Google Maps'
}).addTo(map);

// Capas alternativas secundarias
var mapaNormal = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxNativeZoom: 19,
    maxZoom: 22,
    attribution: '© OpenStreetMap'
}); 

var esriSatelite = L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}', {
    maxNativeZoom: 18,
    maxZoom: 22,
    attribution: '© Esri'
});

// Inicializamos las capas de forma estándar
var capaPuntosPUCP = L.layerGroup().addTo(map); 
var capaReservasJardines = L.layerGroup().addTo(map);

// ==========================================================================
// 3. CARGA DE DATOS EN MEMORIA Y FILTRADO DINÁMICO
// ==========================================================================
var bancoDatosPUCP = []; // Guardará los datos sin pintarlos de golpe

if (typeof Papa !== 'undefined') {
    console.log("⏳ Conectando con la base de datos de Google Sheets...");
    
    var bloqueTiempoPUCP = Math.floor(new Date().getTime() / 300000); 
    var urlConCache = urlPuntosPUCPCSV + (urlPuntosPUCPCSV.includes('?') ? '&' : '?') + "cachebuster=" + bloqueTiempoPUCP;
    
    Papa.parse(urlConCache, {
        download: true,
        header: true,
        skipEmptyLines: true,
        complete: function(resultados) {
            bancoDatosPUCP = resultados.data;
            console.log("📊 Datos cargados en memoria. Total de lugares: " + bancoDatosPUCP.length);
            
            // Centramos el mapa limpio sin marcadores
            if (typeof map !== 'undefined') {
                map.invalidateSize(); 
                map.setView([-12.0685, -77.0815], 16); 
            }
        },
        error: function(err) {
            console.error("❌ Error en PapaParse: ", err);
        }
    });
}

// ==========================================================================
// FUNCIÓN DE BÚSQUEDA Y RENDEREADO DINÁMICO
// ==========================================================================
function buscarYLimpiarMapa(textoBusqueda) {
    if (typeof capaPuntosPUCP === 'undefined' || typeof map === 'undefined') return;

    // 1. Limpiamos la capa
    capaPuntosPUCP.clearLayers();

    var query = textoBusqueda.trim().toLowerCase();

    // 2. Si la búsqueda está vacía, no mostramos nada
    if (query === "") {
        map.setView([-12.0685, -77.0815], 16); // Regresa a la vista general del campus
        return;
    }

    var puntosEncontrados = [];

    // 3. Filtrar y pintar solo los que coincidan
    bancoDatosPUCP.forEach(function(fila) {
        var titulo = fila["title"] || fila["searchString"] || "Lugar PUCP";
        
        if (titulo.toLowerCase().includes(query)) {
            var latRaw = fila["location/lat"] || fila["latitud"] || fila["lat"] || "";
            var lngRaw = fila["location/lng"] || fila["longitud"] || fila["lng"] || "";

            if (latRaw && lngRaw) {
                var lat = parseFloat(latRaw.toString().trim().replace(/,/g, "."));
                var lng = parseFloat(lngRaw.toString().trim().replace(/,/g, "."));

                if (!isNaN(lat) && !isNaN(lng)) {
                    var telefono = fila["phone"] || "No disponible";
                    var urlSitio = fila["url"] || fila["searchPageUrl"] || "";
                    var fotoRaw = fila["imageOfPlace"] || fila["image"] || "";
                    var fotoUrl = typeof transformarEnlaceDrive === 'function' ? transformarEnlaceDrive(fotoRaw) : fotoRaw;

                    var iconoHtml = L.divIcon({
                        className: 'marcador-forzado-visible', 
                        html: `<div style="
                            width: 20px !important; 
                            height: 20px !important; 
                            background-color: #ff2a2a !important; 
                            border: 3px solid #ffffff !important; 
                            border-radius: 50% !important; 
                            box-shadow: 0 0 8px rgba(0,0,0,0.6) !important;
                            cursor: pointer !important;
                        "></div>`,
                        iconSize: [20, 20],
                        iconAnchor: [10, 10] 
                    });

                    var marcador = L.marker([lat, lng], { icon: iconoHtml });

                    var contenidoPopup = `
                        <div style="width: 240px; font-family: sans-serif; font-size: 12px; line-height: 1.4;">
                            <h4 style="margin: 0 0 6px 0; color: #1a73e8; font-size: 14px; font-weight: bold; border-bottom: 1px solid #eee; padding-bottom: 4px;">${titulo}</h4>
                    `;
                    
                    if (fotoUrl && fotoUrl.startsWith("http")) {
                        contenidoPopup += `
                            <div style="width: 100%; height: 120px; overflow: hidden; border-radius: 4px; background-color: #f0f0f0; margin-bottom: 6px;">
                                <img src="${fotoUrl}" style="width: 100%; height: 100%; object-fit: cover;" alt="${titulo}">
                            </div>
                        `;
                    }
                    
                    contenidoPopup += `<p style="margin: 4px 0;"><strong>📞 Teléfono:</strong> ${telefono}</p>`;
                    
                    if (urlSitio) {
                        contenidoPopup += `
                            <p style="margin: 6px 0 0 0; text-align: right;">
                                <a href="${urlSitio}" target="_blank" style="color: #1a73e8; text-decoration: none; font-weight: bold; font-size: 11px;">Ver en Google Maps ➔</a>
                            </p>
                        `;
                    }
                    contenidoPopup += `</div>`;

                    marcador.bindPopup(contenidoPopup, { maxWidth: 260 });
                    
                    capaPuntosPUCP.addLayer(marcador);
                    puntosEncontrados.push([lat, lng]);
                }
            }
        }
    });

    // Aseguramos que la capa esté agregada al mapa
    if (!map.hasLayer(capaPuntosPUCP)) {
        capaPuntosPUCP.addTo(map);
    }

    // 4. Enfocar la cámara según los resultados
    if (puntosEncontrados.length === 1) {
        map.setView(puntosEncontrados[0], 18);
    } else if (puntosEncontrados.length > 1) {
        var bounds = L.latLngBounds(puntosEncontrados);
        map.fitBounds(bounds, { padding: [50, 50] });
    }
}

// Listener para el campo de búsqueda
document.addEventListener("DOMContentLoaded", function() {
    var inputBuscador = document.getElementById("inputBuscador");
    if (inputBuscador) {
        inputBuscador.addEventListener("input", function(e) {
            buscarYLimpiarMapa(e.target.value);
        });
    }
});

// ==========================================================================
// 2. Panel Flotante Izquierdo de Filtros (EMPIEZA CERRADO POR DEFECTO)
// ==========================================================================
var MenuFiltrosFlotante = L.Control.extend({
    options: { position: 'topleft' }, 
    onAdd: function (map) {
        var div = L.DomUtil.create('div', '');
        div.innerHTML = `
            <div style="background: white; padding: 12px 15px; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.15); margin: 10px; min-width: 230px; font-family: Arial, sans-serif;">
                
                <!-- Cabecera Plegable (Click para abrir/cerrar) -->
                <div id="btn-toggle-filtros" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;">
                    <h3 style="margin: 0; font-size: 14px; color: #2c3e50; font-weight: bold;">🔍 Filtros de Actividades</h3>
                    <!-- Flecha rotada indicando estado cerrado por defecto -->
                    <span id="flecha-filtros" style="font-size: 12px; color: #34495e; transition: transform 0.3s ease; display: inline-block; transform: rotate(-90deg);">▼</span>
                </div>

                <!-- Contenido del Panel (Cerrado por defecto con display: none) -->
                <div id="cuerpo-filtros" style="margin-top: 10px; border-top: 2px solid #34495e; padding-top: 10px; display: none;">
                    
                    <!-- 🌿 Sección: Opciones de visualización -->
                    <div style="margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px dashed #e2e8f0;">
                        <div style="font-size: 11px; font-weight: bold; color: #2c3e50; margin-bottom: 5px; text-transform: uppercase;">
                            🌿 ACTIVIDADES DE ÁREAS VERDES
                        </div>
                        
                        <label style="display: flex; align-items: center; font-size: 12.5px; color: #475569; cursor: pointer; padding: 4px 0;">
                            <input type="checkbox" id="chk-actividades" style="margin-right: 8px; accent-color: #2563eb;"> 📌 Ver Marcadores
                        </label>

                        <label style="display: flex; align-items: center; font-size: 12.5px; color: #475569; cursor: pointer; padding: 4px 0;">
                            <input type="checkbox" id="chk-calor-actividades" style="margin-right: 8px; accent-color: #e74c3c;"> 🔥 Ver Mapa de Calor
                        </label>
                    </div>

                    <!-- Selectores de Filtros -->
                    <div style="margin-bottom: 10px;">
                        <label style="display: block; font-size: 11px; font-weight: bold; margin-bottom: 3px; color: #7f8c8d;" for="mes">Mes:</label>
                        <select style="width: 100%; padding: 6px; border-radius: 4px; border: 1px solid #ccc; font-size: 12px;" id="mes">
                            <option value="todos">Todos los meses</option>
                            <option value="enero">Enero</option>
                            <option value="febrero">Febrero</option>
                            <option value="marzo">Marzo</option>
                            <option value="abril">Abril</option>
                            <option value="mayo">Mayo</option>
                            <option value="junio">Junio</option>
                            <option value="julio">Julio</option>
                            <option value="agosto">Agosto</option>
                            <option value="septiembre">Septiembre</option>
                            <option value="octubre">Octubre</option>
                            <option value="noviembre">Noviembre</option>
                            <option value="diciembre">Diciembre</option>
                        </select>
                    </div>

                    <div style="margin-bottom: 10px;">
                        <label style="display: block; font-size: 11px; font-weight: bold; margin-bottom: 3px; color: #7f8c8d;" for="claseActividad">Clase de Actividad:</label>
                        <select style="width: 100%; padding: 6px; border-radius: 4px; border: 1px solid #ccc; font-size: 12px;" id="claseActividad">
                            <option value="todos">Cargando clases...</option>
                        </select>
                    </div>

                    <div style="margin-bottom: 10px;">
                        <label style="display: block; font-size: 11px; font-weight: bold; margin-bottom: 3px; color: #7f8c8d;" for="tipoActividad">Tipo de Actividad:</label>
                        <select style="width: 100%; padding: 6px; border-radius: 4px; border: 1px solid #ccc; font-size: 12px;" id="tipoActividad">
                            <option value="todos">Todos los tipos</option>
                        </select>
                    </div>

                    <div style="margin-bottom: 10px;">
                        <label style="display: block; font-size: 11px; font-weight: bold; margin-bottom: 3px; color: #7f8c8d;" for="responsable">Responsable / Cuadrilla:</label>
                        <select style="width: 100%; padding: 6px; border-radius: 4px; border: 1px solid #ccc; font-size: 12px;" id="responsable">
                            <option value="todos">Todos los responsables</option>
                            <option value="oscar">responsable</option>
                            <option value="andres">responsable</option>
                            <option value="alfonso">responsable</option>
                        </select>
                    </div>
                </div>
            </div>
        `;

        L.DomEvent.disableClickPropagation(div);
        L.DomEvent.disableScrollPropagation(div);

        // Control de apertura y cierre
        var btnToggle = div.querySelector('#btn-toggle-filtros');
        var cuerpo = div.querySelector('#cuerpo-filtros');
        var flecha = div.querySelector('#flecha-filtros');

        L.DomEvent.on(btnToggle, 'click', function (e) {
            L.DomEvent.stopPropagation(e);
            if (cuerpo.style.display === 'none') {
                cuerpo.style.display = 'block';
                flecha.style.transform = 'rotate(0deg)'; // Apunta hacia abajo al abrir
            } else {
                cuerpo.style.display = 'none';
                flecha.style.transform = 'rotate(-90deg)'; // Apunta a la izquierda al cerrar
            }
        });

        return div;
    }
});
map.addControl(new MenuFiltrosFlotante());


// ==========================================
// 4. Capas del Shapefile (.JSON de Áreas Verdes)
// ==========================================
var capasresponsable = L.layerGroup();
var capasresponsable = L.layerGroup();
var capasresponsable = L.layerGroup();
var capasCamposDepo = L.layerGroup();
var capasBosqueHumedo = L.layerGroup();

function obtenerEstiloPoligono(jefe) {
    if (!jefe) return { color: "#7f8c8d", fillColor: "#95a5a6", fillOpacity: 0.3, weight: 1 };
    var texto = jefe.toString().toLowerCase().trim();
    if (texto.includes("oscar") || texto.includes("óscar")) {
        return { color: "#800080", fillColor: "#a020f0", fillOpacity: 0.4, weight: 2 }; 
    } else if (texto.includes("andres") || texto.includes("andrés")) {
        return { color: "#0d47a1", fillColor: "#2196f3", fillOpacity: 0.4, weight: 2 }; 
    } else if (texto.includes("alfonso")) {
        return { color: "#b25900", fillColor: "#e67300", fillOpacity: 0.4, weight: 2 }; 
    } else if (texto.includes("campo") || texto.includes("deportivo")) {
        return { color: "#556b2f", fillColor: "#9acd32", fillOpacity: 0.4, weight: 2 }; 
    } else if (texto.includes("bosque") || texto.includes("húmedo") || texto.includes("hume")) { // 🛠️ Corregido el "" que rompía los colores por defecto
        return { color: "#00a86b", fillColor: "#00ff7f", fillOpacity: 0.4, weight: 2 }; 
    }
    return { color: "#7f8c8d", fillColor: "#95a5a6", fillOpacity: 0.3, weight: 1 };
}

var totalesPorJefe = {
    oscar:   { area: 0, perimetro: 0, nombreOficial: "responsable" },
    andres:  { area: 0, perimetro: 0, nombreOficial: "responsable" },
    alfonso: { area: 0, perimetro: 0, nombreOficial: "responsable" },
    campos:  { area: 0, perimetro: 0, nombreOficial: "Campos Deportivos" },
    bosque:  { area: 0, perimetro: 0, nombreOficial: "Bosque Húmedo" },
    otros:   { area: 0, perimetro: 0, nombreOficial: "Otros / Sin Especificar" }
}; // 🛠️ ¡CERRADO CORRECTAMENTE! Adiós al SyntaxError en la línea siguiente

// ==========================================
// 5. CONFIGURACIÓN DE LA CAPA GEOJSON COMPLETA
// ==========================================
var capaBaseGeoJSON = L.geoJSON(null, {
    style: function(feature) {
        var valorJefe = "";
        if (feature.properties) {
            valorJefe = feature.properties.jefes || feature.properties.jefe || feature.properties.name || "";
        }
        return obtenerEstiloPoligono(valorJefe); 
    },
    onEachFeature: function(feature, layer) {
        if (feature.properties) {
            var props = feature.properties;

            // --- 1. LIMPIEZA DE LLAVES (ÁREA Y PERÍMETRO) ---
            var areaOriginal = null;
            var perimetroOriginal = null;

            for (var llave in props) {
                if (props.hasOwnProperty(llave)) {
                    var llaveLimpia = llave.toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "");
                    if (llaveLimpia === "area") areaOriginal = props[llave];
                    if (llaveLimpia === "perimetro" || llaveLimpia === "perimeter") perimetroOriginal = props[llave];
                }
            }

            var numArea = areaOriginal !== null ? parseFloat(areaOriginal) : 0;
            var numPerim = perimetroOriginal !== null ? parseFloat(perimetroOriginal) : 0;
            if (isNaN(numArea)) numArea = 0;
            if (isNaN(numPerim)) numPerim = 0;

            // --- 2. CLASIFICACIÓN POR JEFE ---
            var jefeOriginal = props.jefes || props.jefe || "No especificado";
            var usoZona = props.Uso || "Área Verde";
            var nombreSector = props.Nombre || "Sector sin nombre";
            
            var textoJefe = jefeOriginal.toString().toLowerCase().trim();
            var categoriaDestino = "otros";

            if (textoJefe.includes("oscar") || textoJefe.includes("óscar")) {
                categoriaDestino = "oscar";
                if (typeof capasresponsable !== 'undefined') capasresponsable.addLayer(layer);
            } else if (textoJefe.includes("andres") || textoJefe.includes("andrés")) {
                categoriaDestino = "andres";
                if (typeof capasresponsable !== 'undefined') capasresponsable.addLayer(layer);
            } else if (textoJefe.includes("alfonso")) {
                categoriaDestino = "alfonso";
                if (typeof capasresponsable !== 'undefined') capasresponsable.addLayer(layer);
            } else if (textoJefe.includes("campo") || textoJefe.includes("depo")) {
                categoriaDestino = "campos";
                if (typeof capasCamposDepo !== 'undefined') capasCamposDepo.addLayer(layer);
            } else if (textoJefe.includes("bosque") || textoJefe.includes("húme") || textoJefe.includes("hume")) {
                categoriaDestino = "bosque";
                if (typeof capasBosqueHumedo !== 'undefined') capasBosqueHumedo.addLayer(layer);
            }

            // --- 3. SUMA EN EL ACUMULADOR GLOBAL ---
            if (totalesPorJefe && totalesPorJefe[categoriaDestino]) {
                totalesPorJefe[categoriaDestino].area += numArea;
                totalesPorJefe[categoriaDestino].perimetro += numPerim;
            }

            // --- 4. EVENTO AL ABRIR EL POPUP (MUESTRA EL TOTAL GRUPAL) ---
            layer.on('popupopen', function() {
                var datosGlobalesJefe = totalesPorJefe[categoriaDestino];
                
                var textoAreaGlobal = datosGlobalesJefe.area > 0 ? `${datosGlobalesJefe.area.toLocaleString('es-PE', {maximumFractionDigits: 2})} m²` : "No registrado";
                var textoPerimGlobal = datosGlobalesJefe.perimetro > 0 ? `${datosGlobalesJefe.perimetro.toLocaleString('es-PE', {maximumFractionDigits: 2})} m` : "No registrado";

                var popupContenido = `
                    <div style="padding:6px; font-size:12px; width:230px; line-height:1.4; color:#2d3748;">
                        <span style="font-size:9px; text-transform:uppercase; color:#8e44ad; font-weight:bold; display:block; margin-bottom:2px;">👨‍💼 Resumen Consolidado</span>
                        <h3 style="margin:0 0 6px 0; font-size:13px; color:#2c3e50; border-bottom:2px solid #8e44ad; padding-bottom:4px; font-weight:bold;">📍 ${nombreSector}</h3>
                        <p style="margin:2px 0;"><b>👨‍💼 Responsable:</b> ${datosGlobalesJefe.nombreOficial}</p>
                        <p style="margin:2px 0;"><b>🏛️ Uso:</b> ${usoZona}</p>
                        <hr style="border:0; border-top:1px dashed #b2bec3; margin:6px 0;">
                        <p style="margin:2px 0;"><b>📐 Área Total de su Color:</b> <br><span style="color:#27ae60; font-weight:bold; font-size:13px;">${textoAreaGlobal}</span></p>
                        <p style="margin:2px 0;"><b>📏 Perímetro Total:</b> <br><span style="color:#2980b9; font-weight:bold; font-size:13px;">${textoPerimGlobal}</span></p>
                    </div>
                `;
                layer.setPopupContent(popupContenido);
            });

            layer.bindPopup("Calculando totales...");
        }
    } // Cierre de onEachFeature
}); // 🛠️ ¡Cierre correcto de L.geoJSON! Esto evitará que la línea 194 (o las siguientes) falle.

function mostrarResumenMétricas() {
    console.log("%c📊 RESUMEN DE MÉTRICAS TOTALES POR CATEGORÍA:", "font-weight: bold; font-size: 14px; color: #2e4053;");
    
    for (var cat in totalesPorJefe) {
        var item = totalesPorJefe[cat];
        console.log(
            `👤 %c${item.nombreOficial.padEnd(20, ' ')}` + 
            `-> 📐 Área Total: %c${item.area.toLocaleString('es-PE', {maximumFractionDigits:2}).padStart(12, ' ')} m² ` + 
            `| 📏 Perímetro Total: %c${item.perimetro.toLocaleString('es-PE', {maximumFractionDigits:2}).padStart(12, ' ')} m`,
            "font-weight: bold; color: #1a5276;", 
            "color: #27ae60; font-weight: bold;", 
            "color: #e67e22;"
        );
    }
}


// ==========================================
// 5. Capas de Fauna - Plan de Manejo (.GEOJSON)
// ==========================================
var capaAves = L.layerGroup();
var capaCanes = L.layerGroup();
var capaInsectos = L.layerGroup();
var capaGallinazos = L.layerGroup();
var capaArdillas = L.layerGroup();
var capaRoedores = L.layerGroup();
var capaColoniasGatos = L.layerGroup(); // Reintegrada como subcapa del Plan de Fauna

function obtenerConfiguracionAnimal(nombreAnimal) {
    if (!nombreAnimal) return { id: "?", color: "#7f8c8d" };
    var name = nombreAnimal.toString().toLowerCase().trim();
    if (name.includes("ave")) return { id: "1", color: "#cddc39" };         
    if (name.includes("cane") || name.includes("perro")) return { id: "2", color: "#673ab7" }; 
    if (name.includes("roedor") || name.includes("rata")) return { id: "3", color: "#4caf50" }; 
    if (name.includes("gallinazo")) return { id: "4", color: "#ff5722" };    
    if (name.includes("ardilla")) return { id: "5", color: "#00bfa5" };      
    if (name.includes("insecto")) return { id: "6", color: "#4a5568" }; // Corregido ID duplicado (tenía 5 igual que ardilla)
    if (name.includes("colonia") || name.includes("gato")) return { id: "🐈", color: "#d32f2f" }; 
    return { id: "⚠️", color: "#e53935" }; 
}

var capaFaunaGeoJSON = L.geoJSON(null, {
    pointToLayer: function(feature, latlng) {
        var animalAttr = feature.properties.Animal || "";
        var config = obtenerConfiguracionAnimal(animalAttr);
        var iconoDiv = L.divIcon({
            className: 'fauna-marker-custom',
            html: `<div style="color:white; font-weight:bold; font-size:11px; text-align:center; border-radius:50%; border:2px solid white; box-shadow:0 2px 5px rgba(0,0,0,0.3); display:flex; align-items:center; justify-content:center; background-color: ${config.color}; width: 24px; height: 24px;">${config.id}</div>`,
            iconSize: [24, 24],
            iconAnchor: [12, 12]
        });
        return L.marker(latlng, { icon: iconoDiv });
    },
    onEachFeature: function(feature, layer) {
        if (feature.properties) {
            var animalNombre = feature.properties.Animal || "Registro UGA";
            var animalLower = animalNombre.toLowerCase();
            
            var recomendacionText = "Seguir los lineamientos del Plan de Manejo Ambiental.";
            var subTituloPopup = "UGA - Plan de Fauna"; // Estandarizado para unificar la pertenencia

            // Alertas internas específicas sin romper el esquema de la categoría
            if (animalLower.includes("gato") || animalLower.includes("colonia")) {
                recomendacionText = "❌ <b>Riesgo Sanitario (Especie Feral):</b> Prohibido alimentar o abandonar. Controlar puntos de concentración.";
            } else if (animalLower.includes("roedor") || animalLower.includes("rata")) {
                recomendacionText = "⚠️ <b>Riesgo Biológico (Vectores):</b> Mantener contenedores sellados. Reportar madrigueras activas.";
            } else if (animalLower.includes("ardilla") || animalLower.includes("ave")) {
                recomendacionText = "🌿 <b>Conservación de Fauna:</b> Especie protegida. Respetar zonas de anidamiento.";
            }

            layer.bindPopup(`
                <div style="padding: 10px; font-size:13px; max-width:250px; word-wrap:break-word; line-height:1.4;">
                    <span style="font-size:10px; text-transform:uppercase; color:#718096; font-weight:bold;">${subTituloPopup}</span>
                    <h3 style="margin: 4px 0 6px 0; font-size:14px; color:#1a365d; border-bottom: 1px solid #edf2f7; padding-bottom:4px;">🐾 Tipo: ${animalNombre}</h3>
                    <b>Recomendación:</b><br><i style="color:#2b6cb0;">${recomendacionText}</i>
                </div>
            `);

            if (animalLower.includes("ardilla")) capaArdillas.addLayer(layer);
            else if (animalLower.includes("ave")) capaAves.addLayer(layer);
            else if (animalLower.includes("cane")) capaCanes.addLayer(layer);
            else if (animalLower.includes("insecto")) capaInsectos.addLayer(layer);
            else if (animalLower.includes("gallinazo")) capaGallinazos.addLayer(layer);
            else if (animalLower.includes("roedor")) capaRoedores.addLayer(layer);
            else capaColoniasGatos.addLayer(layer); // Los inserta directo en el flujo de capas del mapa
        }
    }
});


// ==========================================
// GeoJSON de Zonas (supervisoress.geojson)
// ==========================================
var geojsonSupervisores = {
  "type": "FeatureCollection",
  "name": "supervisoress",
  "crs": { "type": "name", "properties": { "name": "urn:ogc:def:crs:OGC:1.3:CRS84" } },
  "features": [
    { 
      "type": "Feature", 
      "properties": { "id": 1, "ZONA": "Zona1", "Area": 91242.754 }, 
      "geometry": { 
        "type": "MultiPolygon", 
        "coordinates": [ [ [ [ -77.078049096413466, -12.069388227116226 ], [ -77.078057808698276, -12.069363702224305 ], [ -77.078088245325873, -12.069293470428677 ], [ -77.078093689850633, -12.069284648774195 ], [ -77.07810994460371, -12.069253104658062 ], [ -77.078139464987274, -12.069221291405798 ], [ -77.078162149821054, -12.069206132633596 ], [ -77.078180847405847, -12.069198714173577 ], [ -77.078202417829246, -12.069193859852529 ], [ -77.078315194681906, -12.069188146448354 ], [ -77.078354902665708, -12.069192168715363 ], [ -77.078391044457547, -12.069202521385689 ], [ -77.078428285855026, -12.06922044966403 ], [ -77.078446748549325, -12.069233303959519 ], [ -77.078463770851826, -12.069248050937393 ], [ -77.078471468964239, -12.06925602747835 ], [ -77.078488763841577, -12.069275600529748 ], [ -77.07849775362051, -12.069290525208228 ], [ -77.078503281620115, -12.069300367876833 ], [ -77.078510063401993, -12.069311851841295 ], [ -77.078512671459336, -12.069320865148304 ], [ -77.078522034996183, -12.069348108438493 ], [ -77.078523533139702, -12.069363943954693 ], [ -77.078524105650956, -12.069366530239639 ], [ -77.079745173877726, -12.069257515764285 ], [ -77.079694207309046, -12.068711307846472 ], [ -77.079695489282599, -12.06866894283781 ], [ -77.079705922200489, -12.068352980460631 ], [ -77.079711127751125, -12.068089308700211 ], [ -77.079715768897742, -12.067961020193863 ], [ -77.079718764708005, -12.067827993839758 ], [ -77.079723118851746, -12.067662498033092 ], [ -77.079727862106751, -12.067443164377265 ], [ -77.079736281684461, -12.067152946967031 ], [ -77.079709635034831, -12.067089132359538 ], [ -77.079701530481401, -12.066998213723148 ], [ -77.079655305842948, -12.066950315325352 ], [ -77.079697775070954, -12.066910727426006 ], [ -77.079668232436831, -12.066880052389923 ], [ -77.079669075852664, -12.066879439608771 ], [ -77.07966987378262, -12.066878770415785 ], [ -77.079670622339236, -12.066878048071199 ], [ -77.079671317875622, -12.066877276094205 ], [ -77.079671957003214, -12.066876458245799 ], [ -77.079672536608243, -12.066875598510451 ], [ -77.079673053866912, -12.066874701076721 ], [ -77.079673506259198, -12.066873770316798 ], [ -77.079673891581095, -12.066872810765258 ], [ -77.079674207955378, -12.066871827096925 ], [ -77.079674453840667, -12.066870824104161 ], [ -77.079674628039044, -12.06686980667345 ], [ -77.079674729701836, -12.066868779761567 ], [ -77.079674758333766, -12.066867748371568 ], [ -77.079674713795313, -12.066866717528258 ], [ -77.079674596303491, -12.066865692253803 ], [ -77.079674406430712, -12.066864677543256 ], [ -77.07967414510199, -12.066863678340161 ], [ -77.079673813590531, -12.066862699512535 ], [ -77.079673413511387, -12.06686174582914 ], [ -77.07967294681373, -12.066860821936237 ], [ -77.079672415771242, -12.066859932334918 ], [ -77.07967182297115, -12.06685908135923 ], [ -77.07967117130147, -12.066858273155049 ], [ -77.079670463937106, -12.066857511659867 ], [ -77.079669704324232, -12.066856800583599 ], [ -77.079669351773276, -12.066856502359462 ], [ -77.079699996850366, -12.066827834419792 ], [ -77.079659439654279, -12.066785838693647 ], [ -77.079704036262115, -12.066745050376941 ], [ -77.079674649389617, -12.066715197582644 ], [ -77.079675130219698, -12.066714846611971 ], [ -77.0796759281492, -12.066714177419041 ], [ -77.079676676705418, -12.066713455074508 ], [ -77.079677372241449, -12.066712683097562 ], [ -77.079678011368699, -12.0667118652492 ], [ -77.07967859097343, -12.066711005513891 ], [ -77.079679108231858, -12.066710108080194 ], [ -77.079679560623944, -12.066709177320304 ], [ -77.079679945945699, -12.066708217768781 ], [ -77.079680508205072, -12.06670623110772 ], [ -77.079680682403421, -12.066705213677013 ], [ -77.079680812698228, -12.066703155375123 ], [ -77.079680768159903, -12.066702124531806 ], [ -77.079680650668237, -12.066701099257328 ], [ -77.079680460795657, -12.066700084546763 ], [ -77.079680199467177, -12.06669908534364 ], [ -77.079679867955988, -12.066698106515981 ], [ -77.079679467877185, -12.066697152832552 ], [ -77.079679001179869, -12.066696228939596 ], [ -77.079678470137807, -12.066695339338237 ], [ -77.079677877338128, -12.066694488362494 ], [ -77.079677225668917, -12.066693680158256 ], [ -77.079676518305035, -12.066692918663012 ], [ -77.079675529515853, -12.066692018600859 ], [ -77.079706202434622, -12.066661872659111 ], [ -77.079665693424971, -12.066619905794941 ], [ -77.079708333671789, -12.066580036517125 ], [ -77.079680138238857, -12.066550826366353 ], [ -77.079681046173903, -12.06655008953639 ], [ -77.079681900303413, -12.066549292820367 ], [ -77.07968269655828, -12.066548440013893 ], [ -77.079683431145099, -12.066547535179765 ], [ -77.079684100564251, -12.066546582628671 ], [ -77.079684701626618, -12.066545586898622 ], [ -77.079685231468687, -12.066544552733296 ], [ -77.079685687566268, -12.066543485059542 ], [ -77.079686067746493, -12.066542388963798 ], [ -77.079686370198161, -12.066541269667921 ], [ -77.079686593480389, -12.066540132504285 ], [ -77.079686736529439, -12.066538982890398 ], [ -77.079686798663815, -12.066537826303113 ], [ -77.079686779587519, -12.066536668252418 ], [ -77.079686679391429, -12.066535514255369 ], [ -77.079686498552874, -12.066534369809631 ], [ -77.079686237933402, -12.066533240367431 ], [ -77.079685898774599, -12.066532131309463 ], [ -77.079685482692227, -12.066531047919351 ], [ -77.07968442804281, -12.066528978641129 ], [ -77.07968379450017, -12.066528002611147 ], [ -77.079683094058851, -12.066527071918363 ], [ -77.079682330055775, -12.066526190996619 ], [ -77.079681506130711, -12.066525364042667 ], [ -77.079680626208869, -12.06652459499619 ], [ -77.079710305043676, -12.066496844827324 ], [ -77.07966859875286, -12.066455420165193 ], [ -77.079712436286798, -12.066415008684821 ], [ -77.079684240873831, -12.066385798531345 ], [ -77.079685975315371, -12.066384352390925 ], [ -77.079686764423684, -12.066383546426783 ], [ -77.079688168521997, -12.066381786986199 ], [ -77.079689319687361, -12.066379857817866 ], [ -77.079689793705739, -12.06637884041626 ], [ -77.079690527865324, -12.066376723048103 ], [ -77.079690784674042, -12.06637563269274 ], [ -77.079691071926064, -12.066373413363682 ], [ -77.079691053588448, -12.066371176172556 ], [ -77.079690729993743, -12.066368961693557 ], [ -77.079690455344959, -12.066367875557892 ], [ -77.079689686572294, -12.066365770122868 ], [ -77.07968919593803, -12.066364760380472 ], [ -77.079688013298963, -12.066362849742255 ], [ -77.079686580542131, -12.066361112825946 ], [ -77.079685778324972, -12.066360319496725 ], [ -77.079684347202942, -12.066359478322541 ], [ -77.079821563024723, -12.066231845938097 ], [ -77.079725001747946, -12.066132038412439 ], [ -77.079738172981109, -12.066080269935229 ], [ -77.079764939596942, -12.065975065853255 ], [ -77.079738333329914, -12.065968626863206 ], [ -77.079772848694176, -12.065832474443543 ], [ -77.079732631809406, -12.065823492017724 ], [ -77.079700725508346, -12.065815596098702 ], [ -77.079707789914252, -12.065790452167535 ], [ -77.079734023489294, -12.065686236074898 ], [ -77.079749559145625, -12.065628788172509 ], [ -77.079487459512251, -12.065565685941085 ], [ -77.079688141255502, -12.064721114472254 ], [ -77.079702118918092, -12.064666447034258 ], [ -77.078983548783242, -12.064494656380814 ], [ -77.078594218940552, -12.064402047221856 ], [ -77.078562617078518, -12.064396792335245 ], [ -77.078536125596131, -12.064398178478447 ], [ -77.078507276967585, -12.064403309840861 ], [ -77.078489936060151, -12.064408999486465 ], [ -77.078454389394807, -12.064429276960103 ], [ -77.078433340875861, -12.064447875733148 ], [ -77.078413340552501, -12.064472741435386 ], [ -77.078401219797257, -12.064495764035586 ], [ -77.078392998657421, -12.064519706953362 ], [ -77.078386512504906, -12.064551686594802 ], [ -77.078362359007826, -12.064664127269181 ], [ -77.078315186992199, -12.064889016838519 ], [ -77.078304418619808, -12.064944713906529 ], [ -77.078272928002406, -12.065115684348211 ], [ -77.078252716922762, -12.065228915400713 ], [ -77.078214183284942, -12.065454320154078 ], [ -77.078188791364795, -12.065625673359623 ], [ -77.078165299847356, -12.065796532444709 ], [ -77.078141426218323, -12.065967068271204 ], [ -77.078134502841976, -12.066020981025973 ], [ -77.078120443620094, -12.066138247227244 ], [ -77.078094473840281, -12.066366636333795 ], [ -77.078071563531083, -12.066595294915144 ], [ -77.078067146975698, -12.066650195383861 ], [ -77.078062736135081, -12.066707654955653 ], [ -77.078056380615905, -12.066820379737619 ], [ -77.078049000179192, -12.066933389530044 ], [ -77.078037193620659, -12.06713562471394 ], [ -77.07803313966086, -12.06716039282888 ], [ -77.078027593152271, -12.0672792507378 ], [ -77.078019702003914, -12.067465512248647 ], [ -77.07801378037729, -12.067565958528984 ], [ -77.078011158592773, -12.067683906038878 ], [ -77.078005711175607, -12.067909107290088 ], [ -77.07800449336473, -12.068016406081 ], [ -77.078003589150626, -12.068154520292039 ], [ -77.078001740458021, -12.068312714345023 ], [ -77.078008850789985, -12.068541636224348 ], [ -77.07802016030567, -12.068876107810935 ], [ -77.078027185967017, -12.069030523180116 ], [ -77.078040144287129, -12.069241744541459 ], [ -77.078049096413466, -12.069388227116226 ] ] ] ] } 
    },
    { 
      "type": "Feature", 
      "properties": { "id": 2, "ZONA": "Zona2", "Area": 80863.431 }, 
      "geometry": { 
        "type": "MultiPolygon", 
        "coordinates": [ [ [ [ -77.078049096413466, -12.069388227116226 ], [ -77.078048153579203, -12.069388854846398 ], [ -77.078057658022757, -12.069490136974126 ], [ -77.078087819679425, -12.069733405614782 ], [ -77.078118042232617, -12.069966325946247 ], [ -77.078161615935272, -12.070249983543198 ], [ -77.07819232593792, -12.070419515301023 ], [ -77.078214100043667, -12.070532370876922 ], [ -77.07824874067407, -12.070700857709024 ], [ -77.078286071004499, -12.070869266607113 ], [ -77.078369655727101, -12.071203820026376 ], [ -77.078463770851826, -12.071538123482192 ], [ -77.078472016909544, -12.071629410971509 ], [ -77.078494205790534, -12.071843717340485 ], [ -77.078494205790534, -12.071843717340485 ], [ -77.078504506807221, -12.071910937994117 ], [ -77.078508690363918, -12.071945422188163 ], [ -77.078517326152735, -12.072013945914469 ], [ -77.078519828990167, -12.072044465624829 ], [ -77.078520611132916, -12.072079637179126 ], [ -77.078522789045309, -12.072114863555433 ], [ -77.078526236895243, -12.072186520650567 ], [ -77.078527569357249, -12.072212279113328 ], [ -77.078528995995413, -12.07225484029434 ], [ -77.078531072286324, -12.072329022832994 ], [ -77.078532094242775, -12.072364423225542 ], [ -77.078533262091824, -12.072398385350155 ], [ -77.078607160534474, -12.07245458286083 ], [ -77.07861702521798, -12.072623908699235 ], [ -77.078553214872628, -12.072732242170384 ], [ -77.07856111558182, -12.072792600831598 ], [ -77.078576993279682, -12.072881029113161 ], [ -77.078606670889116, -12.07302958741845 ], [ -77.07863074555479, -12.07311975405433 ], [ -77.078642673323685, -12.073160951647111 ], [ -77.078654732399528, -12.073196576377896 ], [ -77.079667765188717, -12.073022769612008 ], [ -77.079929270800577, -12.072949205460166 ], [ -77.079938271254292, -12.073000068783879 ], [ -77.080391698589082, -12.0729079207306 ], [ -77.080482804922752, -12.072886439897854 ], [ -77.080497122532563, -12.072886578920851 ], [ -77.08076915207937, -12.072818172240151 ], [ -77.080772883677611, -12.072811521732609 ], [ -77.0807931387087, -12.072805915049088 ], [ -77.080772342129109, -12.072724677710655 ], [ -77.080636966673652, -12.072602348634877 ], [ -77.080626825208256, -12.072593174436561 ], [ -77.080609922767408, -12.072577884105016 ], [ -77.080599061223211, -12.072569493481733 ], [ -77.080595035305947, -12.072567150205016 ], [ -77.080586245511839, -12.072563129584051 ], [ -77.080576989926996, -12.072560296957938 ], [ -77.080572238755281, -12.072559342769704 ], [ -77.080527107892294, -12.072465535816896 ], [ -77.080522197073535, -12.072441924970189 ], [ -77.080702229714106, -12.072403606631797 ], [ -77.080657046966991, -12.072197435560238 ], [ -77.080631477636402, -12.072141030374004 ], [ -77.080595906713882, -12.071979297440988 ], [ -77.080587577978875, -12.071943991674315 ], [ -77.080527314431279, -12.07169305458158 ], [ -77.080341694147819, -12.071732937797627 ], [ -77.080354671353973, -12.071791942612554 ], [ -77.080257151127199, -12.071812896202053 ], [ -77.080251911767832, -12.071789084758068 ], [ -77.080213689459413, -12.07179659180362 ], [ -77.080147569477688, -12.071810798578513 ], [ -77.080116213243898, -12.071818840096897 ], [ -77.080073100802053, -12.071661244249366 ], [ -77.080007684250575, -12.071473529694281 ], [ -77.079952741056474, -12.071218630417915 ], [ -77.07987976652575, -12.070706170537848 ], [ -77.079840396595031, -12.070286390545473 ], [ -77.079745173877726, -12.069257515764285 ], [ -77.078578678760522, -12.069361751735377 ], [ -77.078524105650956, -12.069366530239639 ], [ -77.078522807435206, -12.069356194015745 ], [ -77.07851965664922, -12.069341819976229 ], [ -77.078516505660957, -12.069331459419921 ], [ -77.078512671459336, -12.069320865148304 ], [ -77.078508259116632, -12.069310491122375 ], [ -77.078503281620115, -12.069300367876833 ], [ -77.07849775362051, -12.069290525208228 ], [ -77.078491982610828, -12.069281698994789 ], [ -77.078485724377273, -12.069273200472864 ], [ -77.078478838818327, -12.069264419269794 ], [ -77.078471468964239, -12.06925602747835 ], [ -77.078463637507312, -12.069248050937393 ], [ -77.078455368561094, -12.069240514207218 ], [ -77.078446748549325, -12.069233303959519 ], [ -77.078437710438195, -12.069226607967973 ], [ -77.078428285855026, -12.06922044966403 ], [ -77.078418507779574, -12.069214850597634 ], [ -77.078404969624088, -12.069208268472496 ], [ -77.078391044457547, -12.069202521385689 ], [ -77.078376785391299, -12.069197631256941 ], [ -77.078359094656378, -12.069193126222903 ], [ -77.078343765518028, -12.069190333986315 ], [ -77.078332542690248, -12.069189104964828 ], [ -77.078320004826026, -12.069188568387686 ], [ -77.078202417829246, -12.069193859852529 ], [ -77.078191560389797, -12.069195976128009 ], [ -77.078170315603501, -12.069202064602743 ], [ -77.078162149821054, -12.069206132633596 ], [ -77.078133454605947, -12.069226149433266 ], [ -77.07810994460371, -12.069253104658062 ], [ -77.078102463080995, -12.069266443120055 ], [ -77.078088245325873, -12.069293470428677 ], [ -77.078057808698276, -12.069363702224305 ], [ -77.078049096413466, -12.069388227116226 ] ] ] ] } 
    },
    { 
      "type": "Feature", 
      "properties": { "id": 3, "ZONA": "Zona3", "Area": 112022.743 }, 
      "geometry": { 
        "type": "MultiPolygon", 
        "coordinates": [ [ [ [ -77.081931191776064, -12.069717817841434 ], [ -77.081926128685609, -12.069717550379048 ], [ -77.081923595217347, -12.069717676776657 ], [ -77.081918585562818, -12.069718446779673 ], [ -77.081916133574708, -12.069719086665746 ], [ -77.081905939856782, -12.069721859006806 ], [ -77.081748742742121, -12.069772129209621 ], [ -77.081746366669933, -12.06977306708132 ], [ -77.081739712945691, -12.069776817780975 ], [ -77.081501869391232, -12.069737502839144 ], [ -77.08146234577498, -12.069740846446161 ], [ -77.081467545320621, -12.069800546806553 ], [ -77.08137029157848, -12.069808774229083 ], [ -77.081370648408793, -12.069812871312983 ], [ -77.081220879450669, -12.069825541324816 ], [ -77.081220522622573, -12.069821444240342 ], [ -77.081161508192309, -12.069826396861131 ], [ -77.081161903732024, -12.069830937124541 ], [ -77.081131711331579, -12.069833484520045 ], [ -77.08111780025061, -12.069678693497858 ], [ -77.081056440810386, -12.069666985464849 ], [ -77.081050099556691, -12.069602845997737 ], [ -77.080944419128841, -12.069611755924093 ], [ -77.080932888208181, -12.069615508883448 ], [ -77.080865752564435, -12.06962139140072 ], [ -77.080870432037528, -12.069672790195712 ], [ -77.080873212181245, -12.06969987510929 ], [ -77.080696048529987, -12.069715146804416 ], [ -77.080695362464198, -12.069705951656578 ], [ -77.080679083064766, -12.069707424088618 ], [ -77.080679769130029, -12.069716619236583 ], [ -77.080589295591338, -12.069724802330844 ], [ -77.080586862577462, -12.069695074225667 ], [ -77.080426475365911, -12.069708997935413 ], [ -77.080414562253409, -12.069707424677706 ], [ -77.080424395199699, -12.069819156806354 ], [ -77.080378310840928, -12.069818080168233 ], [ -77.080291850816181, -12.069919155921157 ], [ -77.080294355415518, -12.069931452711076 ], [ -77.080266741965659, -12.069933973496408 ], [ -77.080268236303638, -12.069950356127265 ], [ -77.080125702876174, -12.069963367721394 ], [ -77.080099802274404, -12.069942371179851 ], [ -77.080088755395863, -12.069821261805938 ], [ -77.0798703272594, -12.06984198206251 ], [ -77.079825276224682, -12.0698952379455 ], [ -77.079804173125083, -12.069896882818329 ], [ -77.079840396595031, -12.070286390545473 ], [ -77.07987976652575, -12.070706170537848 ], [ -77.079952741056474, -12.071218630417915 ], [ -77.080007684250575, -12.071473529694281 ], [ -77.080073100802053, -12.071661244249366 ], [ -77.080116213243898, -12.071818840096897 ], [ -77.080251911767832, -12.071789084758068 ], [ -77.080257151127199, -12.071812896202053 ], [ -77.080354671353973, -12.071791942612554 ], [ -77.080341694147819, -12.071732937797627 ], [ -77.080527314431279, -12.07169305458158 ], [ -77.080587577978875, -12.071943991674315 ], [ -77.080595906713882, -12.071979297440988 ], [ -77.080631477636402, -12.072141030374004 ], [ -77.080657046966991, -12.072197435560238 ], [ -77.080702229714106, -12.072403606631797 ], [ -77.080522197073535, -12.072441924970189 ], [ -77.080572238755281, -12.072559342769704 ], [ -77.080576989926996, -12.072560296957938 ], [ -77.080586245511839, -12.072563129584051 ], [ -77.080595035305947, -12.072567150205016 ], [ -77.080599061223211, -12.072569493481733 ], [ -77.080609922767408, -12.072577884105016 ], [ -77.080772342129109, -12.072724677710655 ], [ -77.0807931387087, -12.072805915049088 ], [ -77.080795659607958, -12.072815506271034 ], [ -77.081910013252255, -12.07251326889889 ], [ -77.081973812710999, -12.072735602448516 ], [ -77.081979013026668, -12.072741805934342 ], [ -77.081050509040338, -12.072980895266792 ], [ -77.080735597187498, -12.073062928660068 ], [ -77.080701949995813, -12.073064832390802 ], [ -77.080645401649477, -12.073074180507982 ], [ -77.078969075847667, -12.073353603454603 ], [ -77.078743810559729, -12.073420916443581 ], [ -77.078786605905762, -12.073512743014168 ], [ -77.078816010569284, -12.073566697663031 ], [ -77.07884434145511, -12.073620836545386 ], [ -77.078876256794345, -12.073673289763233 ], [ -77.078905468441732, -12.07372274949144 ], [ -77.078922822223277, -12.073748739043445 ], [ -77.078939053391665, -12.0737736323376 ], [ -77.0789748269558, -12.073827768957173 ], [ -77.079009584346664, -12.073874384669999 ], [ -77.079046554754854, -12.073922936209714 ], [ -77.079086464504485, -12.07397388292882 ], [ -77.079125658920105, -12.07402067563217 ], [ -77.079162673020363, -12.074065400780366 ], [ -77.079167699669583, -12.074071054896544 ], [ -77.07917322649952, -12.074076049888721 ], [ -77.079178258014608, -12.074079303160881 ], [ -77.079185049461628, -12.074081875120736 ], [ -77.079191581155811, -12.074083114791151 ], [ -77.079198564315405, -12.074083113028722 ], [ -77.07920477624485, -12.07408205796329 ], [ -77.079211285060197, -12.074079587470907 ], [ -77.079362470063614, -12.073994013643029 ], [ -77.079339401555259, -12.073955029724788 ], [ -77.079466949896286, -12.073881482337313 ], [ -77.079489867471978, -12.073921556614076 ], [ -77.079541165788768, -12.073892946749881 ], [ -77.079595908740814, -12.073862201080027 ], [ -77.079617923986916, -12.073850367515837 ], [ -77.079652264869353, -12.073833130778468 ], [ -77.079708157492746, -12.073804589814209 ], [ -77.079764164939178, -12.073782620244858 ], [ -77.079822728547683, -12.073761773138875 ], [ -77.07987147720732, -12.073748247524962 ], [ -77.079944633306383, -12.073728315340777 ], [ -77.079987681383216, -12.073720259790417 ], [ -77.080192628766795, -12.073699193353418 ], [ -77.080217140899109, -12.073700443492273 ], [ -77.080442399988513, -12.07368377566628 ], [ -77.080691799834, -12.073666711047832 ], [ -77.080755727098193, -12.073662373938419 ], [ -77.081103038757988, -12.073639068222624 ], [ -77.081212473043621, -12.073632866450044 ], [ -77.081313680158885, -12.073623356689444 ], [ -77.081396214225549, -12.07361656545676 ], [ -77.08145716300416, -12.073606241376643 ], [ -77.081495347246886, -12.073599428107512 ], [ -77.081534689576856, -12.073590516152549 ], [ -77.081563499626981, -12.073583096704269 ], [ -77.081587634612035, -12.073574289515758 ], [ -77.081614527951643, -12.073566538117401 ], [ -77.081672481768223, -12.073547161779688 ], [ -77.081829456333551, -12.073494226747407 ], [ -77.081877938902579, -12.073476285050296 ], [ -77.081972362599743, -12.073446035742741 ], [ -77.082030336240038, -12.073427632192235 ], [ -77.082203504607733, -12.073369317882859 ], [ -77.082258198343254, -12.073349766128146 ], [ -77.082277127684492, -12.073343077324006 ], [ -77.082298821151056, -12.073336897209908 ], [ -77.082367284689653, -12.07331262787325 ], [ -77.082607286639146, -12.073232716508507 ], [ -77.083063991823408, -12.073079110169621 ], [ -77.083112501018007, -12.073061190700699 ], [ -77.083117290659985, -12.073057971676938 ], [ -77.08312076325744, -12.073054164236728 ], [ -77.083122709980799, -12.073050825203165 ], [ -77.083125162565025, -12.073045906976011 ], [ -77.08312715881182, -12.073038512952241 ], [ -77.083127549458496, -12.073032803148955 ], [ -77.083127117742194, -12.073027445372743 ], [ -77.083124786677757, -12.073019715930444 ], [ -77.08294337136094, -12.072481770140744 ], [ -77.082865718659633, -12.072502672845609 ], [ -77.082797288312335, -12.072520467425859 ], [ -77.082069940586123, -12.072720594595559 ], [ -77.08202363868412, -12.072695636537263 ], [ -77.081983485049278, -12.072705167835098 ], [ -77.081924300654151, -12.07249685069335 ], [ -77.082730636738233, -12.072275765599896 ], [ -77.08273827407757, -12.072274051414679 ], [ -77.082746003831531, -12.072272801276783 ], [ -77.082753796919548, -12.072272019889448 ], [ -77.082757265386718, -12.072271672806204 ], [ -77.08276074976223, -12.072271557355203 ], [ -77.082764234094441, -12.072271674065005 ], [ -77.082767702431894, -12.072272022401279 ], [ -77.082771138896348, -12.072272600769331 ], [ -77.082774527755504, -12.072273406521386 ], [ -77.082777853494989, -12.07227443596862 ], [ -77.082781306704106, -12.072276207437445 ], [ -77.08278293958945, -12.072277253039582 ], [ -77.082784500930856, -12.072278399617568 ], [ -77.082785984291661, -12.072279642444629 ], [ -77.082788692957593, -12.072282395976087 ], [ -77.082789907096227, -12.072283895329052 ], [ -77.082791020967363, -12.072285468275023 ], [ -77.082792310570127, -12.072287445141148 ], [ -77.082793477192354, -12.07228949467518 ], [ -77.082794516621007, -12.072291609475647 ], [ -77.082795425102418, -12.072293781905447 ], [ -77.082813204937196, -12.072347247613866 ], [ -77.082812683488754, -12.072347423494731 ], [ -77.082846938100545, -12.072449059562029 ], [ -77.082847593998352, -12.072448845551385 ], [ -77.082865718659633, -12.072502672845609 ], [ -77.08294337145766, -12.072481769946728 ], [ -77.082907297484539, -12.072373442194262 ], [ -77.082652614596952, -12.071617207165872 ], [ -77.0825707866762, -12.071374750800498 ], [ -77.082550307581712, -12.071314081186365 ], [ -77.082529773432739, -12.071253218758175 ], [ -77.082499102489152, -12.071162447264964 ], [ -77.082474294381868, -12.071094964211094 ], [ -77.082437737904428, -12.070980294164562 ], [ -77.08240149259035, -12.070879547219995 ], [ -77.082385172331271, -12.070824412997691 ], [ -77.082364653136324, -12.07076358152942 ], [ -77.082312041739343, -12.070612623015963 ], [ -77.082293186339669, -12.070551514710706 ], [ -77.082252188780302, -12.070430103213269 ], [ -77.082221491868509, -12.070339175571396 ], [ -77.082201012757508, -12.07027847716353 ], [ -77.08216876967964, -12.070187920704162 ], [ -77.08213822881379, -12.070096863604563 ], [ -77.082119354222257, -12.070035758459229 ], [ -77.082109104318164, -12.070005466870683 ], [ -77.08209742416777, -12.069975664143419 ], [ -77.082088866186282, -12.069944720799839 ], [ -77.08207649589059, -12.06990808591793 ], [ -77.082005252747507, -12.069873433499193 ], [ -77.081961173931973, -12.069741643619929 ], [ -77.081957939742637, -12.069734877081711 ], [ -77.081955020616775, -12.069730796826624 ], [ -77.081951555130374, -12.069727154084958 ], [ -77.081949638023048, -12.069725519117698 ], [ -77.081947610160029, -12.069724019154007 ], [ -77.081945481336547, -12.069722661439153 ], [ -77.08194326183542, -12.069721452531313 ], [ -77.081940962377573, -12.069720398269906 ], [ -77.081938594070053, -12.069719503747326 ], [ -77.081936168352527, -12.069718773284404 ], [ -77.081931191776064, -12.069717817841434 ] ] ] ] } 
    },
    { 
      "type": "Feature", 
      "properties": { "id": 4, "ZONA": "Zona4", "Area": 118736.055 }, 
      "geometry": { 
        "type": "MultiPolygon", 
        "coordinates": [ [ [ [ -77.079702118918092, -12.064666447034258 ], [ -77.079487459512251, -12.065565685941085 ], [ -77.079749559145625, -12.065628788172509 ], [ -77.079700725508346, -12.065815596098702 ], [ -77.079772848694176, -12.065832474443543 ], [ -77.079738333329914, -12.065968626863206 ], [ -77.079764939596942, -12.065975065853255 ], [ -77.079725001747946, -12.066132038412439 ], [ -77.079821563024723, -12.066231845938097 ], [ -77.079684347202942, -12.066359478322541 ], [ -77.079685778324972, -12.066360319496725 ], [ -77.079686580542131, -12.066361112825946 ], [ -77.079688013298963, -12.066362849742255 ], [ -77.079688637335039, -12.066363785445136 ], [ -77.07968919593803, -12.066364760380472 ], [ -77.079689686572294, -12.066365770122868 ], [ -77.079690455344959, -12.066367875557892 ], [ -77.079690729993743, -12.066368961693557 ], [ -77.079691053588448, -12.066371176172556 ], [ -77.079691071926064, -12.066373413363682 ], [ -77.079690784674042, -12.06637563269274 ], [ -77.079690527865324, -12.066376723048103 ], [ -77.079689793705739, -12.06637884041626 ], [ -77.079689319687361, -12.066379857817866 ], [ -77.079688168521997, -12.066381786986199 ], [ -77.079686764423684, -12.066383546426783 ], [ -77.079685975315371, -12.066384352390925 ], [ -77.079684240873831, -12.066385798531345 ], [ -77.079712436286798, -12.066415008684821 ], [ -77.07966859875286, -12.066455420165193 ], [ -77.079710305043676, -12.066496844827324 ], [ -77.079680626208869, -12.06652459499619 ], [ -77.079681506130711, -12.066525364042667 ], [ -77.079682330055775, -12.066526190996619 ], [ -77.079683094058851, -12.066527071918363 ], [ -77.07968379450017, -12.066528002611147 ], [ -77.07968442804281, -12.066528978641129 ], [ -77.079685482692227, -12.066531047919351 ], [ -77.079685898774599, -12.066532131309463 ], [ -77.079686237933402, -12.066533240367431 ], [ -77.079686498552874, -12.066534369809631 ], [ -77.079686679391429, -12.066535514255369 ], [ -77.079686779587519, -12.066536668252418 ], [ -77.079686798663815, -12.066537826303113 ], [ -77.079686736529439, -12.066538982890398 ], [ -77.079686370198161, -12.066541269667921 ], [ -77.079686067746493, -12.066542388963798 ], [ -77.079685687566268, -12.066543485059542 ], [ -77.079685231468687, -12.066544552733296 ], [ -77.079684701626618, -12.066545586898622 ], [ -77.079684100564251, -12.066546582628671 ], [ -77.079683431145099, -12.066547535179765 ], [ -77.07968269655828, -12.066548440013893 ], [ -77.079681900303413, -12.066549292820367 ], [ -77.079681046173903, -12.06655008953639 ], [ -77.079680138238857, -12.066550826366353 ], [ -77.079708333671789, -12.066580036517125 ], [ -77.079665693424971, -12.066619905794941 ], [ -77.079706202434622, -12.066661872659111 ], [ -77.079675529515853, -12.066692018600859 ], [ -77.079676518305035, -12.066692918663012 ], [ -77.079677877338128, -12.066694488362494 ], [ -77.079678470137807, -12.066695339338237 ], [ -77.079679001179869, -12.066696228939596 ], [ -77.079679467877185, -12.066697152832552 ], [ -77.079679867955988, -12.066698106515981 ], [ -77.079680199467177, -12.06669908534364 ], [ -77.079680460795657, -12.066700084546763 ], [ -77.079680650668237, -12.066701099257328 ], [ -77.079680768159903, -12.066702124531806 ], [ -77.079680812698228, -12.066703155375123 ], [ -77.079680682403421, -12.066705213677013 ], [ -77.079680508205072, -12.06670623110772 ], [ -77.079679945945699, -12.066708217768781 ], [ -77.079679560623944, -12.066709177320304 ], [ -77.079679108231858, -12.066710108080194 ], [ -77.07967859097343, -12.066711005513891 ], [ -77.079678011368699, -12.0667118652492 ], [ -77.079677372241449, -12.066712683097562 ], [ -77.079676676705418, -12.066713455074508 ], [ -77.0796759281492, -12.066714177419041 ], [ -77.079675130219698, -12.066714846611971 ], [ -77.079674649389617, -12.066715197582644 ], [ -77.079704036262115, -12.066745050376941 ], [ -77.079659439654279, -12.066785838693647 ], [ -77.079699996850366, -12.066827834419792 ], [ -77.079669351773276, -12.066856502359462 ], [ -77.079669704324232, -12.066856800583599 ], [ -77.079670463937106, -12.066857511659867 ], [ -77.07967117130147, -12.066858273155049 ], [ -77.07967182297115, -12.06685908135923 ], [ -77.079672415771242, -12.066859932334918 ], [ -77.07967294681373, -12.066860821936237 ], [ -77.079673413511387, -12.06686174582914 ], [ -77.079673813590531, -12.066862699512535 ], [ -77.07967414510199, -12.066863678340161 ], [ -77.079674596303491, -12.066865692253803 ], [ -77.079674758333766, -12.066867748371568 ], [ -77.079674628039044, -12.06686980667345 ], [ -77.079674453840667, -12.066870824104161 ], [ -77.079674207955378, -12.066871827096925 ], [ -77.079673891581095, -12.066872810765258 ], [ -77.079673506259198, -12.066873770316798 ], [ -77.079673053866912, -12.066874701076721 ], [ -77.079672536608243, -12.066875598510451 ], [ -77.079671957003214, -12.066876458245799 ], [ -77.079671317875622, -12.066877276094205 ], [ -77.079670622339236, -12.066878048071199 ], [ -77.07966987378262, -12.066878770415785 ], [ -77.079669075852664, -12.066879439608771 ], [ -77.079668182908193, -12.066880085668128 ], [ -77.079697775070954, -12.066910727426006 ], [ -77.079655305842948, -12.066950315325352 ], [ -77.079701530481401, -12.066998213723148 ], [ -77.079709635034831, -12.067089132359538 ], [ -77.079736281684461, -12.067152946967031 ], [ -77.079727862106751, -12.067443164377265 ], [ -77.079723118851746, -12.067662498033092 ], [ -77.079718764708005, -12.067827993839758 ], [ -77.079715768897742, -12.067961020193863 ], [ -77.079711127751125, -12.068089308700211 ], [ -77.079705922200489, -12.068352980460631 ], [ -77.079694207309046, -12.068711307846472 ], [ -77.079745173877726, -12.069257515764285 ], [ -77.079804173125083, -12.069896882818329 ], [ -77.079825276224682, -12.0698952379455 ], [ -77.0798703272594, -12.06984198206251 ], [ -77.080088755395863, -12.069821261805938 ], [ -77.080099802274404, -12.069942371179851 ], [ -77.080125702876174, -12.069963367721394 ], [ -77.080268236303638, -12.069950356127265 ], [ -77.080266741965659, -12.069933973496408 ], [ -77.080294355415518, -12.069931452711076 ], [ -77.080291850816181, -12.069919155921157 ], [ -77.080378310840928, -12.069818080168233 ], [ -77.080424395199699, -12.069819156806354 ], [ -77.080414638989382, -12.069707329544713 ], [ -77.080426475365911, -12.069708997935413 ], [ -77.080586862577462, -12.069695074225667 ], [ -77.080589295591338, -12.069724802330844 ], [ -77.080679769130029, -12.069716619236583 ], [ -77.080679083064766, -12.069707424088618 ], [ -77.080695362464198, -12.069705951656578 ], [ -77.08069536822461, -12.069715208336406 ], [ -77.080873212181245, -12.06969987510929 ], [ -77.080865752564435, -12.06962139140072 ], [ -77.080932888208181, -12.069615508883448 ], [ -77.080944419128841, -12.069611755924093 ], [ -77.081050099556691, -12.069602845997737 ], [ -77.081056440810386, -12.069666985464849 ], [ -77.08111780025061, -12.069678693497858 ], [ -77.081131711331579, -12.069833484520045 ], [ -77.081161903732024, -12.069830937124541 ], [ -77.081161508192309, -12.069826396861131 ], [ -77.081220522622573, -12.069821444240342 ], [ -77.081220879450669, -12.069825541324816 ], [ -77.081370648408793, -12.069812871312983 ], [ -77.08137029157848, -12.069808774229083 ], [ -77.081467545320621, -12.069800546806553 ], [ -77.08146234577498, -12.069740846446161 ], [ -77.081501869391232, -12.069737502839144 ], [ -77.081739712945691, -12.069776817780975 ], [ -77.081746366669933, -12.06977306708132 ], [ -77.081905939856782, -12.069721859006806 ], [ -77.081916133574708, -12.069719086665746 ], [ -77.081918585562818, -12.069718446779673 ], [ -77.081923595217347, -12.069717676776657 ], [ -77.081928664955512, -12.069717597476489 ], [ -77.081931191776064, -12.069717817841434 ], [ -77.081933696941988, -12.069718210409476 ], [ -77.081936168352527, -12.069718773284404 ], [ -77.081938594070053, -12.069719503747326 ], [ -77.081940962377573, -12.069720398269906 ], [ -77.08194326183542, -12.069721452531313 ], [ -77.081945481336547, -12.069722661439153 ], [ -77.081947610160029, -12.069724019154007 ], [ -77.081949638023048, -12.069725519117698 ], [ -77.081951555130374, -12.069727154084958 ], [ -77.081955020616775, -12.069730796826624 ], [ -77.081957939742637, -12.069734877081711 ], [ -77.081961173931973, -12.069741643619929 ], [ -77.082005252747507, -12.069873433499193 ], [ -77.082073751055475, -12.06990686040159 ], [ -77.081947150606794, -12.069523295433171 ], [ -77.08189760281131, -12.069356266528118 ], [ -77.081879188862558, -12.069297123306987 ], [ -77.081861943373411, -12.069235525878689 ], [ -77.081853641178256, -12.069204753276116 ], [ -77.081845384964211, -12.069173864207038 ], [ -77.081838780502125, -12.069142667219479 ], [ -77.081830911929501, -12.069111713521316 ], [ -77.081823209539692, -12.069080772232422 ], [ -77.081815634896614, -12.069049758224493 ], [ -77.081808051069473, -12.069018696748726 ], [ -77.08180063042181, -12.068987619691764 ], [ -77.08179341065572, -12.068956327376265 ], [ -77.081785663438751, -12.068925117717557 ], [ -77.081778908715521, -12.068894740637074 ], [ -77.081770623891018, -12.06886359423784 ], [ -77.081763902204031, -12.068832331614528 ], [ -77.081758819085422, -12.068800751657761 ], [ -77.081752379668657, -12.068769477351244 ], [ -77.08174465548656, -12.068738498273488 ], [ -77.081740065963288, -12.068706838487616 ], [ -77.081729206892135, -12.068668645317608 ], [ -77.081727927164508, -12.068644242921103 ], [ -77.081720576958972, -12.068613193399884 ], [ -77.081715051742606, -12.068581924838934 ], [ -77.0817113284334, -12.06855010661757 ], [ -77.081704390650614, -12.06851886928867 ], [ -77.081700962888078, -12.06848713835741 ], [ -77.081694276966559, -12.06845576936807 ], [ -77.081688910456663, -12.068423341243209 ], [ -77.081682542258406, -12.068369699871209 ], [ -77.081676573834287, -12.068337341169297 ], [ -77.08167218568812, -12.068305556214641 ], [ -77.081667675913806, -12.068273860542607 ], [ -77.081663462963618, -12.068242399686254 ], [ -77.081661162521598, -12.068210565936338 ], [ -77.08165574041081, -12.068179054795566 ], [ -77.081652196544667, -12.068147427104428 ], [ -77.081648709218939, -12.068115547157152 ], [ -77.081645056207535, -12.068083881472624 ], [ -77.081641848971003, -12.068052274360181 ], [ -77.081640376515608, -12.068020257219167 ], [ -77.081637381943594, -12.067988547167101 ], [ -77.081632853188111, -12.067956916561515 ], [ -77.081630098576696, -12.067925031198584 ], [ -77.081629122243996, -12.067893165713409 ], [ -77.081626763490576, -12.067861216271133 ], [ -77.081622611585132, -12.067829590442344 ], [ -77.081622140879659, -12.067797691773526 ], [ -77.081618507192289, -12.067765979502333 ], [ -77.081616849531812, -12.067734147206352 ], [ -77.081615068417477, -12.067702291487157 ], [ -77.081612255303654, -12.067665026975574 ], [ -77.081612137383487, -12.067638465494742 ], [ -77.08161243539233, -12.067606620892827 ], [ -77.081611426762208, -12.067574479912375 ], [ -77.081608708799067, -12.067542407164829 ], [ -77.081607952260597, -12.067514023641573 ], [ -77.081606944342397, -12.067470677391995 ], [ -77.081606175269627, -12.067426065927419 ], [ -77.081606928416448, -12.067355919867447 ], [ -77.081606332992934, -12.067232259761273 ], [ -77.081606236342409, -12.067174979485427 ], [ -77.081608200239657, -12.067093446322902 ], [ -77.081609853630809, -12.067066430724836 ], [ -77.081610599874907, -12.06700529561985 ], [ -77.081613519125185, -12.066947856890918 ], [ -77.081619441191663, -12.066890703697918 ], [ -77.081622328033305, -12.066833450864983 ], [ -77.081626700014198, -12.066776020045269 ], [ -77.081630184394712, -12.066718681326682 ], [ -77.081637753080201, -12.066650419748072 ], [ -77.081641615628911, -12.066621077600516 ], [ -77.081643187705083, -12.066603923959686 ], [ -77.081648556261925, -12.066546763434751 ], [ -77.081652852021477, -12.066493005545665 ], [ -77.081655768923355, -12.066466858270703 ], [ -77.081660935142338, -12.066432525108532 ], [ -77.081662855917486, -12.066416642352321 ], [ -77.081663803574983, -12.066408601327982 ], [ -77.081664767045723, -12.0663926058761 ], [ -77.081667809351899, -12.066369146138527 ], [ -77.081675051761238, -12.066318939618816 ], [ -77.081677804292411, -12.066298594750712 ], [ -77.081684950607084, -12.066251612878645 ], [ -77.081693045612994, -12.066205288849419 ], [ -77.081694786635808, -12.066188927268225 ], [ -77.081702150235628, -12.066148601616767 ], [ -77.081704975350291, -12.066126187359016 ], [ -77.081711490534062, -12.066091806965018 ], [ -77.081716741629037, -12.066064533122834 ], [ -77.081723727547711, -12.066018930939084 ], [ -77.081727811291685, -12.066004098612124 ], [ -77.081732151816126, -12.065976778338856 ], [ -77.081772418996479, -12.065761077523366 ], [ -77.081775003630682, -12.065752774762808 ], [ -77.081818792937469, -12.065527243920069 ], [ -77.081828712962221, -12.06547308522663 ], [ -77.081832726791845, -12.06545365753237 ], [ -77.081835011486561, -12.065435006613336 ], [ -77.081836507621759, -12.065416060186132 ], [ -77.08183627387524, -12.065396684737978 ], [ -77.081834279070563, -12.065377765926904 ], [ -77.081832025946227, -12.065358976240592 ], [ -77.081828182816466, -12.065340127443227 ], [ -77.081823013785865, -12.065321790323654 ], [ -77.081815558360006, -12.06530137669513 ], [ -77.081808717673511, -12.065285628471319 ], [ -77.081799805670272, -12.065268205531448 ], [ -77.081788923152942, -12.065250284711821 ], [ -77.08177756859142, -12.06523706915894 ], [ -77.081765431853981, -12.065221915699258 ], [ -77.081753565920067, -12.065207563089457 ], [ -77.081739442449646, -12.065192903082933 ], [ -77.081725153513858, -12.065181943890746 ], [ -77.081707864546658, -12.065168288591753 ], [ -77.081693498211777, -12.065159164576883 ], [ -77.081675883121449, -12.065149462605065 ], [ -77.081599202029778, -12.065122819085973 ], [ -77.08097926727244, -12.064974981978416 ], [ -77.080704142626232, -12.064908176633065 ], [ -77.079702118918092, -12.064666447034258 ] ] ] ] } 
    }
  ]
};

// ==========================================
// 6. Inventario de Tachos (CSV)
// ==========================================
var urlPuntosEcologicosCSV = "https://docs.google.com/spreadsheets/d/e/2PACX-1vQK3aVBwfr2kRO1b7mmOZ22PuVKSZwCTgAk8OUZxyh4ZfmDQq8IjW6kTHeIQJvnAw/pub?gid=657037007&single=true&output=csv";

var capaNoAprovechables     = L.layerGroup();
var capaPapelCarton         = L.layerGroup();
var capaPlastico            = L.layerGroup();
var capaVidrio              = L.layerGroup();
var capaPilas               = L.layerGroup();
var capaPeligrosos          = L.layerGroup();
var capaRAEE                 = L.layerGroup();
var capaMetales              = L.layerGroup();
var capaAniquem              = L.layerGroup();
var capaIntermediosPlastico = L.layerGroup();
var capaIntermediosMetal    = L.layerGroup();

var configuracionTachos = {
    "No Aprovechables":     { color: "#2D3748", capa: capaNoAprovechables,     nombreFila: "No Aprovechables" }, // Negro
    "Papel y Cartón":       { color: "#2B6CB0", capa: capaPapelCarton,         nombreFila: "Papel y Cartón" }, // Azul
    "Plástico":             { color: "#FFFFFF", capa: capaPlastico,            nombreFila: "Plástico", borde: "#CBD5E0" }, // Blanco
    "Vidrio":               { color: "#718096", capa: capaVidrio,              nombreFila: "Vidrio" }, // Plomo
    "Pilas":                { color: "#C53030", capa: capaPilas,               nombreFila: "Pilas" }, // Rojo
    "Peligrosos":           { color: "#C53030", capa: capaPeligrosos,          nombreFila: "Peligrosos" }, // Rojo
    "RAEE":                 { color: "#ED8936", capa: capaRAEE,                nombreFila: "RAEE" }, // Naranja
    "Metales":              { color: "#ECC94B", capa: capaMetales,             nombreFila: "Metales" }, // Amarillo
    "Aniquem":              { color: "#1A365D", capa: capaAniquem,             nombreFila: "Aniquem" }, // Azul Oscuro
    "Intermedios Plástico": { color: "#718096", capa: capaIntermediosPlastico, nombreFila: "Intermedios Plastico" }, // Plomo
    "Intermedios Metal":    { color: "#718096", capa: capaIntermediosMetal,    nombreFila: "Intermedios Metal" } // Plomo
};

// Array global para almacenar todos los registros de la base de datos de tachos
var listaEstacionesTachos = [];

// Variable global para filtrar por zona
var zonaSeleccionadaFiltro = "TODAS";

// Función auxiliar para calcular si un punto está dentro de las Zonas GeoJSON
function obtenerZonaDeCoordenada(lng, lat, geojson) {
    if (!geojson || !geojson.features) return "Sin Zona";

    function puntoEnPoligono(point, vs) {
        var x = point[0], y = point[1];
        var inside = false;
        for (var i = 0, j = vs.length - 1; i < vs.length; j = i++) {
            var xi = vs[i][0], yi = vs[i][1];
            var xj = vs[j][0], yj = vs[j][1];
            var intersect = ((yi > y) !== (yj > y)) && (x < (xj - xi) * (y - yi) / (yj - yi) + xi);
            if (intersect) inside = !inside;
        }
        return inside;
    }

    for (var i = 0; i < geojson.features.length; i++) {
        var feature = geojson.features[i];
        var nombreZona = feature.properties ? feature.properties.ZONA : null;
        var geom = feature.geometry;

        if (geom && nombreZona) {
            if (geom.type === "Polygon") {
                if (puntoEnPoligono([lng, lat], geom.coordinates[0])) return nombreZona;
            } else if (geom.type === "MultiPolygon") {
                for (var k = 0; k < geom.coordinates.length; k++) {
                    if (puntoEnPoligono([lng, lat], geom.coordinates[k][0])) return nombreZona;
                }
            }
        }
    }
    return "Sin Zona";
}

// Función auxiliar para transformar enlaces de Drive
function transformarEnlaceDrive(url) {
    if (!url) return "";
    var match = url.match(/\/d\/([a-zA-Z0-9_-]+)/) || url.match(/id=([a-zA-Z0-9_-]+)/);
    return match ? "https://lh3.googleusercontent.com/d/" + match[1] : url;
}

function generarHTMLIconoTacho(colorFondo, bordeColor) {
    var finalBorde = bordeColor || "#ffffff";
    var strokeSVG = colorFondo === "#FFFFFF" ? "#2D3748" : "#ffffff";
    return `
        <div style="background-color: ${colorFondo}; width: 18px; height: 18px; border-radius: 4px; display: inline-flex; align-items: center; justify-content: center; box-shadow: 0 2px 4px rgba(0,0,0,0.3); border: 1px solid ${finalBorde};">
            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="${strokeSVG}" stroke-width="3">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
            </svg>
        </div>
    `;
}

function crearIconoTacho(config) {
    return L.divIcon({
        className: 'tacho-marker-mapa',
        html: generarHTMLIconoTacho(config.color, config.borde),
        iconSize: [20, 20], 
        iconAnchor: [10, 10]
    });
}

// 🟢 FUNCIÓN QUE DESPLIEGA EL SIDEBAR LATERAL CON LA INFORMACIÓN DEL TACHO
function abrirSidebarTacho(datos) {
    var sidebar = document.getElementById('sidebar-info');
    var contenido = document.getElementById('sidebar-contenido');

    if (!sidebar || !contenido) {
        console.error("❌ Error: No se encontró el elemento #sidebar-info en el HTML.");
        return;
    }

    sidebar.style.zIndex = "10000";
    sidebar.style.position = "fixed";

    // Generar lista de tachos
    var listaTachosHTML = "<ul style='margin: 8px 0; padding-left: 20px; font-size: 14px; line-height: 1.6; color: #2d3748;'>";
    datos.tachosDisponibles.forEach(function(t) {
        listaTachosHTML += `<li><b>${t}</b></li>`;
    });
    listaTachosHTML += "</ul>";

    // Contenido dinámico para inyectar en el Sidebar
    contenido.innerHTML = `
        <div style="padding-top: 15px;">
            <span style="font-size:11px; text-transform:uppercase; color:#2b6cb0; font-weight:bold; letter-spacing:0.5px;">${datos.espacio} ${datos.zona ? `| 📍 ${datos.zona}` : ''}</span>
            <h2 style="margin: 4px 0 12px 0; color:#1a202c; font-size:20px; font-weight: bold;">♻️ ${datos.id}</h2>
            
            ${datos.urlFoto ? `
                <div style="width:100%; border-radius:10px; overflow:hidden; margin-bottom:16px; border:1px solid #cbd5e1; box-shadow:0 3px 6px rgba(0,0,0,0.1);">
                    <a href="${datos.urlFotoOriginal}" target="_blank" title="Ver imagen completa">
                        <img src="${datos.urlFoto}" style="width:100%; height:200px; object-fit:cover; display:block;" alt="Foto Estación ${datos.id}">
                    </a>
                </div>
            ` : ''}

            <div style="background:#f8fafc; border:1px solid #e2e8f0; border-radius:8px; padding:12px; margin-bottom:14px;">
                <span style="font-size:11px; text-transform:uppercase; color:#718096; font-weight:bold; display:block; margin-bottom:4px;">📍 Ubicación</span>
                <p style="margin:0; font-size:14px; color:#2d3748; font-weight:500;">${datos.lugar}</p>
            </div>

            <div style="background:#f8fafc; border:1px solid #e2e8f0; border-radius:8px; padding:12px; margin-bottom:14px;">
                <span style="font-size:11px; text-transform:uppercase; color:#718096; font-weight:bold; display:block; margin-bottom:4px;">🗑️ Tachos en esta estación</span>
                ${listaTachosHTML}
            </div>

            ${datos.nota ? `
                <div style="background:#fffaf0; border:1px dashed #dd6b20; border-radius:8px; padding:10px; font-size:12px; color:#c05621;">
                    <b>⚠️ Nota:</b> ${datos.nota}
                </div>
            ` : ''}
        </div>
    `;

    // Deslizar Sidebar hacia adentro
    sidebar.classList.remove('sidebar-oculto');
    sidebar.classList.add('sidebar-visible');
}

function cerrarSidebar() {
    var sidebar = document.getElementById('sidebar-info');
    if (sidebar) {
        sidebar.classList.remove('sidebar-visible');
        sidebar.classList.add('sidebar-oculto');
    }
}

// 📊 FUNCIÓN PARA FILTRAR POR ZONA
function filtrarTachosPorZona(zonaSeleccionada) {
    zonaSeleccionadaFiltro = zonaSeleccionada;

    listaEstacionesTachos.forEach(function(estacion) {
        var mostrar = (zonaSeleccionada === "TODAS" || estacion.zona === zonaSeleccionada);

        if (estacion.elementosMarcadores) {
            estacion.elementosMarcadores.forEach(function(item) {
                if (mostrar) {
                    if (!item.capa.hasLayer(item.marcador)) {
                        item.capa.addLayer(item.marcador);
                    }
                } else {
                    if (item.capa.hasLayer(item.marcador)) {
                        item.capa.removeLayer(item.marcador);
                    }
                }
            });
        }
    });

    actualizarResumenTachos();
}

// 📊 FUNCIÓN PARA ACTUALIZAR EL PANEL DE TOTALES DE TACHOS VISIBLES EN MAPA
function actualizarResumenTachos() {
    var panel = document.getElementById("panel-resumen-tachos");
    var contenedorLista = document.getElementById("lista-desglose-tachos");
    var elemTotalPuntos = document.getElementById("cant-total-puntos");
    var elemTotalTachos = document.getElementById("cant-total-tachos"); // Contenedor del Total

    if (!panel || !contenedorLista) return;

    var totalesPorTipo = {};
    Object.keys(configuracionTachos).forEach(function(clave) {
        totalesPorTipo[clave] = 0;
    });

    var puntosTotalesVisibles = 0;
    var totalTachosUnidades = 0; // Suma acumulada de todos los tachos
    var hayTachosVisibles = false;

    listaEstacionesTachos.forEach(function(estacion) {
        // Filtrar por selección de zona
        if (zonaSeleccionadaFiltro !== "TODAS" && estacion.zona !== zonaSeleccionadaFiltro) {
            return;
        }

        var estacionVisible = false;

        Object.keys(configuracionTachos).forEach(function(clave) {
            var conf = configuracionTachos[clave];
            var cant = estacion.cantidades[clave] || 0;

            if (cant > 0 && typeof map !== 'undefined' && map.hasLayer(conf.capa)) {
                totalesPorTipo[clave] += cant;
                totalTachosUnidades += cant; // Suma cada tacho
                estacionVisible = true;
                hayTachosVisibles = true;
            }
        });

        if (estacionVisible) {
            puntosTotalesVisibles++;
        }
    });

    // Inyectar el desglose
    var htmlDesglose = "";
    Object.keys(configuracionTachos).forEach(function(clave) {
        var conf = configuracionTachos[clave];
        var cantidad = totalesPorTipo[clave];

        if (cantidad > 0) {
            var bordeCss = conf.borde ? `border: 1px solid ${conf.borde};` : '';
            htmlDesglose += `
                <div class="item-tacho-row">
                    <span style="display: flex; align-items: center;">
                        <span class="indicador-color-tacho" style="background-color: ${conf.color}; ${bordeCss}"></span>
                        ${clave}
                    </span>
                    <strong>${cantidad}</strong>
                </div>
            `;
        }
    });

    contenedorLista.innerHTML = htmlDesglose;
    if (elemTotalPuntos) elemTotalPuntos.innerText = puntosTotalesVisibles;
    if (elemTotalTachos) elemTotalTachos.innerText = totalTachosUnidades; // Asigna el valor final

    // Mostrar/Ocultar el panel
    if (hayTachosVisibles) {
        panel.classList.remove("panel-resumen-oculto");
        panel.classList.add("panel-resumen-visible");
    } else {
        panel.classList.remove("panel-resumen-visible");
        panel.classList.add("panel-resumen-oculto");
    }
}

function cargarPuntosEcologicos() {
    Papa.parse(urlPuntosEcologicosCSV, {
        download: true, header: true, skipEmptyLines: true,
        complete: function(results) {
            results.data.forEach(function(fila) {
                var latRaw = fila["Latitude"] || ""; var lonRaw = fila["Longitude"] || "";
                if (latRaw && lonRaw) {
                    var lat = parseFloat(latRaw.toString().replace(",", "."));
                    var lon = parseFloat(lonRaw.toString().replace(",", "."));

                    if (!isNaN(lat) && !isNaN(lon)) {
                        var id = fila["ID"] || "Sin ID"; 
                        var nota = fila["Note"] || "";
                        var lugar = fila["Lugar"] || "Campus"; 
                        var espacio = fila["Espacios"] || "General";
                        
                        // Determinar la Zona evaluando la latitud y longitud contra el GeoJSON de supervisores
                        var zonaAsignada = obtenerZonaDeCoordenada(lon, lat, geojsonSupervisores);

                        // Búsqueda dinámica para evitar descuadres si se movió la columna de la foto
                        var llaveFoto = Object.keys(fila).find(k => k.trim().toLowerCase().includes("foto") || k.trim().toLowerCase().includes("imagen")) || "Foto";
                        var urlOriginalFoto = (fila[llaveFoto] || "").trim();
                        var urlProcesadaFoto = urlOriginalFoto.startsWith("http") ? transformarEnlaceDrive(urlOriginalFoto) : "";

                        var tachosEnEstePunto = [];
                        var todosLosDisponibles = [];
                        var cantidadesPorTipo = {};

                        Object.keys(configuracionTachos).forEach(function(clave) {
                            var conf = configuracionTachos[clave];
                            var cantidad = parseInt(fila[conf.nombreFila], 10) || 0;

                            cantidadesPorTipo[clave] = cantidad;

                            if (cantidad > 0) {
                                tachosEnEstePunto.push(conf);
                                var textoCantidad = cantidad > 1 ? `${clave} (x${cantidad})` : clave;
                                todosLosDisponibles.push(textoCantidad);
                            }
                        });

                        if (tachosEnEstePunto.length === 0) return;

                        var datosEstacion = {
                            id: id,
                            lugar: lugar,
                            espacio: espacio,
                            zona: zonaAsignada,
                            tachosDisponibles: todosLosDisponibles,
                            cantidades: cantidadesPorTipo,
                            urlFoto: urlProcesadaFoto,
                            urlFotoOriginal: urlOriginalFoto,
                            nota: (nota && nota.toLowerCase() !== "null" && nota.trim() !== "") ? nota : null,
                            elementosMarcadores: []
                        };

                        // Crear marcadores desplazados para cada tipo de tacho disponible en la estación
                        var distanciaSeparacion = 0.000035; 
                        var totalTachos = tachosEnEstePunto.length;

                        tachosEnEstePunto.forEach(function(conf, index) {
                            var offset = (index - (totalTachos - 1) / 2) * distanciaSeparacion;
                            var latDesplazada = lat;
                            var lonDesplazada = lon + offset;

                            var iconoOpcion = crearIconoTacho(conf);
                            var marcadorTacho = L.marker([latDesplazada, lonDesplazada], { icon: iconoOpcion });
                            
                            marcadorTacho.on('click', function(e) {
                                L.DomEvent.stopPropagation(e);
                                abrirSidebarTacho(datosEstacion);
                            });
                            
                            conf.capa.addLayer(marcadorTacho);

                            datosEstacion.elementosMarcadores.push({
                                marcador: marcadorTacho,
                                capa: conf.capa
                            });
                        });

                        listaEstacionesTachos.push(datosEstacion);
                    }
                }
            });

            // Registrar los eventos para recalcular los totales cuando el usuario activa/desactiva capas
            Object.keys(configuracionTachos).forEach(function(clave) {
                var conf = configuracionTachos[clave];
                conf.capa.on('add remove', function() {
                    actualizarResumenTachos();
                });
            });

            actualizarResumenTachos();
        }
    });
}


// ==========================================================================
// 6.B Inventario de Flora (Google Sheets + Conversión UTM Proj4js)
// ==========================================================================
var urlFloraCSV = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTp1v1vtF9EI7zefB-PyT_vHaDfIPckJp5h9cskot0TmmZuEHFF1n6kpI9mxMju1Ay2hlaRqqZy3JHr/pub?gid=730733478&single=true&output=csv";
var urlCafetosCSV = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTp1v1vtF9EI7zefB-PyT_vHaDfIPckJp5h9cskot0TmmZuEHFF1n6kpI9mxMju1Ay2hlaRqqZy3JHr/pub?gid=746966399&single=true&output=csv";

var capaPalmeras = L.layerGroup();
var capaCafetos = L.layerGroup(); 

var utm18S = "+proj=utm +zone=18 +south +datum=WGS84 +units=m +no_defs";
var wgs84 = "+proj=longlat +datum=WGS84 +no_defs";

function cargarPuntosFlora() {
    Papa.parse(urlFloraCSV, {
        download: true, 
        header: false, 
        skipEmptyLines: true,
        complete: function(results) {
            console.log("Datos brutos de flora recibidos:", results.data.length);
            
            results.data.forEach(function(fila, index) {
                if (index < 2) return; 

                if (!fila || fila.length < 4 || !fila[0] || fila[0].trim() === "" || fila[0].includes("TOTAL")) {
                    return; 
                }

                var id = fila[0].trim();
                if (isNaN(parseInt(id))) return;

                // Mapeo exhaustivo de tus columnas originales
                var lugar = (fila[1] || "").trim();               // Columna B: Lugar
                var rawLat = (fila[2] || "").trim();              // Columna C: Latitud
                var rawLng = (fila[3] || "").trim();              // Columna D: Longitud
                var utmNorte = (fila[4] || "-").trim();           // Columna E: Norte UTM
                var utmEste = (fila[5] || "-").trim();            // Columna F: Este UTM
                var nombreComun = (fila[6] || "Palmera").trim();   // Columna G: Nombre común
                var nombreCientifico = (fila[7] || "-").trim();   // Columna H: Nombre científico
                
                // Tipo de Planta (Columnas I, J, K)
                var esArbol = (fila[8] || "").trim().toLowerCase() !== "";
                var esPalmera = (fila[9] || "").trim().toLowerCase() !== "";
                var esArbusto = (fila[10] || "").trim().toLowerCase() !== "";
                var tipoPlanta = [];
                if (esArbol) tipoPlanta.push("🌳 Árbol");
                if (esPalmera) tipoPlanta.push("🌴 Palmera");
                if (esArbusto) tipoPlanta.push("🌿 Arbusto");

                // Medidas y Atributos (Columnas L a P)
                var alturatotal = (fila[11] || "-").trim();       // Columna L: altura (m)
                var alturafuste = (fila[12] || "-").trim();       // Columna M: altura del fuste
                var dap = (fila[13] || "-").trim();               // Columna N: DAP(cm)
                var radioCopa = (fila[14] || "-").trim();         // Columna O: Radio (m)
                var zunchado = (fila[15] || "No").trim();         // Columna P: Zunchado
                var medidaActual = (fila[16] || "Estable").trim(); // Columna Q: Estado Actual
                
                // Fotos y enlaces de Google Drive
                var fotoNombre = (fila[17] || "").trim();          // Columna R: Enlace de foto en Drive
                var linkUrl = (fila[18] || "").trim();            // Columna S: Enlace informativo (LINK)

                if (!rawLat || !rawLng) return;

                var strLat = rawLat.replace(/\s+/g, '').replace(',', '.');
                var strLng = rawLng.replace(/\s+/g, '').replace(',', '.');

                var latitud = parseFloat(strLat);
                var longitud = parseFloat(strLng);

                if (!isNaN(latitud) && !isNaN(longitud) && latitud < 0 && longitud < 0) {
                    
                    var urlProcesadaFlora = fotoNombre.startsWith("http") ? transformarEnlaceDrive(fotoNombre) : "";

                    // Contenedor principal del popup adaptado a 430px de ancho total para el diseño de dos columnas
                    var popupHTML = `
                        <div style="padding: 4px; font-size: 11px; width: 430px; line-height: 1.3; color: #2d3748; font-family: Arial, sans-serif; box-sizing: border-box;">
                            
                            <span style="font-size: 9px; text-transform: uppercase; color: #27ae60; font-weight: bold; display: block; margin-bottom: 2px;">🌴 Inventario de Flora (N° ${id})</span>
                            <h3 style="margin: 0 0 8px 0; font-size: 13px; color: #1a365d; border-bottom: 2px solid #27ae60; padding-bottom: 4px; font-weight: bold;">
                                ${lugar || "Ubicación General"}
                            </h3>

                            <div style="display: flex; flex-direction: row; align-items: flex-start; gap: 12px; width: 100%;">
                                
                                <div style="flex: 1;">
                                    <table style="width: 100%; border-collapse: collapse;">
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold; width: 40%;">Nombre Común:</td><td style="padding: 2px;">${nombreComun}</td></tr>
                                        <tr><td style="padding: 2px; font-weight: bold; font-style: italic;">Científico:</td><td style="padding: 2px; font-style: italic;">${nombreCientifico}</td></tr>
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold;">Tipo:</td><td style="padding: 2px;">${tipoPlanta.join(', ') || 'No especificado'}</td></tr>
                                        
                                        <tr><td colspan="2" style="padding: 4px 2px 2px 2px; font-weight: bold; color: #4a5568; font-size: 9px; text-transform: uppercase; border-bottom: 1px solid #edf2f7;">📏 Dimensiones y Estado</td></tr>
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold;">Alt. total / Fuste:</td><td style="padding: 2px;">${alturatotal} m / ${alturafuste} m</td></tr>
                                        <tr><td style="padding: 2px; font-weight: bold;">DAP / Copa:</td><td style="padding: 2px;">${dap} cm / ${radioCopa} m</td></tr>
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold;">Zunchado / Estado:</td><td style="padding: 2px;">${zunchado} / ${medidaActual}</td></tr>
                                        
                                        <tr><td colspan="2" style="padding: 4px 2px 2px 2px; font-weight: bold; color: #4a5568; font-size: 9px; text-transform: uppercase; border-bottom: 1px solid #edf2f7;">🌐 Coordenadas</td></tr>
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold;">Geográficas:</td><td style="padding: 2px; font-size: 10px; color: #718096;">${latitud.toFixed(5)}, ${longitud.toFixed(5)}</td></tr>
                                        <tr><td style="padding: 2px; font-weight: bold;">UTM (Z18S):</td><td style="padding: 2px; font-size: 10px; color: #718096;">N: ${utmNorte}<br>E: ${utmEste}</td></tr>
                                    </table>
                                </div>

                                ${urlProcesadaFlora !== "" ? `
                                    <div style="width: 160px; max-height: 200px; overflow-y: auto; overflow-x: hidden; border-radius: 6px; box-shadow: 0 1px 3px rgba(0,0,0,0.15); border: 1px solid #cbd5e0; background: #f7fafc; flex-shrink: 0;">
                                        <img src="${urlProcesadaFlora}" style="width: 100%; height: auto; display: block;" onload="this.closest('.leaflet-popup').querySelector('.leaflet-popup-content-wrapper').style.height='auto';">
                                    </div>
                                ` : ''}

                            </div>

                            ${linkUrl && linkUrl.startsWith("http") ? `
                                <div style="margin-top: 8px; text-align: center;">
                                    <a href="${linkUrl}" target="_blank" style="display: block; background-color: #1a365d; color: white; text-align: center; padding: 5px; border-radius: 4px; text-decoration: none; font-weight: bold; font-size: 11px; box-shadow: 0 1px 3px rgba(0,0,0,0.15);">
                                        🔗 Ver documento informativo
                                    </a>
                                </div>
                            ` : ''}
                        </div>
                    `;

                    var iconoPalmera = L.divIcon({
                        className: 'flora-marker-mapa', 
                        html: `<div style="font-size:26px; cursor:pointer; line-height:1;">🌴</div>`,
                        iconSize: [30, 30],
                        iconAnchor: [15, 25]
                    });

                    // Modificados los límites de tamaño para admitir un popup más ancho (450px)
                    L.marker([latitud, longitud], { icon: iconoPalmera })
                        .bindPopup(popupHTML, { maxWidth: 450, minWidth: 430 })
                        .addTo(capaPalmeras);
                }
            });
            console.log("¡Éxito! Estructura procesada de forma segura sin riesgos de bucle.");
        }
    });
}

function cargarPuntosCafetos() {
    Papa.parse(urlCafetosCSV, {
        download: true, 
        header: false, 
        skipEmptyLines: true,
        complete: function(results) {
            console.log("Filas totales leídas del Excel de Cafetos:", results.data.length);
            
            results.data.forEach(function(fila, index) {
                if (index < 3) return; 

                if (!fila || fila.length < 4) return;
                
                var id = (fila[0] || "").trim(); 
                if (id === "" || isNaN(parseInt(id))) return; 

                var lugar = (fila[1] || "Biblioteca Central").trim();        
                var rawLat = (fila[2] || "").trim();                         
                var rawLng = (fila[3] || "").trim();                         
                var utmNorte = (fila[4] || "-").trim();                      
                var utmEste = (fila[5] || "-").trim();                       
                var especie = (fila[6] || "Cafeto").trim();                  
                var nombreCientifico = (fila[7] || "Coffea arabica").trim(); 
                
                var esArbol = (fila[8] || "").trim().toLowerCase() !== "";
                var esPalmera = (fila[9] || "").trim().toLowerCase() !== "";
                var esArbusto = (fila[10] || "").trim().toLowerCase() !== "";
                var tipoPlanta = [];
                if (esArbol) tipoPlanta.push("🌳 Árbol");
                if (esPalmera) tipoPlanta.push("🌴 Palmera");
                if (esArbusto) tipoPlanta.push("🌿 Arbusto");

                var alturatotal = (fila[11] || "-").trim();       
                var alturafuste = (fila[12] || "-").trim();       
                var dap = (fila[13] || "-").trim();               
                var radioCopa = (fila[14] || "-").trim();         
                var zunchado = (fila[15] || "No").trim();         
                var medidaActual = (fila[16] || "Estable").trim(); 
                
                var fotoNombre = (fila[17] || "").trim();          
                var linkUrl = (fila[18] || "").trim();            

                if (!rawLat || !rawLng) return;

                var strLat = rawLat.replace(/\s+/g, '').replace(',', '.');
                var strLng = rawLng.replace(/\s+/g, '').replace(',', '.');

                var lat = parseFloat(strLat);
                var lon = parseFloat(strLng);

                if (id === "3" && lon > -10) {
                    lon = parseFloat(strLng.replace('.', '').replace('-', '-77.08039'));
                }

                if (!isNaN(lat) && !isNaN(lon) && lat < 0 && lon < 0) {
                    
                    var urlProcesadaCafeto = fotoNombre.startsWith("http") ? transformarEnlaceDrive(fotoNombre) : "";

                    var popupHTML = `
                        <div style="padding: 4px; font-size: 11px; width: 430px; line-height: 1.3; color: #2d3748; font-family: Arial, sans-serif; box-sizing: border-box;">
                            <span style="font-size: 9px; text-transform: uppercase; color: #e67e22; font-weight: bold; display: block; margin-bottom: 2px;">☕ Inventario Arbóreo (N° ${id})</span>
                            <h3 style="margin: 0 0 8px 0; font-size: 13px; color: #1a365d; border-bottom: 2px solid #e67e22; padding-bottom: 4px; font-weight: bold;">
                                ${lugar}
                            </h3>
                            <div style="display: flex; flex-direction: row; align-items: flex-start; gap: 12px; width: 100%;">
                                <div style="flex: 1;">
                                    <table style="width: 100%; border-collapse: collapse;">
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold; width: 40%;">Nombre Común:</td><td style="padding: 2px;">${especie}</td></tr>
                                        <tr><td style="padding: 2px; font-weight: bold; font-style: italic;">Científico:</td><td style="padding: 2px; font-style: italic;">${nombreCientifico}</td></tr>
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold;">Tipo:</td><td style="padding: 2px;">${tipoPlanta.join(', ') || 'No especificado'}</td></tr>
                                        <tr><td colspan="2" style="padding: 4px 2px 2px 2px; font-weight: bold; color: #4a5568; font-size: 9px; text-transform: uppercase; border-bottom: 1px solid #edf2f7;">📏 Dimensiones</td></tr>
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold;">Alt. total / Fuste:</td><td style="padding: 2px;">${alturatotal} m / ${alturafuste} m</td></tr>
                                        <tr><td style="padding: 2px; font-weight: bold;">DAP / Copa:</td><td style="padding: 2px;">${dap} cm / ${radioCopa} m</td></tr>
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold;">Zunchado / Estado:</td><td style="padding: 2px;">${zunchado} / ${medidaActual}</td></tr>
                                        <tr><td colspan="2" style="padding: 4px 2px 2px 2px; font-weight: bold; color: #4a5568; font-size: 9px; text-transform: uppercase; border-bottom: 1px solid #edf2f7;">🌐 Coordenadas</td></tr>
                                        <tr style="background-color: #f7fafc;"><td style="padding: 2px; font-weight: bold;">Geográficas:</td><td style="padding: 2px; font-size: 10px; color: #718096;">${lat.toFixed(5)}, ${lon.toFixed(5)}</td></tr>
                                        <tr><td style="padding: 2px; font-weight: bold;">UTM (Z18S):</td><td style="padding: 2px; font-size: 10px; color: #718096;">N: ${utmNorte}<br>E: ${utmEste}</td></tr>
                                    </table>
                                </div>
                                ${urlProcesadaCafeto !== "" ? `
                                    <div style="width: 160px; max-height: 200px; overflow-y: auto; overflow-x: hidden; border-radius: 6px; box-shadow: 0 1px 3px rgba(0,0,0,0.15); border: 1px solid #cbd5e0; background: #f7fafc; flex-shrink: 0;">
                                        <img src="${urlProcesadaCafeto}" style="width: 100%; height: auto; display: block;" onload="this.closest('.leaflet-popup').querySelector('.leaflet-popup-content-wrapper').style.height='auto';">
                                    </div>
                                ` : ''}
                            </div>
                            ${linkUrl && linkUrl.startsWith("http") ? `
                                <div style="margin-top: 8px; text-align: center;">
                                    <a href="${linkUrl}" target="_blank" style="display: block; background-color: #1a365d; color: white; text-align: center; padding: 5px; border-radius: 4px; text-decoration: none; font-weight: bold; font-size: 11px; box-shadow: 0 1px 3px rgba(0,0,0,0.15);">
                                        🔗 Ver documento informativo
                                    </a>
                                </div>
                            ` : ''}
                        </div>
                    `;

                    var iconoCafeto = L.divIcon({
                        className: 'cafeto-marker-custom', 
                        html: `<div style="font-size:24px; cursor:pointer; filter: drop-shadow(1px 1px 1px rgba(0,0,0,0.4)); line-height:1;">🌱</div>`,
                        iconSize: [30, 30],
                        iconAnchor: [15, 25]
                    });

                    L.marker([lat, lon], { icon: iconoCafeto })
                        .bindPopup(popupHTML, { maxWidth: 450, minWidth: 430 })
                        .addTo(capaCafetos);
                }
            });
            console.log("¡Éxito! Cafetos inyectados correctamente.");
        }
    });
}


// ==========================================================================
// 7. Monitoreo PUCP - Solución Completa con Fotos, Filtros y Fichas de Registro
// ==========================================================================

const BASE_URL = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTakay8F0nlgM_t333fY-rOIG0auZ_vMh4-q0_d0_FJI9kS-bVv03Q0MxxgeX7XV9tz7jjPYJdMGI73/pub?single=true&output=csv";

const URL_LUGARES_CSV     = `${BASE_URL}&gid=216669177`;
const URL_ACTIVIDADES_CSV = `${BASE_URL}&gid=344271829`;
const URL_PODA_CSV        = `${BASE_URL}&gid=1845362857`;
const URL_2026_CSV        = `${BASE_URL}&gid=530107837`; // Nueva hoja 2026


let diccionarioLugares = {}; 
let mapaClasesYTipos = {}; 
let listaTiposPorClase = {}; 
let datosPodaGuardados = []; 

const capaMarkers = L.layerGroup();
const capaCalorActividades = L.heatLayer([], { 
    radius: 45, blur: 15, max: 1.0,
    gradient: { 0.2: 'blue', 0.4: 'lime', 0.6: 'yellow', 0.8: 'orange', 1.0: 'red' }
});

const NOMBRES_MESES = [
    "enero", "febrero", "marzo", "abril", "mayo", "junio",
    "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"
];

// --------------------------------------------------------------------------
// Helper: Normalización y Extracción de Datos
// --------------------------------------------------------------------------
function obtenerIdDrive(url = "") {
    if (!url) return "";
    const str = url.toString().trim();
    const match = str.match(/\/d\/([a-zA-Z0-9_-]+)/) || 
                  str.match(/[?&]id=([a-zA-Z0-9_-]+)/) ||
                  str.match(/file\/d\/([a-zA-Z0-9_-]+)/);
    return (match && match[1]) ? match[1] : "";
}

function transformarEnlaceDrive(url = "") {
    if (!url) return "";
    const str = url.toString().trim();
    const id = obtenerIdDrive(str);
    
    if (id) {
        return `https://drive.google.com/thumbnail?id=${id}&sz=w800`;
    }
    return str;
}

function normalizarTexto(str = "") {
    return str.toString()
        .trim()
        .toLowerCase()
        .normalize("NFD")
        .replace(/[\u0300-\u036f]/g, "") 
        .replace(/[^a-z0-9]/g, "");     
}

function obtenerNombreMesDeFecha(cadenaFecha = "") {
    if (!cadenaFecha) return "";
    const texto = cadenaFecha.toLowerCase();
    
    for (const mes of NOMBRES_MESES) {
        if (texto.includes(mes)) return mes;
    }

    const partes = texto.match(/(\d{1,2})[\/\-](\d{1,2})[\/\-](\d{2,4})/);
    if (partes && partes[2]) {
        const numMes = parseInt(partes[2], 10);
        if (numMes >= 1 && numMes <= 12) {
            return NOMBRES_MESES[numMes - 1];
        }
    }
    return "";
}

function limpiarNombrePersonal(cadena = "") {
    if (!cadena) return "";
    let limpio = cadena.replace(/Personal\s*a\s*cargo/gi, "").trim();
    if (!limpio) return "";

    const coincideRepetido = limpio.match(/^([A-ZÁÉÍÓÚa-záéíóúñÑ]+?)\1+$/i);
    if (coincideRepetido && coincideRepetido[1]) {
        limpio = coincideRepetido[1];
    } else {
        const nombres = limpio.match(/[A-ZÁÉÍÓÚ][a-záéíóúñÑ]+/g);
        if (nombres && nombres.length > 0) {
            limpio = nombres[0];
        }
    }
    return limpio.charAt(0).toUpperCase() + limpio.slice(1).toLowerCase();
}

function obtenerCoordenadas(ubicacionFila) {
    const normFila = normalizarTexto(ubicacionFila);
    if (!normFila) return null;

    if (diccionarioLugares[normFila]) {
        return diccionarioLugares[normFila];
    }

    for (const [lugarNorm, coords] of Object.entries(diccionarioLugares)) {
        if (lugarNorm.includes(normFila) || normFila.includes(lugarNorm)) {
            return coords;
        }
    }

    return null;
}

function obtenerElementosSelect() {
    const selects = document.querySelectorAll(".leaflet-left select, div select");
    return {
        mes:         document.getElementById("selectMes")         || selects[0],
        clase:       document.getElementById("selectClase")       || selects[1],
        tipo:        document.getElementById("selectTipo")        || selects[2],
        responsable: document.getElementById("selectResponsable") || selects[3]
    };
}

function vincularEscuchadoresCheckboxes() {
    const chkMarcadores = document.getElementById('chk-actividades');
    const chkCalor = document.getElementById('chk-calor-actividades');

    if (chkMarcadores && !chkMarcadores.dataset.listener) {
        chkMarcadores.addEventListener("change", cargarDatosEnMapa);
        chkMarcadores.dataset.listener = "true";
    }
    if (chkCalor && !chkCalor.dataset.listener) {
        chkCalor.addEventListener("change", cargarDatosEnMapa);
        chkCalor.dataset.listener = "true";
    }
}

function actualizarSelectTipos(claseSeleccionada) {
    const { tipo: selectTipo } = obtenerElementosSelect();
    if (!selectTipo) return;

    selectTipo.innerHTML = '<option value="">Todos los tipos</option>';
    const claseNorm = normalizarTexto(claseSeleccionada);
    
    const conjuntoTipos = new Set();

    if (claseNorm && listaTiposPorClase[claseNorm]) {
        listaTiposPorClase[claseNorm].forEach(t => t && conjuntoTipos.add(t));
    } else if (!claseNorm) {
        Object.values(listaTiposPorClase).flat().forEach(t => t && conjuntoTipos.add(t));
    }

    Array.from(conjuntoTipos).sort().forEach(tipo => {
        const opt = document.createElement("option");
        opt.value = tipo;
        opt.textContent = tipo;
        selectTipo.appendChild(opt);
    });
}

function poblarFiltroResponsables(responsablesSet) {
    const { responsable: selectResp } = obtenerElementosSelect();
    if (!selectResp) return;

    selectResp.innerHTML = '<option value="">Todos los responsables</option>';
    Array.from(responsablesSet).sort().forEach(resp => {
        if (!resp) return;
        const opt = document.createElement("option");
        opt.value = resp;
        opt.textContent = resp;
        selectResp.appendChild(opt);
    });

    if (!selectResp.dataset.listener) {
        selectResp.addEventListener("change", cargarDatosEnMapa);
        selectResp.dataset.listener = "true";
    }
}

function poblarFiltrosClaseYTipo(clasesSet) {
    const { mes: selectMes, clase: selectClase, tipo: selectTipo } = obtenerElementosSelect();

    if (selectClase) {
        selectClase.innerHTML = '<option value="">Todas las clases</option>';
        Array.from(clasesSet).sort().forEach(clase => {
            if (!clase || /^PO\d+/i.test(clase)) return;

            const opt = document.createElement("option");
            opt.value = clase;
            opt.textContent = clase;
            selectClase.appendChild(opt);
        });

        if (!selectClase.dataset.listener) {
            selectClase.addEventListener("change", function() {
                const { tipo: sTipo } = obtenerElementosSelect();
                if (sTipo) sTipo.value = ""; 

                actualizarSelectTipos(this.value); 
                cargarDatosEnMapa();
            });
            selectClase.dataset.listener = "true";
        }
    }

    if (selectTipo && !selectTipo.dataset.listener) {
        selectTipo.addEventListener("change", cargarDatosEnMapa);
        selectTipo.dataset.listener = "true";
    }

    if (selectMes && !selectMes.dataset.listener) {
        selectMes.addEventListener("change", cargarDatosEnMapa);
        selectMes.dataset.listener = "true";
    }
}

// Promesa para parsear CSV ordenadamente
function parsearCSV(url, options = {}) {
    return new Promise((resolve) => {
        Papa.parse(url, {
            download: true,
            skipEmptyLines: true,
            ...options,
            complete: function(res) {
                resolve(res);
            }
        });
    });
}

// --------------------------------------------------------------------------
// Carga de Datos desde Google Sheets (CSV)
// --------------------------------------------------------------------------
async function inicializarDescargaExcel() {
    if (typeof Papa === 'undefined') {
        console.error("❌ ERROR: PapaParse no está cargado.");
        return;
    }

    datosPodaGuardados = [];
    const responsablesSet = new Set();
    const conjuntoClases = new Set();

    // 1. Cargar Lugares y Coordenadas
    const resLugares = await parsearCSV(URL_LUGARES_CSV, { header: true });
    diccionarioLugares = {};
    (resLugares.data || []).forEach(row => {
        const lugar = (row["lugar"] || row["Lugar"] || row["UBICACION"] || row["Ubicación"] || row["Nombre"] || "").trim();
        const latRaw = row["latitud"] || row["Latitud"] || row["LATITUD"] || row["lat"] || row["Y"] || "";
        const lonRaw = row["longitud"] || row["Longitud"] || row["LONGITUD"] || row["lon"] || row["X"] || "";

        if (lugar && latRaw && lonRaw) {
            const lat = parseFloat(latRaw.toString().replace(",", "."));
            const lon = parseFloat(lonRaw.toString().replace(",", "."));
            if (!isNaN(lat) && !isNaN(lon)) {
                diccionarioLugares[normalizarTexto(lugar)] = [lat, lon];
            }
        }
    });

    // 2. Cargar Clases y Tipos Oficiales
    mapaClasesYTipos = {};
    listaTiposPorClase = {};
    const resActividades = await parsearCSV(URL_ACTIVIDADES_CSV, { header: false });
    let claseActual = ""; 
    const filasAct = resActividades.data || [];

    for (let i = 0; i < filasAct.length; i++) {
        const fila = filasAct[i];
        if (!fila || fila.length < 2) continue;

        const colA = fila[0] ? fila[0].toString().trim() : "";
        const colB = fila[1] ? fila[1].toString().trim() : "";

        if (normalizarTexto(colA).includes("clase") && normalizarTexto(colB).includes("tipo")) {
            continue;
        }

        if (colA !== "" && !/^PO\d+/i.test(colA)) {
            claseActual = colA;
        }

        if (colB !== "" && claseActual !== "") {
            const tipoNorm = normalizarTexto(colB);
            const claseNorm = normalizarTexto(claseActual);

            mapaClasesYTipos[tipoNorm] = claseActual;
            conjuntoClases.add(claseActual);

            if (!listaTiposPorClase[claseNorm]) {
                listaTiposPorClase[claseNorm] = [];
            }
            if (!listaTiposPorClase[claseNorm].includes(colB)) {
                listaTiposPorClase[claseNorm].push(colB);
            }
        }
    }

    // 3. Cargar Podas
    const resPoda = await parsearCSV(URL_PODA_CSV, { header: false });
    const filasPoda = resPoda.data || [];

    if (filasPoda.length >= 3) {
        const encabezados = filasPoda[1].map(h => normalizarTexto(h));

        const idxId           = encabezados.findIndex(h => h.includes("id") || h.includes("incidencia"));
        const idxTipo         = encabezados.findIndex(h => h.includes("tipo"));
        const idxFechaRep     = encabezados.findIndex(h => h.includes("fechadereporte") || h.includes("reporte"));
        const idxFechaEjec    = encabezados.findIndex(h => h.includes("fechadeejecucion") || h.includes("ejecucion"));
        const idxPersonal     = encabezados.findIndex(h => h.includes("personal") || h.includes("cargo"));
        const idxUbicacion    = encabezados.findIndex(h => h.includes("ubicacion"));
        const idxNombreComun  = encabezados.findIndex(h => h.includes("nombrecomun"));
        const idxNombreCient  = encabezados.findIndex(h => h.includes("nombrecientifico"));
        const idxFoto         = encabezados.findIndex(h => h.includes("foto"));
        const idxLinkRegistro = encabezados.findIndex(h => h.includes("linkderegistro") || h.includes("link"));

        for (let i = 2; i < filasPoda.length; i++) {
            const fila = filasPoda[i];
            if (!fila || fila.length === 0) continue;

            const getVal = (idx, fallbackIdx) => {
                if (idx !== -1 && fila[idx] !== undefined) return fila[idx].toString().trim();
                if (fallbackIdx !== undefined && fila[fallbackIdx] !== undefined) return fila[fallbackIdx].toString().trim();
                return "";
            };

            const tipoVal = getVal(idxTipo, 2);
            const tipoNorm = normalizarTexto(tipoVal);
            
            // Determinar clase basada en la tabla maestro
            const claseCorrespondiente = mapaClasesYTipos[tipoNorm] || "Poda";

            const registro = {};
            registro["_id"]               = getVal(idxId, 0);
            registro["_tipo"]              = tipoVal;
            registro["_clase"]             = claseCorrespondiente;
            registro["_fechaReporte"]      = getVal(idxFechaRep, 4);
            registro["_fechaEjecucion"]    = getVal(idxFechaEjec, 5);
            registro["_personal"]          = limpiarNombrePersonal(getVal(idxPersonal, 7));
            registro["_ubicacion"]         = getVal(idxUbicacion, 8);
            registro["_nombreComun"]       = getVal(idxNombreComun, 13); 
            registro["_nombreCientifico"]  = getVal(idxNombreCient, 14); 
            registro["_foto"]              = getVal(idxFoto, 15);         
            registro["_linkRegistro"]      = getVal(idxLinkRegistro, 18); 

            if (registro["_personal"]) {
                responsablesSet.add(registro["_personal"]);
            }

            if (registro["_ubicacion"]) {
                datosPodaGuardados.push(registro);
            }
        }
    }

    // 4. Cargar Hoja 2026 (gid=530107837)
    const res2026 = await parsearCSV(URL_2026_CSV, { header: true });
    const filas2026 = res2026.data || [];

    filas2026.forEach((fila, idx) => {
        const actividadVal   = (fila["Actividad"] || fila["actividad"] || fila["Actividad/Tipo"] || "").trim();
        const tipoNorm       = normalizarTexto(actividadVal);

        // VALIDACIÓN: Si el tipo/actividad NO pertenece a las clases registradas oficiales, se salta esta fila.
        if (!actividadVal || !mapaClasesYTipos[tipoNorm]) {
            return; // SALTAR FILA
        }

        const claseOficial   = mapaClasesYTipos[tipoNorm]; // Clase limpia mapeada desde la tabla oficial
        const fechaVal       = (fila["fecha_atención"] || fila["fecha_atencion"] || fila["Fecha"] || "").trim();
        const mesVal         = (fila["Mes"] || fila["mes"] || "").trim();
        const lugarVal       = (fila["lugar"] || fila["Lugar"] || "").trim();
        const comentarioVal  = (fila["comentario"] || fila["Comentario"] || "").trim();
        const fotoVal        = (fila["Foto"] || fila["foto"] || "").trim();
        const responsableVal = limpiarNombrePersonal((fila["Responsable"] || fila["responsable"] || "").trim());
        const detalleVal     = (fila["Detalle"] || fila["detalle"] || "").trim();
        const linkVal        = (fila["Link"] || fila["link"] || fila["LinkRegistro"] || fila["link_registro"] || "").trim();

        if (lugarVal) {
            const registro = {
                "_id": `2026-${idx + 1}`,
                "_clase": claseOficial,
                "_tipo": actividadVal,
                "_fechaReporte": fechaVal || mesVal,
                "_fechaEjecucion": fechaVal,
                "_personal": responsableVal,
                "_ubicacion": lugarVal,
                "_comentario": comentarioVal,
                "_detalle": detalleVal,
                "_foto": fotoVal,
                "_linkRegistro": linkVal
            };

            if (responsableVal) responsablesSet.add(responsableVal);
            datosPodaGuardados.push(registro);
        }
    });


    // Rellenar filtros únicamente con las clases base oficiales
    poblarFiltrosClaseYTipo(conjuntoClases);
    poblarFiltroResponsables(responsablesSet);
    diagnosticarYRenderizar();
}

function diagnosticarYRenderizar() {
    vincularEscuchadoresCheckboxes();
    if (Object.keys(diccionarioLugares).length === 0 || datosPodaGuardados.length === 0) return;

    const { clase: selectClase } = obtenerElementosSelect();
    if (selectClase && selectClase.value) {
        actualizarSelectTipos(selectClase.value);
    }

    cargarDatosEnMapa();
}

// --------------------------------------------------------------------------
// Dibujado de Capas en el Mapa
// --------------------------------------------------------------------------
function cargarDatosEnMapa() {
    capaMarkers.clearLayers();
    vincularEscuchadoresCheckboxes();

    const chkMarcadores = document.getElementById('chk-actividades');
    const chkCalor = document.getElementById('chk-calor-actividades');

    if (typeof map !== 'undefined') {
        if (chkMarcadores && chkMarcadores.checked) {
            if (!map.hasLayer(capaMarkers)) capaMarkers.addTo(map);
        } else {
            if (map.hasLayer(capaMarkers)) map.removeLayer(capaMarkers);
        }

        if (chkCalor && chkCalor.checked) {
            if (!map.hasLayer(capaCalorActividades)) capaCalorActividades.addTo(map);
        } else {
            if (map.hasLayer(capaCalorActividades)) map.removeLayer(capaCalorActividades);
        }
    }

    const { mes: selectMes, clase: selectClase, tipo: selectTipo, responsable: selectResp } = obtenerElementosSelect();

    const mesFiltro          = selectMes ? normalizarTexto(selectMes.value) : "";
    const claseFiltro        = selectClase ? normalizarTexto(selectClase.value) : "";
    const tipoFiltro         = selectTipo ? normalizarTexto(selectTipo.value) : "";
    const responsableFiltro = selectResp ? normalizarTexto(selectResp.value) : "";

    const puntosCalor = [];
    const mapaUbicacionesAgrupadas = {};

    datosPodaGuardados.forEach(fila => {
        const claseRegistro = normalizarTexto(fila["_clase"]);
        const tipoRegistro  = normalizarTexto(fila["_tipo"]);
        const respRegistro  = normalizarTexto(fila["_personal"]);

        if (claseFiltro !== "" && !claseFiltro.includes("todas") && claseRegistro !== claseFiltro) return;
        if (tipoFiltro !== "" && !tipoFiltro.includes("todos") && tipoRegistro !== tipoFiltro) return;
        if (responsableFiltro !== "" && !responsableFiltro.includes("todos") && respRegistro !== responsableFiltro) return;

        if (mesFiltro !== "" && !mesFiltro.includes("todos")) {
            const mesNombre = obtenerNombreMesDeFecha(fila["_fechaReporte"]);
            if (mesNombre !== mesFiltro) return;
        }

        const ubicacionRaw = fila["_ubicacion"].toString().trim();
        const coords = obtenerCoordenadas(ubicacionRaw);
        if (!coords) return;

        const [lat, lon] = coords;
        puntosCalor.push([lat, lon, 0.8]);

        const keyCoord = `${lat},${lon}`;
        if (!mapaUbicacionesAgrupadas[keyCoord]) {
            mapaUbicacionesAgrupadas[keyCoord] = {
                lat,
                lon,
                ubicacionNombre: ubicacionRaw,
                actividades: []
            };
        }
        mapaUbicacionesAgrupadas[keyCoord].actividades.push(fila);
    });

    // Construcción de Popups
    Object.values(mapaUbicacionesAgrupadas).forEach(grupo => {
        const total = grupo.actividades.length;

        let popupHTML = `
            <div style="font-family: Arial, sans-serif; font-size:12px; color:#2d3748; width: 400px;">
                <div style="margin-bottom:8px; padding-bottom:4px; border-bottom:2px solid #cbd5e0; font-weight:bold; color:#1a365d; font-size:13px;">
                    📍 ${grupo.ubicacionNombre} 
                    <span style="background:#3182ce; color:white; border-radius:12px; padding:2px 8px; font-size:10px; float:right;">
                        ${total} ${total === 1 ? 'actividad' : 'actividades'}
                    </span>
                </div>
                <div style="max-height: 280px; overflow-y: auto; padding-right: 4px;">
        `;

        grupo.actividades.forEach(fila => {
            const rawFoto = (fila["_foto"] || "").toString().trim();
            const driveId = obtenerIdDrive(rawFoto);
            const esFotoValida = rawFoto.startsWith("http") || driveId !== "";
            const fotoProcesada = esFotoValida ? transformarEnlaceDrive(rawFoto) : "";

            const linkDoc = (fila["_linkRegistro"] || "").toString().trim();
            const esDocValido = linkDoc.length > 3;

            popupHTML += `
                <div style="margin-bottom: 10px; padding: 8px; background: #f8fafc; border: 1px solid #e2e8f0; border-left: 4px solid #3182ce; border-radius: 6px;">
                    
                    <div style="display: flex; gap: 10px; align-items: flex-start;">
                        
                        <!-- Información del registro -->
                        <div style="flex: 1; min-width: 0;">
                            <h4 style="margin:0 0 4px 0; font-size:12px; color:#1a365d; word-break: break-word;">
                                📌 ${fila["_id"]} - ${fila["_tipo"]}
                            </h4>
                            <p style="margin:2px 0;">
                                <span style="background:#e2e8f0; color:#2d3748; padding:1px 5px; border-radius:3px; font-size:9px; font-weight:bold;">
                                    CLASE: ${fila["_clase"]}
                                </span>
                            </p>
                            ${fila["_fechaEjecucion"] ? `<p style="margin:2px 0;"><b>Ejecución:</b> ${fila["_fechaEjecucion"]}</p>` : ''}
                            ${fila["_comentario"] ? `<p style="margin:2px 0;"><b>Comentario:</b> ${fila["_comentario"]}</p>` : ''}
                            ${fila["_detalle"] ? `<p style="margin:2px 0;"><b>Detalle:</b> ${fila["_detalle"]}</p>` : ''}
                            ${fila["_nombreComun"] ? `<p style="margin:2px 0;"><b>Especie:</b> ${fila["_nombreComun"]} <i>(${fila["_nombreCientifico"]})</i></p>` : ''}
                            ${fila["_personal"] ? `<p style="margin:2px 0;"><b>Responsable:</b> ${fila["_personal"]}</p>` : ''}
                            
                            <!-- Botón de Link de Registro / Ficha Técnica -->
                            ${esDocValido ? `
                                <div style="margin-top: 8px;">
                                    <a href="${linkDoc.startsWith('http') ? linkDoc : 'https://drive.google.com/file/d/' + obtenerIdDrive(linkDoc)}" 
                                       target="_blank" 
                                       style="display: inline-block; background: #2b6cb0; color: #ffffff; padding: 4px 8px; border-radius: 4px; text-decoration: none; font-weight: bold; font-size: 10px; box-shadow: 0 1px 2px rgba(0,0,0,0.1);">
                                        📄 Ver Ficha / Documento ➔
                                    </a>
                                </div>
                            ` : ''}
                        </div>

                        <!-- Foto del registro -->
                        ${esFotoValida ? `
                            <div style="width: 120px; flex-shrink: 0; text-align: center;">
                                <div style="border-radius: 6px; overflow: hidden; border: 1px solid #cbd5e0; background: #edf2f7; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
                                    <a href="${rawFoto.startsWith('http') ? rawFoto : 'https://drive.google.com/file/d/' + driveId}" target="_blank">
                                        <img src="${fotoProcesada}" 
                                             style="width: 100%; height: 85px; object-fit: cover; display: block;" 
                                             onerror="if('${driveId}'){ this.onerror=null; this.src='https://lh3.googleusercontent.com/d/${driveId}=s800'; }"
                                             onload="if(this.closest('.leaflet-popup-content-wrapper')) this.closest('.leaflet-popup-content-wrapper').style.height='auto';" />
                                    </a>
                                </div>
                                <a href="${rawFoto.startsWith('http') ? rawFoto : 'https://drive.google.com/file/d/' + driveId}" target="_blank" style="display: block; font-size: 10px; color: #2b6cb0; margin-top: 3px; font-weight: bold; text-decoration: none;">
                                    🔍 Ampliar Foto
                                </a>
                            </div>
                        ` : ''}

                    </div>
                </div>
            `;
        });

        popupHTML += `
                </div>
            </div>
        `;

        const marker = L.marker([grupo.lat, grupo.lon]).bindPopup(popupHTML, { maxWidth: 450, minWidth: 410 });
        capaMarkers.addLayer(marker);
    });

    capaCalorActividades.setLatLngs(puntosCalor);
    if (typeof capaCalorActividades.redraw === 'function') {
        capaCalorActividades.redraw();
    }
}

function cargarDatos() {
    cargarDatosEnMapa();
}

// Iniciar aplicación
inicializarDescargaExcel();


// ==========================================================================
// MÓDULO VIVERO - AGRUPACIÓN POR LUGAR Y FILTRO POR MESES
// ==========================================================================

const URL_VIVERO_CSV = "https://docs.google.com/spreadsheets/d/e/2PACX-1vRrEaaf7wWTRZl4Z0uuolTuger7oSUH5jN9EPIT2X0nW5N2gbxrifSTRCcSaXx1_Q/pub?gid=1814328638&single=true&output=csv";

const capaVivero = L.layerGroup();
let datosVivero = []; // Guardará las filas procesadas

const MESES_VIVERO = [
    "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
    "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"
];

// Helper para extraer el mes de la cadena de texto de la columna 'Fecha'
function obtenerMesDeFecha(fechaStr) {
    if (!fechaStr) return "Sin Fecha";
    const txt = fechaStr.trim().toLowerCase();

    for (let i = 0; i < MESES_VIVERO.length; i++) {
        if (txt.includes(MESES_VIVERO[i].toLowerCase())) {
            return MESES_VIVERO[i];
        }
    }

    const partes = fechaStr.split(/[\/\-]/);
    if (partes.length >= 2) {
        let mesNum = null;
        if (partes[0].length === 4) {
            mesNum = parseInt(partes[1], 10);
        } else {
            mesNum = parseInt(partes[1], 10);
        }
        if (!isNaN(mesNum) && mesNum >= 1 && mesNum <= 12) {
            return MESES_VIVERO[mesNum - 1];
        }
    }

    return "Otros";
}

function cargarVivero() {
    if (typeof Papa === 'undefined') {
        console.error("PapaParse no está cargado.");
        return;
    }

    Papa.parse(URL_VIVERO_CSV, {
        download: true,
        header: true,
        skipEmptyLines: 'greedy',
        complete: function(res) {
            datosVivero = res.data || [];

            poblarDesplegableMeses(datosVivero);
            renderizarMarcadoresVivero();

            console.log(`✅ Módulo Vivero: ${datosVivero.length} registros cargados.`);
        }
    });
}

function poblarDesplegableMeses(filas) {
    const cboMes = document.getElementById('cbo-vivero-mes');
    if (!cboMes) return;

    const mesesEncontrados = new Set();
    filas.forEach(fila => {
        const fecha = (fila["Fecha"] || "").trim();
        if (fecha) {
            const mes = obtenerMesDeFecha(fecha);
            if (mes) mesesEncontrados.add(mes);
        }
    });

    cboMes.innerHTML = '<option value="TODOS">📅 Todos los meses</option>';

    MESES_VIVERO.forEach(mes => {
        if (mesesEncontrados.has(mes)) {
            const opt = document.createElement('option');
            opt.value = mes;
            opt.textContent = mes;
            cboMes.appendChild(opt);
        }
    });

    if (mesesEncontrados.has("Otros")) {
        const opt = document.createElement('option');
        opt.value = "Otros";
        opt.textContent = "Otros / Sin Formato";
        cboMes.appendChild(opt);
    }
}

function renderizarMarcadoresVivero() {
    capaVivero.clearLayers();

    const chk = document.getElementById('chk-vivero');
    const cboMes = document.getElementById('cbo-vivero-mes');
    const mesSeleccionado = cboMes ? cboMes.value : "TODOS";

    if (!chk || !chk.checked) {
        if (typeof map !== 'undefined') map.removeLayer(capaVivero);
        return;
    }

    // 1. AGRUPAR ACTIVIDADES POR COORDENADAS / LUGAR
    const gruposPorCoordenada = {};

    datosVivero.forEach((fila, idx) => {
        const fecha       = (fila["Fecha"] || "").trim();
        const area        = (fila["Área"] || fila["Area"] || "").trim();
        const subproceso  = (fila["Subproceso"] || "").trim();
        const etapa       = (fila["Etapa"] || "").trim();
        const descripcion = (fila["Descripción"] || fila["Descripcion"] || "").trim();
        const responsable = (fila["Responsables del reporte"] || "").trim();
        const obs         = (fila["Observaciones"] || "").trim();

        if (!fecha && !area && !subproceso && !descripcion) return;

        // Filtrado por mes
        if (mesSeleccionado !== "TODOS") {
            const mesFila = obtenerMesDeFecha(fecha);
            if (mesFila !== mesSeleccionado) return;
        }

        let lugarVal = (fila["Lugar"] || fila["lugar"] || "").trim();
        if (!lugarVal) lugarVal = "Vivero";

        const coords = typeof obtenerCoordenadas === 'function' 
            ? obtenerCoordenadas(lugarVal) 
            : (typeof diccionarioLugares !== 'undefined' && typeof normalizarTexto === 'function' 
                ? diccionarioLugares[normalizarTexto(lugarVal)] 
                : null);

        if (!coords) {
            console.warn(`⚠️ Vivero (Fila ${idx + 2}): El lugar "${lugarVal}" no se encontró.`);
            return;
        }

        const key = `${coords[0].toFixed(6)}_${coords[1].toFixed(6)}`;

        if (!gruposPorCoordenada[key]) {
            gruposPorCoordenada[key] = {
                lat: coords[0],
                lon: coords[1],
                lugar: lugarVal,
                actividades: []
            };
        }

        gruposPorCoordenada[key].actividades.push({
            fecha, area, subproceso, etapa, descripcion, responsable, obs
        });
    });

    // 2. CREAR UN MARCADOR POR CADA LUGAR CON SU POPUP MULTI-ACTIVIDAD
    let marcadoresCreados = 0;

    Object.values(gruposPorCoordenada).forEach(grupo => {
        const { lat, lon, lugar, actividades } = grupo;
        marcadoresCreados++;

        let popupHTML = "";

        if (actividades.length === 1) {
            // Un solo registro
            const act = actividades[0];
            popupHTML = `
                <div style="font-family: Arial, sans-serif; font-size:12px; color:#2d3748; width:270px;">
                    <div style="border-bottom:2px solid #38a169; padding-bottom:4px; margin-bottom:6px; font-weight:bold; color:#22543d; font-size:13px;">
                        🌱 Vivero: ${act.subproceso || act.area || 'Registro'}
                    </div>
                    <p style="margin:3px 0;"><b>📍 Lugar:</b> ${lugar}</p>
                    ${act.fecha ? `<p style="margin:3px 0;"><b>📅 Fecha:</b> ${act.fecha}</p>` : ''}
                    ${act.area ? `<p style="margin:3px 0;"><b>Área:</b> ${act.area}</p>` : ''}
                    ${act.etapa ? `<p style="margin:3px 0;"><b>Etapa:</b> ${act.etapa}</p>` : ''}
                    ${act.responsable ? `<p style="margin:3px 0;"><b>Responsable:</b> ${act.responsable}</p>` : ''}
                    ${act.descripcion ? `<p style="margin:3px 0;"><b>Descripción:</b> ${act.descripcion}</p>` : ''}
                    ${act.obs ? `<p style="margin:3px 0;"><b>Observaciones:</b> ${act.obs}</p>` : ''}
                </div>
            `;
        } else {
            // Múltiples registros: Lista desplegable con Scroll
            popupHTML = `
                <div style="font-family: Arial, sans-serif; font-size:12px; color:#2d3748; width:280px;">
                    <div style="border-bottom:2px solid #38a169; padding-bottom:6px; margin-bottom:8px; font-weight:bold; color:#22543d; font-size:13px; display:flex; justify-content:space-between; align-items:center;">
                        <span>📍 ${lugar}</span>
                        <span style="background:#28a745; color:white; border-radius:12px; padding:2px 8px; font-size:11px;">${actividades.length} actividades</span>
                    </div>
                    <div style="max-height: 260px; overflow-y: auto; padding-right: 4px;">
            `;

            actividades.forEach((act, i) => {
                popupHTML += `
                    <div style="background:#f7fafc; border:1px solid #e2e8f0; border-radius:6px; padding:8px; margin-bottom:8px;">
                        <div style="font-weight:bold; color:#22543d; margin-bottom:4px; font-size:12px; border-bottom: 1px dashed #cbd5e0; padding-bottom:2px;">
                            #${i + 1} - ${act.subproceso || act.area || 'Registro'}
                        </div>
                        ${act.fecha ? `<p style="margin:2px 0;"><b>📅 Fecha:</b> ${act.fecha}</p>` : ''}
                        ${act.area ? `<p style="margin:2px 0;"><b>Área:</b> ${act.area}</p>` : ''}
                        ${act.etapa ? `<p style="margin:2px 0;"><b>Etapa:</b> ${act.etapa}</p>` : ''}
                        ${act.responsable ? `<p style="margin:2px 0;"><b>Responsable:</b> ${act.responsable}</p>` : ''}
                        ${act.descripcion ? `<p style="margin:2px 0;"><b>Descripción:</b> ${act.descripcion}</p>` : ''}
                        ${act.obs ? `<p style="margin:2px 0;"><b>Observaciones:</b> ${act.obs}</p>` : ''}
                    </div>
                `;
            });

            popupHTML += `
                    </div>
                </div>
            `;
        }

        const marker = L.marker([lat, lon]).bindPopup(popupHTML);
        capaVivero.addLayer(marker);
    });

    if (typeof map !== 'undefined') {
        capaVivero.addTo(map);
    }

    console.log(`📍 Ubicaciones con actividades [${mesSeleccionado}]: ${marcadoresCreados}`);
}

// ==========================================================================
// LISTENERS PARA CHECKBOX Y SELECCIÓN DE MES
// ==========================================================================
document.addEventListener("DOMContentLoaded", function() {
    const chkVivero = document.getElementById('chk-vivero');
    const cboMes = document.getElementById('cbo-vivero-mes');

    if (chkVivero) {
        chkVivero.addEventListener('change', function() {
            if (this.checked) {
                if (datosVivero.length === 0) {
                    cargarVivero();
                } else {
                    renderizarMarcadoresVivero();
                }
            } else {
                if (typeof map !== 'undefined') map.removeLayer(capaVivero);
            }
        });

        if (chkVivero.checked) {
            cargarVivero();
        }
    }

    if (cboMes) {
        cboMes.addEventListener('change', function() {
            if (datosVivero.length === 0 && chkVivero && chkVivero.checked) {
                cargarVivero();
            } else {
                renderizarMarcadoresVivero();
            }
        });
    }
});

// ==========================================================================
// LÓGICA DE RESERVAS DE JARDINES (Google Sheets GViz API + GeoJSON)
// ==========================================================================

// ID extraído directamente de tu documento de reservas
const SHEET_ID_RESERVAS = '1R3Xz8A5xIVm-s0duMQhav2YAJrQdlSAYgKYeoEvoT8s';
const GID_RESERVAS = '1907703639';

var datosReservasGuardados = null;
var datosJardinesReservaGeoJSON = null;
var capaReservasJardines = L.layerGroup();

// 1. Normalizador de texto estricto
function normalizarTexto(texto) {
    if (!texto) return "";
    return texto.toString()
        .toLowerCase()
        .normalize("NFD")
        .replace(/[\u0300-\u036f]/g, "")    // Remueve tildes
        .replace(/[^a-z0-9]/g, "")          // Elimina espacios y símbolos
        .replace(/jardin|zona|sector/g, "") // Limpia términos genéricos
        .trim();
}

// 2. Extractor estricto de columnas
function obtenerValor(fila, campoBuscado) {
    if (!fila) return "";
    var busquedaNorm = normalizarTexto(campoBuscado);
    var keys = Object.keys(fila);

    var claveExacta = keys.find(k => normalizarTexto(k) === busquedaNorm);
    if (claveExacta && fila[claveExacta] !== undefined && fila[claveExacta] !== null) {
        return fila[claveExacta];
    }

    var claveInicio = keys.find(k => normalizarTexto(k).startsWith(busquedaNorm));
    if (claveInicio && fila[claveInicio] !== undefined && fila[claveInicio] !== null) {
        return fila[claveInicio];
    }

    return "";
}

// 3. Convertidor de mes desde el selector HTML ("Agosto" -> 7 en JS)
function obtenerMesJS() {
    var selectorMes = document.getElementById("reserva-mes");
    if (!selectorMes || selectorMes.selectedIndex === -1) return 7;
    
    var txt = selectorMes.options[selectorMes.selectedIndex].text.toLowerCase().trim();
    var val = selectorMes.value.toString().trim().toLowerCase();

    var listaMeses = ["enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "setiembre", "octubre", "noviembre", "diciembre"];
    
    var idxTexto = listaMeses.findIndex(m => txt.includes(m));
    if (idxTexto !== -1) return idxTexto;

    var idxVal = listaMeses.indexOf(val);
    if (idxVal !== -1) return idxVal;

    var num = parseInt(val, 10);
    if (!isNaN(num)) {
        var primerValor = parseInt(selectorMes.options[0].value, 10);
        return (primerValor === 1) ? num - 1 : num;
    }

    return 7;
}

// 4. Cálculo del rango semanal del calendario (Lunes a Domingo)
function obtenerRangoSemanaCalendario(numeroSemana) {
    var anioActual = 2026;
    var mesJS = obtenerMesJS(); 
    
    var diaUno = new Date(anioActual, mesJS, 1);
    var diaSemana = diaUno.getDay(); 
    var diasAtras = (diaSemana === 0) ? 6 : (diaSemana - 1);
    
    var lunesSemana1 = new Date(diaUno);
    lunesSemana1.setDate(diaUno.getDate() - diasAtras);
    
    var inicioSemana = new Date(lunesSemana1);
    inicioSemana.setDate(lunesSemana1.getDate() + (numeroSemana - 1) * 7);
    
    var finSemana = new Date(inicioSemana);
    finSemana.setDate(inicioSemana.getDate() + 6);
    
    inicioSemana.setHours(0, 0, 0, 0);
    finSemana.setHours(23, 59, 59, 999);
    
    return { inicio: inicioSemana, fin: finSemana };
}

// 5. Descarga de Reservas mediante CSV desde la Web Publicada
async function inicializarDescargaReservas() {
    // URL de la hoja publicada en formato CSV
    const url = 'https://docs.google.com/spreadsheets/d/e/2PACX-1vSqBX3fIW5O1ttB7LTCXVKez9PpxMqJS92BasZmiA9agTAVp0uLVrqejffIeHlf6z3J1ffQ2B_eDdpA/pub?gid=1907703639&single=true&output=csv';

    try {
        const response = await fetch(url);
        const csvText = await response.text();

        // Parsea el texto CSV a un array de líneas
        const lineas = csvText.split(/\r\n|\n/);
        if (lineas.length === 0) return;

        // Función para limpiar comillas de los valores CSV
        const limpiarCampo = campo => campo ? campo.replace(/^"(.*)"$/, '$1').trim() : '';

        // Extrae las cabeceras (primera fila)
        const cabeceras = lineas[0].split(',').map(limpiarCampo);

        // Mapea las filas restantes a objetos clave-valor
        datosReservasGuardados = lineas.slice(1).filter(linea => linea.trim() !== '').map(linea => {
            // Separa respetando las comillas que contienen comas
            const valores = linea.match(/(".*?"|[^",\s]+)(?=\s*,|\s*$)/g) || linea.split(',');
            let filaObj = {};

            cabeceras.forEach((col, idx) => {
                if (col) {
                    filaObj[col] = valores[idx] ? limpiarCampo(valores[idx]) : '';
                }
            });

            return filaObj;
        });

        console.log("📊 Total real de filas leídas desde Google Sheets (CSV):", datosReservasGuardados ? datosReservasGuardados.length : 0);
        if (datosReservasGuardados && datosReservasGuardados.length > 0) {
            console.log("🔍 Columnas detectadas:", Object.keys(datosReservasGuardados[0]));
        }

        procesarReservasSemanales();

    } catch (err) {
        console.error("❌ Error al cargar las reservas desde Google Sheets:", err);
    }
}


function coincidenJardines(nombreExcel, nombreGeoJSON) {
    var normExcel = normalizarTexto(nombreExcel);
    var normGeo = normalizarTexto(nombreGeoJSON);

    if (!normExcel || !normGeo) return false;
    return normExcel === normGeo;
}

function obtenerNombreDesdeShape(feature) {
    if (!feature || !feature.properties) return "";
    var props = feature.properties;
    var claveEncontrada = Object.keys(props).find(key => key.toLowerCase() === 'nombre' || key.toLowerCase() === 'jardin');
    return claveEncontrada ? props[claveEncontrada] : "";
}

// 6. Parseo universal de fechas
function parsearFechaUniversal(fechaRaw) {
    if (!fechaRaw) return null;
    if (fechaRaw instanceof Date) return isNaN(fechaRaw.getTime()) ? null : fechaRaw;

    var str = fechaRaw.toString().trim();
    if (!str) return null;

    if (str.includes('T')) {
        var partesIso = str.split('T')[0].split('-');
        if (partesIso.length === 3) {
            return new Date(parseInt(partesIso[0], 10), parseInt(partesIso[1], 10) - 1, parseInt(partesIso[2], 10));
        }
    }
    
    if (str.includes('/')) {
        var partes = str.split('/');
        if (partes.length === 3) {
            if (partes[0].length === 4) {
                return new Date(parseInt(partes[0], 10), parseInt(partes[1], 10) - 1, parseInt(partes[2], 10));
            } else {
                var d = parseInt(partes[0], 10);
                var m = parseInt(partes[1], 10) - 1;
                var a = parseInt(partes[2], 10);
                if (a < 100) a += 2000;
                return new Date(a, m, d);
            }
        }
    }
    
    if (str.includes('-')) {
        var partesDash = str.split('-');
        if (partesDash.length === 3) {
            if (partesDash[0].length === 4) {
                return new Date(parseInt(partesDash[0], 10), parseInt(partesDash[1], 10) - 1, parseInt(partesDash[2], 10));
            } else {
                return new Date(parseInt(partesDash[2], 10), parseInt(partesDash[1], 10) - 1, parseInt(partesDash[0], 10));
            }
        }
    }
    
    var dObj = new Date(str);
    return isNaN(dObj.getTime()) ? null : dObj;
}

// 7. Filtrado y renderizado en Leaflet
function procesarReservasSemanales() {
    var selectorSemana = document.getElementById("reserva-semana");
    var selectorDia = document.getElementById("reserva-dia");
    var chkReservas = document.getElementById("chk-reservas-jardines");
    
    if (!selectorSemana || !selectorDia) return;
    
    var semSel = parseInt(selectorSemana.value, 10);
    var filterDia = selectorDia.value;

    capaReservasJardines.clearLayers();

    if (chkReservas && !chkReservas.checked) return; 
    if (datosJardinesReservaGeoJSON === null) return;
    if (datosReservasGuardados === null) return; 

    var reservasFiltradas = [];
    var rangoSemana = obtenerRangoSemanaCalendario(semSel);

    datosReservasGuardados.forEach(function(fila) {
        var estadoVal = obtenerValor(fila, "estado");
        var fechaRaw = obtenerValor(fila, "fecha"); 

        if (estadoVal && !estadoVal.toString().toLowerCase().includes("reservad") && !estadoVal.toString().toLowerCase().includes("confirmad")) {
            return;
        }

        if (!fechaRaw) return;

        var fechaObjeto = parsearFechaUniversal(fechaRaw);
        if (!fechaObjeto) return;
        
        fechaObjeto.setHours(0, 0, 0, 0);

        var indiceDia = fechaObjeto.getDay(); 
        var dentroDeRango = (fechaObjeto >= rangoSemana.inicio && fechaObjeto <= rangoSemana.fin);

        if (dentroDeRango) {
            if (filterDia === "todos" || indiceDia === parseInt(filterDia, 10)) {
                reservasFiltradas.push(fila);
            }
        }
    });

    console.log("📅 Rango Evaluado Exacto:", rangoSemana.inicio.toLocaleDateString(), "al", rangoSemana.fin.toLocaleDateString());
    console.log("📍 Eventos encontrados para este rango:", reservasFiltradas.length);

    L.geoJSON(datosJardinesReservaGeoJSON, {
        style: function(feature) {
            var nombreJardinShape = obtenerNombreDesdeShape(feature);
            var tieneReserva = reservasFiltradas.some(function(reserva) {
                var jardinExcel = obtenerValor(reserva, "jardín") || obtenerValor(reserva, "jardin");
                return coincidenJardines(jardinExcel, nombreJardinShape);
            });

            return tieneReserva ? { color: "#d35400", fillColor: "#ff781f", fillOpacity: 0.75, weight: 3 } 
                                : { color: "#27ae60", fillColor: "#2ecc71", fillOpacity: 0.15, weight: 1.5, dashArray: "3" };
        },
        onEachFeature: function(feature, layer) {
            var nombreJardinShape = obtenerNombreDesdeShape(feature);
            
            var todasMisReservas = reservasFiltradas.filter(function(reserva) {
                var jardinExcel = obtenerValor(reserva, "jardín") || obtenerValor(reserva, "jardin");
                return coincidenJardines(jardinExcel, nombreJardinShape);
            });

            if (todasMisReservas.length > 0) {
                var listaEventosHTML = "";
                
                todasMisReservas.forEach(function(reserva, index) {
                    var fStr = obtenerValor(reserva, "fecha");
                    var horaStr = obtenerValor(reserva, "hora del evento") || obtenerValor(reserva, "hora");
                    var evStr = obtenerValor(reserva, "evento") || obtenerValor(reserva, "actividad");
                    var orgStr = obtenerValor(reserva, "unidad responsable") || obtenerValor(reserva, "organiza");

                    listaEventosHTML += `
                        <div style="margin-bottom: 8px; padding-bottom: 8px; ${index < todasMisReservas.length - 1 ? 'border-bottom: 1px dashed #cbd5e1;' : ''}">
                            <span style="font-size:10px; padding:2px 5px; background:#feebc8; color:#c05621; font-weight:bold; border-radius:3px; display:inline-block; margin-bottom:4px;">
                                📅 Evento #${index + 1} (${fStr})
                            </span><br>
                            <b>Lugar exacto:</b> ${nombreJardinShape}<br>
                            ${evStr ? `<b>Actividad:</b> ${evStr}<br>` : ''}
                            ${horaStr ? `<b>Horario:</b> ${horaStr}<br>` : ''}
                            ${orgStr ? `<b>Organiza:</b> ${orgStr}` : ''}
                        </div>
                    `;
                });

                layer.bindPopup(`
                    <div style="padding:6px; font-size:12px; max-width:270px; line-height:1.4;">
                        <h3 style="margin:0 0 8px 0; font-size:14px; color:#1a365d; border-bottom:2px solid #edf2f7; padding-bottom:4px;">
                            🌿 ${nombreJardinShape}
                        </h3>
                        <div style="max-height: 200px; overflow-y: auto;">
                            ${listaEventosHTML}
                        </div>
                    </div>
                `);
            } else {
                layer.bindPopup(`
                    <div style="font-size:12px; padding:4px; font-family: sans-serif;">
                        <b style="color:#27ae60; font-size:13px;">🌿 ${nombreJardinShape || "Jardín Activo"}</b><br>
                        <span style="color:gray; font-size:11px;">Disponible para la fecha seleccionada.</span>
                    </div>
                `);
            }
            capaReservasJardines.addLayer(layer);
        }
    });

    if (typeof map !== 'undefined' && !map.hasLayer(capaReservasJardines)) {
        capaReservasJardines.addTo(map);
    }
}

// 8. Carga inicial
function cargarJardinesReserva() {
    fetch('jardines_reserva.geojson')
        .then(function(response) { return response.json(); })
        .then(function(data) {
            datosJardinesReservaGeoJSON = data; 
            inicializarDescargaReservas(); 
        })
        .catch(function(error) { console.error("❌ Error leyendo jardines_reserva.geojson:", error); });
}

// Escuchadores de eventos para los filtros del panel
document.addEventListener("DOMContentLoaded", function() {
    var inputsFiltro = ['chk-reservas-jardines', 'reserva-mes', 'reserva-semana', 'reserva-dia'];
    inputsFiltro.forEach(function(id) {
        var el = document.getElementById(id);
        if (el) {
            el.addEventListener('change', procesarReservasSemanales);
        }
    });
    
    cargarJardinesReserva();
});


// ==========================================================================
// 8. GESTIÓN DE AGUA (CON MARKERCLUSTER Y APERTURA DE SIDEBAR LATERAL)
// ==========================================================================

document.addEventListener("DOMContentLoaded", function () {

    if (typeof map !== 'undefined' && !window.map) {
        window.map = map;
    }

    var URL_MI_API_DRIVE = "https://script.google.com/macros/s/AKfycbyul5ascYRZ1L5KpXECCnOzQC-SGS8iTHCBRkAgQfDvy4vAEJZiK0JI115PY4P2yyuQ/exec";
    var cacheImagenesDrive = {}; 

    // Verificación defensiva para asegurarnos que la librería MarkerCluster exista
    if (typeof L.markerClusterGroup !== 'function') {
        console.error("❌ Error: La librería Leaflet.markercluster no se ha cargado en el index.html.");
        return;
    }

    // 1. Configuración avanzada de MarkerCluster
    var clusterCentralAgua = L.markerClusterGroup({ 
        spiderfyOnMaxZoom: true,
        showCoverageOnHover: false,
        maxClusterRadius: 40,
        spiderfyDistanceMultiplier: 2.0,
        animate: true
    });

    if (window.map) {
        clusterCentralAgua.addTo(window.map);
    }

    // 2. Capas de almacenamiento temporal
    var capaBebedFuente = L.layerGroup();
    var capaBebedLlenador = L.layerGroup();
    var capaBebedNuevo = L.layerGroup();
    var capaBebedRenovacion = L.layerGroup(); // Unifica deterioro y baja

    function crearIconoBebedero(color, simbolo) {
        return L.divIcon({
            className: '', 
            html: `<div style="color:white; font-weight:bold; font-size:12px; text-align:center; border-radius:50%; border:2px solid white; box-shadow:0 2px 5px rgba(0,0,0,0.4); display:flex; align-items:center; justify-content:center; background-color: ${color}; width: 24px; height: 24px; margin-left: -12px; margin-top: -12px;">${simbolo}</div>`,
            iconSize: [24, 24],
            iconAnchor: [12, 12]
        });
    }

    // FUNCIÓN PARA ABRIR LA BARRA LATERAL (SIDEBAR)
    function abrirSidebarAgua(datos) {
        var sidebar = document.getElementById('sidebar-info');
        var contenido = document.getElementById('sidebar-contenido');

        if (!sidebar || !contenido) {
            console.error("❌ Error: No se encontró el elemento #sidebar-info.");
            return;
        }

        sidebar.style.zIndex = "10000";
        sidebar.style.position = "fixed";

        contenido.innerHTML = `
            <div style="padding-top: 15px;">
                <span style="font-size:11px; text-transform:uppercase; color:#2b6cb0; font-weight:bold; letter-spacing:0.5px;">UGA - Gestión de Agua</span>
                <h2 style="margin: 4px 0 12px 0; color:#1a365d; font-size:20px; font-weight: bold;">🚰 ${datos.tituloTipo}</h2>
                
                ${datos.urlFotoDirecta ? `
                    <div style="width:100%; border-radius:10px; overflow:hidden; margin-bottom:16px; border:1px solid #cbd5e1; box-shadow:0 3px 6px rgba(0,0,0,0.1);">
                        <a href="${datos.urlOriginalDrive}" target="_blank" title="Ver imagen completa en Google Drive">
                            <img src="${datos.urlFotoDirecta}" style="width:100%; height:200px; object-fit:cover; display:block;" alt="Foto Bebedero ${datos.nombrePunto}">
                        </a>
                    </div>
                ` : `
                    <div style="width:100%; height:120px; border:1px dashed #cbd5e1; border-radius:8px; display:flex; align-items:center; justify-content:center; background:#f7fafc; margin-bottom:16px;">
                        <p style="font-size:12px; color:#a0aec0; margin:0; text-align:center; font-style:italic;">Sin foto asociada (${datos.codigoBuscado})</p>
                    </div>
                `}

                <div style="background:#f8fafc; border:1px solid #e2e8f0; border-radius:8px; padding:12px; margin-bottom:14px;">
                    <span style="font-size:11px; text-transform:uppercase; color:#718096; font-weight:bold; display:block; margin-bottom:4px;">📍 Punto / Código</span>
                    <p style="margin:0; font-size:14px; color:#2d3748; font-weight:600;">${datos.nombrePunto}</p>
                </div>

                <div style="background:#f7fafc; padding:12px; border-radius:8px; border-left:4px solid #2b6cb0;">
                    <span style="font-size:11px; text-transform:uppercase; color:#718096; font-weight:bold; display:block; margin-bottom:4px;">ℹ️ Estado y Mantenimiento</span>
                    <span style="font-size:13px; color:#2d3748; font-weight: 500; line-height: 1.4; display: block;">${datos.recomendacion}</span>
                </div>
            </div>
        `;

        sidebar.classList.remove('sidebar-oculto');
        sidebar.classList.add('sidebar-visible');
    }

    function inicializarFotosDrive(callback) {
        fetch(URL_MI_API_DRIVE)
            .then(response => {
                if (!response.ok) throw new Error("Error en la respuesta de la Web App");
                return response.json();
            })
            .then(data => {
                cacheImagenesDrive = data;
                console.log("🚀 Fotos de Drive indexadas con éxito:", Object.keys(cacheImagenesDrive).length);
                if (callback) callback();
            })
            .catch(err => {
                console.error("❌ Error Web App Drive:", err);
                if (callback) callback();
            });
    }

    function cargarCapaBebedero(archivoNombre, capaDestino, color, simbolo, tituloTipo, recomendacion, idCheckbox) {
        fetch(archivoNombre)
            .then(response => {
                if (!response.ok) throw new Error("No se encontró: " + archivoNombre);
                return response.json();
            })
            .then(data => {
                var nuevosMarcadores = [];

                L.geoJSON(data, {
                    pointToLayer: function(feature, latlng) {
                        return L.marker(latlng, { icon: crearIconoBebedero(color, simbolo) });
                    },
                    onEachFeature: function(feature, layer) {
                        var nombrePunto = feature.properties.Name || "Bebedero Campus";
                        var codigoBuscado = nombrePunto.replace("PT_", "").toLowerCase().trim();
                        var idRealDrive = cacheImagenesDrive[codigoBuscado];
                        
                        var urlFotoDirecta = idRealDrive ? `https://lh3.googleusercontent.com/d/${idRealDrive}` : null;
                        var urlOriginalDrive = idRealDrive ? `https://drive.google.com/open?id=${idRealDrive}` : null;

                        var datosBebedero = {
                            nombrePunto: nombrePunto,
                            tituloTipo: tituloTipo,
                            recomendacion: recomendacion,
                            codigoBuscado: codigoBuscado,
                            urlFotoDirecta: urlFotoDirecta,
                            urlOriginalDrive: urlOriginalDrive
                        };

                        layer.on('click', function(e) {
                            if (e.originalEvent) {
                                L.DomEvent.stopPropagation(e.originalEvent);
                            }
                            abrirSidebarAgua(datosBebedero);
                        });
                        
                        capaDestino.addLayer(layer);
                        nuevosMarcadores.push(layer);
                    }
                });

                var cb = document.getElementById(idCheckbox);
                if (cb && cb.checked && nuevosMarcadores.length > 0) {
                    if (window.map && !window.map.hasLayer(clusterCentralAgua)) {
                        clusterCentralAgua.addTo(window.map);
                    }
                    clusterCentralAgua.addLayers(nuevosMarcadores);
                }

            }).catch(err => console.warn("Aviso en Agua:", err.message));
    }

    // Carga de archivos
    inicializarFotosDrive(function() {
        // OPERATIVOS
        cargarCapaBebedero('bebedero Tipo fuente.geojson', capaBebedFuente, '#3182ce', '⛲', 'Bebedero: Tipo Fuente', '✅ Estado: Operativo. Mantenimiento trimestral.', 'chk-bebed-fuente');
        cargarCapaBebedero('bebedero Tipo llenador de botella.geojson', capaBebedLlenador, '#00bfa5', '🧴', 'Bebedero: Llenador de Botella', '✅ Estado: Operativo. Filtro en óptimas condiciones.', 'chk-bebed-llenador');
        
        // PENDIENTES DE INSTALACIÓN
        cargarCapaBebedero('bebedero Nuevo.geojson', capaBebedNuevo, '#48bb78', '🟢', 'Bebedero: Proyecto Nuevo', '📌 Ubicación proyectada para instalación futura.', 'chk-bebed-nuevo');
        
        // RENOVACIÓN (Carga ambos GeoJSON dentro de la misma capa capaBebedRenovacion)
        cargarCapaBebedero('bebedero por deterioro.geojson', capaBebedRenovacion, '#ecc94b', '🔄', 'Bebedero: Renovación (Deterioro)', '🔄 Pendiente de reemplazo por deterioro.', 'chk-bebed-renovacion');
        cargarCapaBebedero('bebedero por baja del equipo.geojson', capaBebedRenovacion, '#e53e3e', '🔄', 'Bebedero: Renovación (Baja)', '🔄 Pendiente de reemplazo por baja del equipo.', 'chk-bebed-renovacion');
    });

   // FUNCIÓN DE ACTUALIZACIÓN DEL PANEL FLOTANTE DE TOTALES
    function actualizarPanelResumenAgua() {
        var chkFuente = document.getElementById('chk-bebed-fuente')?.checked;
        var chkLlenador = document.getElementById('chk-bebed-llenador')?.checked;
        var chkNuevo = document.getElementById('chk-bebed-nuevo')?.checked;
        var chkRenovacion = document.getElementById('chk-bebed-renovacion')?.checked;

        var cantFuente = chkFuente ? capaBebedFuente.getLayers().length : 0;
        var cantLlenador = chkLlenador ? capaBebedLlenador.getLayers().length : 0;
        var cantNuevo = chkNuevo ? capaBebedNuevo.getLayers().length : 0;
        var cantRenovacion = chkRenovacion ? capaBebedRenovacion.getLayers().length : 0;

        var total = cantFuente + cantLlenador + cantNuevo + cantRenovacion;

        // Actualizar textos HTML
        if (document.getElementById('cant-fuente')) document.getElementById('cant-fuente').textContent = cantFuente;
        if (document.getElementById('cant-llenador')) document.getElementById('cant-llenador').textContent = cantLlenador;
        if (document.getElementById('cant-nuevo')) document.getElementById('cant-nuevo').textContent = cantNuevo;
        if (document.getElementById('cant-renovacion')) document.getElementById('cant-renovacion').textContent = cantRenovacion;
        if (document.getElementById('cant-total-agua')) document.getElementById('cant-total-agua').textContent = total;

        // Mostrar u ocultar el panel flotante
        var panel = document.getElementById('panel-resumen-agua');
        if (panel) {
            if (chkFuente || chkLlenador || chkNuevo || chkRenovacion) {
                panel.classList.remove('panel-resumen-oculto');
                panel.classList.add('panel-resumen-visible');
            } else {
                panel.classList.remove('panel-resumen-visible');
                panel.classList.add('panel-resumen-oculto');
            }
        }
    }

    function reconstruirClusterAgua() {
        if (!window.map) return;
        
        clusterCentralAgua.clearLayers();
        
        var todosLosMarcadores = [];
        
        if (document.getElementById('chk-bebed-fuente')?.checked) capaBebedFuente.eachLayer(l => todosLosMarcadores.push(l));
        if (document.getElementById('chk-bebed-llenador')?.checked) capaBebedLlenador.eachLayer(l => todosLosMarcadores.push(l));
        if (document.getElementById('chk-bebed-nuevo')?.checked) capaBebedNuevo.eachLayer(l => todosLosMarcadores.push(l));
        if (document.getElementById('chk-bebed-renovacion')?.checked) capaBebedRenovacion.eachLayer(l => todosLosMarcadores.push(l));
        
        if (todosLosMarcadores.length > 0) {
            if (!window.map.hasLayer(clusterCentralAgua)) {
                clusterCentralAgua.addTo(window.map);
            }
            clusterCentralAgua.addLayers(todosLosMarcadores);
        }

        // Actualizar totales cada vez que se altere el cluster
        actualizarPanelResumenAgua();
    }

    function conectarCheckboxAgua(idCheckbox) {
        var cb = document.getElementById(idCheckbox);
        if (cb) {
            cb.addEventListener('change', function() {
                reconstruirClusterAgua();
            });
        }
    }

    conectarCheckboxAgua('chk-bebed-fuente');
    conectarCheckboxAgua('chk-bebed-llenador');
    conectarCheckboxAgua('chk-bebed-nuevo');
    conectarCheckboxAgua('chk-bebed-renovacion');

    var chkTodoAgua = document.getElementById('chk-todo-agua');
    if (chkTodoAgua) {
        chkTodoAgua.addEventListener('change', function() {
            var estado = this.checked;
            var aguaIds = ['chk-bebed-fuente', 'chk-bebed-llenador', 'chk-bebed-nuevo', 'chk-bebed-renovacion'];
            
            aguaIds.forEach(function(id) {
                var cb = document.getElementById(id);
                if (cb) cb.checked = estado;
            });
            
            reconstruirClusterAgua();
        });
    }

    // Botón Opcional para ocultar el panel manualmente si el usuario quiere
    var btnCerrarResumen = document.getElementById('btn-cerrar-resumen');
    if (btnCerrarResumen) {
        btnCerrarResumen.addEventListener('click', function() {
            var panel = document.getElementById('panel-resumen-agua');
            if (panel) {
                panel.classList.remove('panel-resumen-visible');
                panel.classList.add('panel-resumen-oculto');
            }
        });
    }

});


// ==========================================
// 8B. CONTROL DE CAPAS: CONEXIÓN INTERFAZ PROFESIONAL
// ==========================================
// Capas independientes para Gestión de Agua
var capaBebedFuente = L.layerGroup();
var capaBebedLlenador = L.layerGroup();
var capaBebedNuevo = L.layerGroup();
var capaBebedDeterioro = L.layerGroup();
var capaBebedBaja = L.layerGroup();

// ✨ Capa independiente para los cafetos
var capaCafetos = L.layerGroup(); 

// ✨ NUEVO: Capa independiente para los puntos del campus desde Google Sheets
var capaPuntosPUCP = L.layerGroup();

// --- MAPAS BASE (Satélite Google, Esri vs Clásico) ---
function cambiarMapaBase(idRadio, capaBase, nombreMapa) {
    var radio = document.getElementById(idRadio);
    if (!radio) return;
    
    radio.addEventListener('change', function() {
        if (this.checked) {
            console.log("🔄 [Mapa Base] Cambiando a: " + nombreMapa);
            
            // Removemos todas las capas base para evitar solapamientos invisibles en memoria
            if (map.hasLayer(mapaNormal)) map.removeLayer(mapaNormal);
            if (map.hasLayer(googleSatelite)) map.removeLayer(googleSatelite);
            if (map.hasLayer(esriSatelite)) map.removeLayer(esriSatelite);
            
            // Agregamos la capa base seleccionada
            map.addLayer(capaBase);
        }
    });
}

// Vinculación de los tres mapas base a tu interfaz
cambiarMapaBase('rb-satelite', googleSatelite, "Satélite Google 🛰️");
cambiarMapaBase('rb-esri', esriSatelite, "Satélite Esri (Alternativo) 🌍");
cambiarMapaBase('rb-clasico', mapaNormal, "Mapa Clásico 🗺️");


// --- CAPAS SUPERPUESTAS (Checkboxes individuales) ---
function vincularCapaPro(idCheckbox, capaLeaflet) {
    var checkbox = document.getElementById(idCheckbox);
    if (!checkbox) return;

    checkbox.addEventListener('change', function() {
        // CORRECCIÓN CLAVE: Si es un checkbox de agua, delegamos el renderizado al ClusterGroup
        if (idCheckbox.startsWith('chk-bebed-')) {
            if (typeof reconstruirClusterAgua === 'function') {
                reconstruirClusterAgua();
            }
            return;
        }

        if (this.checked) {
            map.addLayer(capaLeaflet);
            if (idCheckbox === 'chk-reservas-jardines' && typeof procesarReservasSemanales === 'function') {
                procesarReservasSemanales();
            }
        } else {
            map.removeLayer(capaLeaflet);
        }
    });
}

// Vinculación automática de capas individuales con tu interfaz
vincularCapaPro('chk-actividades', capaMarkers);
vincularCapaPro('chk-no-aprov', capaNoAprovechables);
vincularCapaPro('chk-papel', capaPapelCarton); 
vincularCapaPro('chk-plastico', capaPlastico);
vincularCapaPro('chk-vidrio', capaVidrio);
vincularCapaPro('chk-pilas', capaPilas);
vincularCapaPro('chk-peligrosos', capaPeligrosos);
vincularCapaPro('chk-raee', capaRAEE); 
vincularCapaPro('chk-metales', capaMetales);
vincularCapaPro('chk-aniquem', capaAniquem);

// 🟢 Reemplazo de Intermedios por Intermedios Plástico e Intermedios Metal
vincularCapaPro('chk-intermedios-plastico', capaIntermediosPlastico);
vincularCapaPro('chk-intermedios-metal', capaIntermediosMetal);

vincularCapaPro('chk-oscar', capasresponsable);
vincularCapaPro('chk-andres', capasresponsable);
vincularCapaPro('chk-alfonso', capasresponsable);
vincularCapaPro('chk-camposdepo', capasCamposDepo); 
vincularCapaPro('chk-bosque', capasBosqueHumedo);

vincularCapaPro('chk-ardillas', capaArdillas);
vincularCapaPro('chk-aves', capaAves);
vincularCapaPro('chk-canes', capaCanes);
vincularCapaPro('chk-insectos', capaInsectos);
vincularCapaPro('chk-gallinazos', capaGallinazos);
vincularCapaPro('chk-roedores', capaRoedores);
vincularCapaPro('chk-gatos', capaColoniasGatos); 

vincularCapaPro('chk-palmeras', capaPalmeras); 
vincularCapaPro('chk-cafetos', capaCafetos); 

// ✨ NUEVO: Vinculación del checkbox para los puntos del campus (puedes ajustar el ID en tu HTML)
vincularCapaPro('chk-fotos-pucp', capaPuntosPUCP);

// Registro de escuchas para bebederos de agua
vincularCapaPro('chk-bebed-fuente', capaBebedFuente);
vincularCapaPro('chk-bebed-llenador', capaBebedLlenador);
vincularCapaPro('chk-bebed-nuevo', capaBebedNuevo);
vincularCapaPro('chk-bebed-deterioro', capaBebedDeterioro);
vincularCapaPro('chk-bebed-baja', capaBebedBaja);

vincularCapaPro('chk-reservas-jardines', capaReservasJardines);


// --- LÓGICA DE BOTONES MAESTROS POR CATEGORÍA ---

// 1. Botón Maestro: GESTIÓN DE RESIDUOS (TACHOS)
var chkTodosTachos = document.getElementById('chk-todos-tachos');
if (chkTodosTachos) {
    chkTodosTachos.addEventListener('change', function() {
        var estado = this.checked;
        var tachoIds = [
            'chk-no-aprov', 'chk-papel', 'chk-plastico', 'chk-vidrio', 
            'chk-pilas', 'chk-peligrosos', 'chk-raee', 'chk-metales', 
            'chk-aniquem', 'chk-intermedios-plastico', 'chk-intermedios-metal'
        ];
        tachoIds.forEach(function(id) {
            var cb = document.getElementById(id);
            if (cb && cb.checked !== estado) {
                cb.checked = estado;
                cb.dispatchEvent(new Event('change')); 
            }
        });
    });
}

// 2. Botón Maestro: ZONAS DE ÁREAS VERDES
var chkTodasVerdes = document.getElementById('chk-todas-verdes');
if (chkTodasVerdes) {
    chkTodasVerdes.addEventListener('change', function() {
        var estado = this.checked;
        var verdesIds = [
            'chk-oscar', 'chk-andres', 'chk-alfonso', 'chk-camposdepo', 'chk-bosque'
        ];
        verdesIds.forEach(function(id) {
            var cb = document.getElementById(id);
            if (cb && cb.checked !== estado) {
                cb.checked = estado;
                cb.dispatchEvent(new Event('change')); 
            }
        });
    });
}

// 3. Botón Maestro: PLAN DE FAUNA (Incluye a los gatos)
var chkTodaFauna = document.getElementById('chk-toda-fauna');
if (chkTodaFauna) {
    chkTodaFauna.addEventListener('change', function() {
        var estado = this.checked;
        var faunaIds = [
            'chk-ardillas', 'chk-aves', 'chk-canes', 'chk-insectos', 'chk-gallinazos', 'chk-roedores', 'chk-gatos'
        ];
        faunaIds.forEach(function(id) {
            var cb = document.getElementById(id);
            if (cb && cb.checked !== estado) {
                cb.checked = estado;
                cb.dispatchEvent(new Event('change')); 
            }
        });
    });
}

// 4. Botón Maestro: GESTIÓN DE FLORA (Incluye Palmeras y Cafetos)
/*var chkTodaFlora = document.getElementById('chk-toda-flora');
if (chkTodaFlora) {
    chkTodaFlora.addEventListener('change', function() {
        var estado = this.checked;
        var floraIds = [
            'chk-palmeras', 'chk-cafetos'
        ];
        floraIds.forEach(function(id) {
            var cb = document.getElementById(id);
            if (cb && cb.checked !== estado) {
                cb.checked = estado;
                cb.dispatchEvent(new Event('change')); 
            }
        });
    });
}*/
// ==========================================
// 4.2. Botón Maestro: GESTIÓN DE FLORA (MAPA INTERACTIVO DE DATOS)
// ==========================================

/* 
[NOTA]: El botón "chk-toda-flora" activa y desactiva este nuevo mapa. 
Los datos se obtienen directamente desde el Google Sheet.
*/
/*
// --- Función para convertir enlaces de Google Drive en imágenes directas válidas ---
function transformarEnlaceDrive(urlDrive) {
    if (!urlDrive) return '';
    const match = urlDrive.match(/\/d\/([a-zA-Z0-9_-]+)/) || urlDrive.match(/id=([a-zA-Z0-9_-]+)/);
    if (match && match[1]) {
        return `https://lh3.googleusercontent.com/d/${match[1]}`;
    }
    return urlDrive;
}

// --- Obtener datos de Flora desde el Google Sheet ---
async function cargarNuevosDatosFlora() {
    // ID de tu hoja de cálculo
    const SHEET_ID = '1QQxHKnefZ94Nk2raTwOe0N_zxBlb3AdQC_MP3ZzfRoI';
    
    // URL para obtener JSON de la API de Visualización de Google
    const url = `https://docs.google.com/spreadsheets/d/${SHEET_ID}/gviz/tq?tqx=out:json`;

    try {
        const response = await fetch(url);
        const text = await response.text();
        
        // Limpiamos la respuesta de Google para parsear solo el JSON
        const json = JSON.parse(text.substring(47).slice(0, -2));
        
        const data = [];
        const cols = json.table.cols;
        const rows = json.table.rows;

        if (!cols || cols.length === 0) {
            console.error("No se pudieron leer las columnas del Excel.");
            return [];
        }

        // Búsqueda flexible de índices de columnas
        const getColIndex = (nombreBuscado) => {
            return cols.findIndex(c => c && c.label && c.label.trim().toLowerCase() === nombreBuscado.toLowerCase());
        };

        const idxLat = getColIndex('Latitud');
        const idxLng = getColIndex('Longitud');
        const idxNom = getColIndex('Nombre común');
        const idxCien = getColIndex('Nombre científico');
        const idxVeg = getColIndex('Tipo de vegetación');
        const idxFoto = getColIndex('Foto');

        if (idxLat === -1 || idxLng === -1) {
            console.warn("⚠️ No se encontraron las columnas 'Latitud' o 'Longitud' en el Excel.");
            return [];
        }

        // Función para arreglar coordenadas con comas decimales (ej: -12,067 -> -12.067)
        const parseCoordenada = (celda) => {
            if (!celda || celda.v === null || celda.v === undefined) return null;
            const strVal = String(celda.v).replace(',', '.').trim();
            const num = parseFloat(strVal);
            return isNaN(num) ? null : num;
        };

        rows.forEach(row => {
            if (!row.c) return;

            const lat = parseCoordenada(row.c[idxLat]);
            const lng = parseCoordenada(row.c[idxLng]);

            // Solo agregamos puntos con coordenadas válidas
            if (lat !== null && lng !== null) {
                const getVal = (idx) => (idx !== -1 && row.c[idx] && row.c[idx].v !== null) ? row.c[idx].v : '';

                data.push({
                    lat: lat,
                    lng: lng,
                    nombre: getVal(idxNom) || 'Sin nombre común',
                    cientifico: getVal(idxCien) || 'Sin nombre científico',
                    vegetacion: getVal(idxVeg) || 'Sin tipo',
                    foto: getVal(idxFoto) ? transformarEnlaceDrive(getVal(idxFoto)) : ''
                });
            }
        });

        console.log(`🌳 Carga exitosa: ${data.length} puntos de flora obtenidos.`);
        return data;
    } catch (error) {
        console.error("Error al cargar datos de flora:", error);
        return [];
    }
}

// --- Variable para guardar la capa de Flora ---
var capaNuevaFlora = L.layerGroup();
if (typeof map !== 'undefined') {
    capaNuevaFlora.addTo(map);
}

// --- Función para dibujar el mapa de Flora ---
function iniciarMapaNuevaFlora() {
    capaNuevaFlora.clearLayers(); // Limpiamos marcas anteriores

    cargarNuevosDatosFlora().then(data => {
        if (data.length === 0) {
            console.warn("No se encontraron datos de flora válidos para desplegar.");
            return;
        }

        data.forEach(item => {
            // Icono personalizado con emoji de árbol
            var iconoFlora = L.divIcon({
                className: 'flora-marker-custom', 
                html: `<div style="font-size:26px; cursor:pointer; line-height:1; filter: drop-shadow(0px 2px 2px rgba(0,0,0,0.3));">🌳</div>`,
                iconSize: [30, 30],
                iconAnchor: [15, 25]
            });

            const marker = L.marker([item.lat, item.lng], { icon: iconoFlora });
            
            // HTML de la foto (se ajusta proporcionalmente sin deforma o cortar la imagen)
            let fotoHtml = '';
            if (item.foto && item.foto.startsWith('http')) {
                fotoHtml = `
                    <div style="text-align: center; margin-top: 8px; background-color: #f8fafc; border-radius: 8px; padding: 4px; border: 1px solid #e2e8f0; overflow: hidden;">
                        <a href="${item.foto}" target="_blank">
                            <img src="${item.foto}" 
                                 style="width: 100%; max-height: 180px; object-fit: contain; border-radius: 6px; display: block; margin: 0 auto;" 
                                 alt="${item.nombre}">
                        </a>
                    </div>`;
            }

            // Diseño final del Popup
            const popupContent = `
                <div style="font-family: Arial, sans-serif; text-align: left; font-size: 13px; width: 250px; padding: 2px;">
                    <h3 style="margin: 0 0 4px 0; color: #1e3a8a; font-size: 15px; font-weight: bold; border-bottom: 2px solid #22c55e; padding-bottom: 4px;">
                        ${item.nombre}
                    </h3>
                    <p style="margin: 4px 0; font-style: italic; color: #475569; font-size: 12px;">
                        ${item.cientifico}
                    </p>
                    <div style="margin: 6px 0;">
                        <span style="background: #f1f5f9; color: #334155; padding: 3px 8px; border-radius: 12px; font-size: 11px; font-weight: 600; border: 1px solid #cbd5e1;">
                            Tipo: ${item.vegetacion}
                        </span>
                    </div>
                    ${fotoHtml}
                </div>
            `;

            marker.bindPopup(popupContent, { maxWidth: 280 });
            capaNuevaFlora.addLayer(marker);
        });

        console.log("🌳 Mapa de Flora renderizado correctamente.");
    });
}

// --- Control del Botón Maestro de Flora ---
var chkTodaFlora = document.getElementById('chk-toda-flora');
if (chkTodaFlora) {
    // Clonamos el botón para eliminar eventos antiguos y evitar ejecuciones dobles
    var nuevoEvento = chkTodaFlora.cloneNode(true);
    chkTodaFlora.parentNode.replaceChild(nuevoEvento, chkTodaFlora);
    chkTodaFlora = nuevoEvento;

    chkTodaFlora.addEventListener('change', function() {
        var estado = this.checked;
        if (estado) {
            if (!map.hasLayer(capaNuevaFlora)) {
                capaNuevaFlora.addTo(map);
            }
            iniciarMapaNuevaFlora();
        } else {
            if (map.hasLayer(capaNuevaFlora)) {
                map.removeLayer(capaNuevaFlora);
            }
            capaNuevaFlora.clearLayers();
        }
    });
}*/


// ==========================================
// 4.3 Botón Maestro: GESTIÓN DE FLORA (SIMBOLOGÍA EXACTA DEL EXCEL, FILTROS Y PANEL LATERAL)
// ==========================================

// Variables globales para la edición, el panel y los filtros
var capaNuevaFlora = L.layerGroup();
var ultimaFilaEditada = null;
var panelConteoFloraControl = null;
var datosFloraOriginales = [];

// --- Simbología exacta según el Tipo de Vegetación del Excel ---
function obtenerSimboloVegetacion(tipo) {
    const val = (tipo || '').trim().toLowerCase();

    if (val.includes('palmera')) return '🌴';

    // SVG Nube verde para Arbustos (Agrandado a 28x28)
    if (val.includes('arbusto')) {
        return '<svg width="28" height="28" viewBox="0 0 24 24" fill="#22c55e" style="display: inline-block; vertical-align: middle;" xmlns="http://www.w3.org/2000/svg">' +
               '<path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96z"/>' +
               '</svg>';
    }

    if (val.includes('suculenta')) return '🌵';
    if (val.includes('trepadora')) return '🌺';
    if (val.includes('herbácea') || val.includes('herbacea')) return '🪴'; // <-- Emoji de maceta asignado
    if (val.includes('árbol') || val.includes('arbol')) return '🌳';

    return '🌳'; // Ícono por defecto
}

// --- Función para convertir enlaces de Google Drive en imágenes visibles ---
function transformarEnlaceDrive(urlDrive) {
    if (!urlDrive) return '';
    const match = urlDrive.match(/\/d\/([a-zA-Z0-9_-]+)/) || urlDrive.match(/[?&]id=([a-zA-Z0-9_-]+)/);
    if (match && match[1]) {
        return `https://lh3.googleusercontent.com/d/${match[1]}=s800`;
    }
    return urlDrive;
}

// --- Panel Lateral Derecho ---
function mostrarPanelLateralFlora(item, latActual, lngActual) {
    let panel = document.getElementById('panel-lateral-flora');
    if (!panel) {
        panel = document.createElement('div');
        panel.id = 'panel-lateral-flora';
        panel.style.cssText = `
            position: fixed; top: 15px; right: 15px; width: 310px; max-height: calc(100vh - 40px);
            background: white; z-index: 99999; border-radius: 10px; box-shadow: 0 4px 25px rgba(0,0,0,0.35);
            padding: 16px; font-family: Arial, sans-serif; overflow-y: auto; border-top: 5px solid #1e3a8a;
        `;
        document.body.appendChild(panel);
    }

    const icono = obtenerSimboloVegetacion(item.tipoVegetacion);

    let fotoHtml = item.foto 
        ? `<div style="text-align: center; margin: 10px 0; background: #f8fafc; border-radius: 8px; padding: 4px; border: 1px solid #e2e8f0;">
             <a href="${item.foto}" target="_blank">
                <img src="${item.foto}" style="width: 100%; max-height: 180px; object-fit: contain; border-radius: 6px;" alt="${item.nombre}">
             </a>
           </div>`
        : '<p style="font-size: 11px; color: #94a3b8; font-style: italic;">Sin foto disponible</p>';

    let codigoTexto = (!item.codigo || item.codigo.toLowerCase() === 'null' || item.codigo === '0') ? 'No tiene' : item.codigo;

    let botonGuardarHtml = (typeof modoEdicionFlora !== 'undefined' && modoEdicionFlora) ? `
        <button onclick="guardarCoordenadasEnSheets('${item.nro}', ${latActual}, ${lngActual})" 
            style="background-color: #28a745; color: white; border: none; padding: 10px 12px; border-radius: 4px; cursor: pointer; font-weight: bold; width: 100%; margin-top: 10px; font-size: 12px;">
            💾 Guardar en Google Sheets
        </button>
    ` : '';

    panel.innerHTML = `
        <div style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #e2e8f0; padding-bottom: 8px;">
            <strong style="color: #1e3a8a; font-size: 14px;">N° ${item.nro || 'S/N'} - ${item.referencia || 'S/Ref'}</strong>
            <button onclick="cerrarPanelLateralFlora()" style="border: none; background: transparent; font-size: 16px; cursor: pointer;">✖</button>
        </div>
        <div style="margin-top: 10px; font-size: 12px; color: #334155; line-height: 1.4;">
            <h3 style="margin: 0 0 4px 0; color: #1e3a8a; font-size: 15px; font-weight: bold;">${item.nombre}</h3>
            <p style="margin: 0 0 4px 0; font-style: italic; color: #475569;">${item.cientifico}</p>
            <p style="margin: 4px 0; font-weight: bold; color: #15803d; display: flex; align-items: center; gap: 4px;">${icono} ${item.tipoVegetacion}</p>
            <p style="margin: 2px 0; color: #64748b;">📍 Ubicación: ${item.ubicacion}</p>
            <p style="margin: 4px 0;"><strong>Código:</strong> ${codigoTexto}</p>
            <p style="margin: 4px 0;"><strong>Lat:</strong> ${latActual.toFixed(6)}</p>
            <p style="margin: 4px 0;"><strong>Lng:</strong> ${lngActual.toFixed(6)}</p>
            ${fotoHtml}
            ${botonGuardarHtml}
        </div>
    `;
    panel.style.display = 'block';
}

function cerrarPanelLateralFlora() {
    const panel = document.getElementById('panel-lateral-flora');
    if (panel) panel.style.display = 'none';
}

// --- Actualización limpia de los números del contador ---
function actualizarMetricasPanelFlora(data) {
    if (!data) return;

    const totalRegistros = data.length;
    const especiesUnicas = new Set(
        data.map(item => item.cientifico && item.cientifico !== 'Sin nombre científico' 
            ? item.cientifico.toLowerCase().trim() 
            : item.nombre.toLowerCase().trim())
    ).size;

    const totalPalmeras = data.filter(item => {
        const text = (item.tipoVegetacion + ' ' + item.nombre + ' ' + item.cientifico).toLowerCase();
        return text.includes('palmera');
    }).length;

    const totalArboles = data.filter(item => {
        const text = (item.tipoVegetacion + ' ' + item.nombre + ' ' + item.cientifico).toLowerCase();
        return text.includes('árbol') || text.includes('arbol');
    }).length;

    const elEspecies = document.getElementById('stat-flora-especies');
    const elArboles = document.getElementById('stat-flora-arboles');
    const elPalmeras = document.getElementById('stat-flora-palmeras');
    const elTotal = document.getElementById('stat-flora-total');

    if (elEspecies) elEspecies.textContent = especiesUnicas;
    if (elArboles) elArboles.textContent = totalArboles;
    if (elPalmeras) elPalmeras.textContent = totalPalmeras;
    if (elTotal) elTotal.textContent = totalRegistros;
}

// --- Creación del Panel de Conteo y Filtros ---
function crearPanelControlFlora() {
    if (panelConteoFloraControl) {
        map.removeControl(panelConteoFloraControl);
        panelConteoFloraControl = null;
    }

    const ubicaciones = [...new Set(datosFloraOriginales.map(d => d.ubicacion))].filter(Boolean).sort();
    const nombresComunes = [...new Set(datosFloraOriginales.map(d => d.nombre))].filter(Boolean).sort();
    const tiposVegetacion = [...new Set(datosFloraOriginales.map(d => d.tipoVegetacion))].filter(Boolean).sort();

    var PanelConteo = L.Control.extend({
        options: { position: 'bottomleft' },
        onAdd: function(map) {
            var div = L.DomUtil.create('div', '');
            div.innerHTML = `
                <div style="background: rgba(255, 255, 255, 0.95); padding: 12px 16px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.15); margin: 10px; min-width: 250px; font-size: 11px; border-left: 4px solid #1e3a8a; font-family: Arial, sans-serif;">
                    <div style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #e2e8f0; padding-bottom: 4px; margin-bottom: 8px;">
                        <h4 style="margin: 0; color: #1e3a8a; font-size: 12px; text-transform: uppercase; font-weight: bold;">
                            📊 Conteo y Filtros
                        </h4>
                        <button onclick="limpiarFiltrosFlora()" style="background: #e2e8f0; border: none; border-radius: 4px; padding: 2px 6px; font-size: 10px; cursor: pointer; color: #334155; font-weight: bold;">
                            🔄 Limpiar
                        </button>
                    </div>
                    
                    <div style="margin-bottom: 8px; display: flex; flex-direction: column; gap: 4px;">
                        <select id="filtro-ubicacion" onchange="ejecutarFiltroFlora()" style="width: 100%; font-size: 11px; padding: 3px 4px; border-radius: 4px; border: 1px solid #cbd5e1;">
                            <option value="">-- Todas las Ubicaciones --</option>
                            ${ubicaciones.map(u => `<option value="${u}">${u}</option>`).join('')}
                        </select>
                        <select id="filtro-nombre-comun" onchange="ejecutarFiltroFlora()" style="width: 100%; font-size: 11px; padding: 3px 4px; border-radius: 4px; border: 1px solid #cbd5e1;">
                            <option value="">-- Todos los Nombres Comunes --</option>
                            ${nombresComunes.map(n => `<option value="${n}">${n}</option>`).join('')}
                        </select>
                        <select id="filtro-tipo-veg" onchange="ejecutarFiltroFlora()" style="width: 100%; font-size: 11px; padding: 3px 4px; border-radius: 4px; border: 1px solid #cbd5e1;">
                            <option value="">-- Tipos de Vegetación --</option>
                            ${tiposVegetacion.map(t => `<option value="${t}">${t}</option>`).join('')}
                        </select>
                    </div>

                    <div style="display: flex; justify-content: space-between; margin-bottom: 3px; color: #334155;">
                        <span>Especies Únicas:</span>
                        <strong id="stat-flora-especies" style="color: #0f172a;">0</strong>
                    </div>
                    <div style="display: flex; justify-content: space-between; margin-bottom: 3px; color: #334155;">
                        <span>Total Árboles:</span>
                        <strong id="stat-flora-arboles" style="color: #15803d;">0</strong>
                    </div>
                    <div style="display: flex; justify-content: space-between; margin-bottom: 3px; color: #334155;">
                        <span>Total Palmeras:</span>
                        <strong id="stat-flora-palmeras" style="color: #b45309;">0</strong>
                    </div>
                    <div style="display: flex; justify-content: space-between; margin-top: 6px; padding-top: 4px; border-top: 1px solid #cbd5e1; font-weight: bold; color: #1e3a8a;">
                        <span>Individuos Visibles:</span>
                        <span id="stat-flora-total">0</span>
                    </div>
                </div>
            `;
            L.DomEvent.disableClickPropagation(div);
            L.DomEvent.disableScrollPropagation(div);
            return div;
        }
    });

    panelConteoFloraControl = new PanelConteo();
    map.addControl(panelConteoFloraControl);
}

function removerPanelConteoFlora() {
    if (panelConteoFloraControl) {
        map.removeControl(panelConteoFloraControl);
        panelConteoFloraControl = null;
    }
}

function limpiarFiltrosFlora() {
    const selUbi = document.getElementById('filtro-ubicacion');
    const selNom = document.getElementById('filtro-nombre-comun');
    const selTipo = document.getElementById('filtro-tipo-veg');

    if (selUbi) selUbi.value = "";
    if (selNom) selNom.value = "";
    if (selTipo) selTipo.value = "";

    ejecutarFiltroFlora();
}

// --- Carga de datos desde Google Sheets ---
async function cargarNuevosDatosFlora() {
    const SHEET_ID = '1QQxHKnefZ94Nk2raTwOe0N_zxBlb3AdQC_MP3ZzfRoI';
    const url = `https://docs.google.com/spreadsheets/d/${SHEET_ID}/gviz/tq?tqx=out:json&gid=1270218104`;

    try {
        const response = await fetch(url);
        const text = await response.text();
        
        const json = JSON.parse(text.substring(47).slice(0, -2));
        const data = [];
        const rows = json.table.rows;

        const idxNro = 0;        // Col A
        const idxUbicacion = 1;  // Col B
        const idxReferencia = 2; // Col C
        const idxLat = 3;        // Col D
        const idxLng = 4;        // Col E
        const idxNom = 5;        // Col F
        const idxCien = 6;       // Col G
        const idxTipoVeg = 7;    // Col H
        const idxFoto = 9;       // Col J
        const idxCodigo = 10;    // Col K

        const parseCoordenada = (celda) => {
            if (!celda || celda.v === null || celda.v === undefined) return null;
            const strVal = String(celda.v).replace(',', '.').trim();
            const num = parseFloat(strVal);
            return isNaN(num) ? null : num;
        };

        const getVal = (idx, row) => {
            if (row.c && row.c[idx] && row.c[idx].v !== null && row.c[idx].v !== undefined) {
                return String(row.c[idx].v).trim();
            }
            return '';
        };

        for (let i = 0; i < rows.length; i++) {
            const row = rows[i];
            if (!row.c) continue;

            const lat = parseCoordenada(row.c[idxLat]);
            const lng = parseCoordenada(row.c[idxLng]);

            if (lat !== null && lng !== null) {
                data.push({
                    nro: getVal(idxNro, row),
                    ubicacion: getVal(idxUbicacion, row) || 'Sin ubicación',
                    referencia: getVal(idxReferencia, row),
                    lat: lat,
                    lng: lng,
                    nombre: getVal(idxNom, row) || 'Sin nombre común',
                    cientifico: getVal(idxCien, row) || 'Sin nombre científico',
                    tipoVegetacion: getVal(idxTipoVeg, row) || 'No especificado',
                    foto: getVal(idxFoto, row) ? transformarEnlaceDrive(getVal(idxFoto, row)) : '',
                    codigo: getVal(idxCodigo, row) || 'null'
                });
            }
        }

        datosFloraOriginales = data;
        console.log(`🌳 Carga exitosa: ${data.length} puntos de flora obtenidos.`);
        return data;
    } catch (error) {
        console.error("Error al cargar datos de flora:", error);
        return [];
    }
}

// --- Renderizar marcadores dinámicos ---
function renderizarPuntosFlora(data) {
    capaNuevaFlora.clearLayers();
    actualizarMetricasPanelFlora(data);

    data.forEach(item => {
        const contenidoSimbolo = obtenerSimboloVegetacion(item.tipoVegetacion);

        var iconoFlora = L.divIcon({
            className: 'flora-marker-custom', 
            html: `<div style="cursor:pointer; filter: drop-shadow(0px 1px 1px rgba(0,0,0,0.4)); display: flex; align-items: center; justify-content: center; font-size: 28px;">${contenidoSimbolo}</div>`,
            iconSize: [30, 30],
            iconAnchor: [15, 15]
        });

        const marker = L.marker([item.lat, item.lng], { 
            icon: iconoFlora,
            draggable: (typeof modoEdicionFlora !== 'undefined' ? modoEdicionFlora : false)
        });

        marker.nro = item.nro;
        marker.datosOriginales = item;

        // --- VINCULACIÓN DE ETIQUETA PERMANENTE (REFERENCIA) ---
        if (item.referencia) {
            marker.bindTooltip(item.referencia, {
                permanent: true,                      // Siempre visible
                direction: 'top',                     // Por encima del icono
                offset: [0, -12],                     // Ajuste visual según tamaño del marcador
                className: 'etiqueta-solo-referencia', // Estilo CSS
                interactive: false                    // No bloquea clics
            });
        }

        marker.on('click', function() {
            ultimaFilaEditada = item;
            var pos = marker.getLatLng();
            mostrarPanelLateralFlora(item, pos.lat, pos.lng);
        });

        marker.on('dragend', function(e) {
            var pos = e.target.getLatLng();
            mostrarPanelLateralFlora(item, pos.lat, pos.lng);
        });

        capaNuevaFlora.addLayer(marker);
    });
}

// --- Ejecutar filtrado ---
function ejecutarFiltroFlora() {
    const valUbi = document.getElementById('filtro-ubicacion')?.value || '';
    const valNom = document.getElementById('filtro-nombre-comun')?.value || '';
    const valTipo = document.getElementById('filtro-tipo-veg')?.value || '';

    const datosFiltrados = datosFloraOriginales.filter(item => {
        const cumpleUbi = !valUbi || item.ubicacion === valUbi;
        const cumpleNom = !valNom || item.nombre === valNom;
        const cumpleTipo = !valTipo || item.tipoVegetacion === valTipo;
        return cumpleUbi && cumpleNom && cumpleTipo;
    });

    renderizarPuntosFlora(datosFiltrados);
}


// --- Inicialización del mapa ---
function iniciarMapaNuevaFlora() {
    cargarNuevosDatosFlora().then(data => {
        if (data.length === 0) {
            console.warn("No se encontraron datos de flora válidos para desplegar.");
            removerPanelConteoFlora();
            return;
        }
        crearPanelControlFlora();
        renderizarPuntosFlora(data);
        console.log("🌳 Mapa de Flora renderizado correctamente.");
    });
}

// --- Control del Botón Maestro de Flora ---
var chkTodaFlora = document.getElementById('chk-toda-flora');
if (chkTodaFlora) {
    var nuevoEvento = chkTodaFlora.cloneNode(true);
    chkTodaFlora.parentNode.replaceChild(nuevoEvento, chkTodaFlora);
    chkTodaFlora = nuevoEvento;

    chkTodaFlora.addEventListener('change', function() {
        var estado = this.checked;
        if (estado) {
            if (!map.hasLayer(capaNuevaFlora)) {
                capaNuevaFlora.addTo(map);
            }
            iniciarMapaNuevaFlora();
        } else {
            if (map.hasLayer(capaNuevaFlora)) {
                map.removeLayer(capaNuevaFlora);
            }
            capaNuevaFlora.clearLayers();
            removerPanelConteoFlora();
            cerrarPanelLateralFlora();
        }
    });
}


// ==========================================
// SECCIÓN: PANEL DE CONTEO Y FILTROS DINÁMICOS
// ==========================================

// --- Actualización de métricas dinámicas por Tipo de Vegetación ---
function actualizarMetricasPanelFlora(data) {
    if (!data) return;

    const totalRegistros = data.length;

    // Conteo de Especies Únicas
    const especiesUnicas = new Set(
        data.map(item => item.cientifico && item.cientifico !== 'Sin nombre científico' 
            ? item.cientifico.toLowerCase().trim() 
            : item.nombre.toLowerCase().trim())
    ).size;

    // Conteo agrupado dinámicamente según lo que venga en la columna "tipoVegetacion"
    const conteoPorTipo = {};
    data.forEach(item => {
        const tipo = (item.tipoVegetacion && item.tipoVegetacion !== '') ? item.tipoVegetacion.trim() : 'Sin especificar';
        conteoPorTipo[tipo] = (conteoPorTipo[tipo] || 0) + 1;
    });

    // Generación del HTML para cada tipo encontrado (Arbusto, Palmera, Herbácea, Suculenta, etc.)
    const htmlTipos = Object.entries(conteoPorTipo)
        .sort((a, b) => b[1] - a[1]) // Ordenados de mayor a menor
        .map(([tipo, cantidad]) => `
            <div style="display: flex; justify-content: space-between; margin-bottom: 3px; color: #334155;">
                <span>${tipo}:</span>
                <strong style="color: #15803d;">${cantidad}</strong>
            </div>
        `).join('');

    // Inyección de valores en el panel Leaflet
    const elEspecies = document.getElementById('stat-flora-especies');
    const contenedorTipos = document.getElementById('contenedor-metricas-flora');
    const elTotal = document.getElementById('stat-flora-total');

    if (elEspecies) elEspecies.textContent = especiesUnicas;
    if (contenedorTipos) contenedorTipos.innerHTML = htmlTipos;
    if (elTotal) elTotal.textContent = totalRegistros;
}

// --- Creación y renderizado del Panel Leaflet ---
function crearPanelControlFlora() {
    if (panelConteoFloraControl) {
        map.removeControl(panelConteoFloraControl);
        panelConteoFloraControl = null;
    }

    const ubicaciones = [...new Set(datosFloraOriginales.map(d => d.ubicacion))].filter(Boolean).sort();
    const nombresComunes = [...new Set(datosFloraOriginales.map(d => d.nombre))].filter(Boolean).sort();
    const tiposVegetacion = [...new Set(datosFloraOriginales.map(d => d.tipoVegetacion))].filter(Boolean).sort();

    var PanelConteo = L.Control.extend({
        options: { position: 'bottomleft' },
        onAdd: function(map) {
            var div = L.DomUtil.create('div', '');
            div.innerHTML = `
                <div style="background: rgba(255, 255, 255, 0.95); padding: 12px 16px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.15); margin: 10px; min-width: 250px; font-size: 11px; border-left: 4px solid #1e3a8a; font-family: Arial, sans-serif;">
                    
                    <!-- Encabezado -->
                    <div style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #e2e8f0; padding-bottom: 4px; margin-bottom: 8px;">
                        <h4 style="margin: 0; color: #1e3a8a; font-size: 12px; text-transform: uppercase; font-weight: bold;">
                            📊 Conteo y Filtros
                        </h4>
                        <button onclick="limpiarFiltrosFlora()" style="background: #e2e8f0; border: none; border-radius: 4px; padding: 2px 6px; font-size: 10px; cursor: pointer; color: #334155; font-weight: bold;">
                            🔄 Limpiar
                        </button>
                    </div>
                    
                    <!-- Desplegables de Filtros -->
                    <div style="margin-bottom: 8px; display: flex; flex-direction: column; gap: 4px;">
                        <select id="filtro-ubicacion" onchange="ejecutarFiltroFlora()" style="width: 100%; font-size: 11px; padding: 3px 4px; border-radius: 4px; border: 1px solid #cbd5e1;">
                            <option value="">-- Todas las Ubicaciones --</option>
                            ${ubicaciones.map(u => `<option value="${u}">${u}</option>`).join('')}
                        </select>
                        <select id="filtro-nombre-comun" onchange="ejecutarFiltroFlora()" style="width: 100%; font-size: 11px; padding: 3px 4px; border-radius: 4px; border: 1px solid #cbd5e1;">
                            <option value="">-- Todos los Nombres Comunes --</option>
                            ${nombresComunes.map(n => `<option value="${n}">${n}</option>`).join('')}
                        </select>
                        <select id="filtro-tipo-veg" onchange="ejecutarFiltroFlora()" style="width: 100%; font-size: 11px; padding: 3px 4px; border-radius: 4px; border: 1px solid #cbd5e1;">
                            <option value="">-- Tipos de Vegetación --</option>
                            ${tiposVegetacion.map(t => `<option value="${t}">${t}</option>`).join('')}
                        </select>
                    </div>

                    <!-- Métricas -->
                    <div style="display: flex; justify-content: space-between; margin-bottom: 4px; color: #334155; border-bottom: 1px dashed #cbd5e1; padding-bottom: 4px;">
                        <span>Especies Únicas:</span>
                        <strong id="stat-flora-especies" style="color: #0f172a;">0</strong>
                    </div>

                    <!-- Contenedor dinámico de tipos de vegetación -->
                    <div id="contenedor-metricas-flora"></div>

                    <div style="display: flex; justify-content: space-between; margin-top: 6px; padding-top: 4px; border-top: 1px solid #cbd5e1; font-weight: bold; color: #1e3a8a;">
                        <span>Individuos Visibles:</span>
                        <span id="stat-flora-total">0</span>
                    </div>
                </div>
            `;
            L.DomEvent.disableClickPropagation(div);
            L.DomEvent.disableScrollPropagation(div);
            return div;
        }
    });

    panelConteoFloraControl = new PanelConteo();
    map.addControl(panelConteoFloraControl);
}

// --- Funciones auxiliares del panel ---
function removerPanelConteoFlora() {
    if (panelConteoFloraControl) {
        map.removeControl(panelConteoFloraControl);
        panelConteoFloraControl = null;
    }
}

function limpiarFiltrosFlora() {
    const selUbi = document.getElementById('filtro-ubicacion');
    const selNom = document.getElementById('filtro-nombre-comun');
    const selTipo = document.getElementById('filtro-tipo-veg');

    if (selUbi) selUbi.value = "";
    if (selNom) selNom.value = "";
    if (selTipo) selTipo.value = "";

    ejecutarFiltroFlora();
}


// =======================================================
// CONTROL DE EDICION DE FLORA Y GUARDADO EN GOOGLE SHEETS
// =======================================================

var WEB_APP_URL = 'https://script.google.com/macros/s/AKfycbxGSndXYFuJj3NxpW7TLz3USQ4EgNqWkK6NctE2oKYXRE3Q8XOZ5f6mQCgOc7G1hPj0/exec';
var modoEdicionFlora = false;

// Botón flotante de Edición ✏️ en Leaflet
var BotonEditarFlora = L.Control.extend({
    options: { position: 'topleft' },
    onAdd: function (map) {
        var div = L.DomUtil.create('div', 'leaflet-bar');
        div.innerHTML = '<button id="btn-editar-flora" title="Activar/Desactivar Edición" style="width: 34px; height: 34px; background: white; border: none; font-size: 18px; cursor: pointer; line-height: 30px;">✏️</button>';
        return div;
    }
});
map.addControl(new BotonEditarFlora());

// Activar o desactivar edición
document.addEventListener('click', function(e) {
    if (e.target && e.target.id === 'btn-editar-flora') {
        modoEdicionFlora = !modoEdicionFlora;

        capaNuevaFlora.eachLayer(function(marcador) {
            if (marcador instanceof L.Marker) {
                if (modoEdicionFlora) {
                    marcador.dragging.enable();
                } else {
                    marcador.dragging.disable();
                }
            }
        });

        alert(modoEdicionFlora 
            ? 'MODO EDICIÓN ACTIVADO:\n\n1. Arrastra cualquier árbol.\n2. Presiona "💾 Guardar en Google Sheets" en el panel lateral derecho.' 
            : 'Modo edición desactivado.');
    }
});

// Enviar las nuevas coordenadas a Google Sheets
function guardarCoordenadasEnSheets(nro, lat, lng) {
    var latFormatted = String(lat).replace('.', ',');
    var lngFormatted = String(lng).replace('.', ',');

    var urlConParametros = WEB_APP_URL + '?nro=' + encodeURIComponent(nro) + '&lat=' + encodeURIComponent(latFormatted) + '&lng=' + encodeURIComponent(lngFormatted);

    fetch(urlConParametros)
        .then(function(respuesta) { return respuesta.text(); })
        .then(function(resultado) {
            var res = resultado.trim();
            if (res === 'OK') {
                alert('✅ ¡Guardado con éxito en Google Sheets!\n\nÁrbol N° ' + nro + '\nLat: ' + latFormatted + '\nLng: ' + lngFormatted);
            } else if (res === 'NOT_FOUND') {
                alert('❌ Error: El Árbol N° "' + nro + '" no existe en la Columna A de la pestaña "catastro".');
            } else {
                alert('⚠️ Respuesta de Google Sheets: ' + res);
            }
        })
        .catch(function(error) {
            console.error('Error de red:', error);
            alert('❌ Error de red: No se pudo conectar con Google Sheets.');
        });
}

// ==========================================
// 5. Botón Maestro: GESTIÓN DE AGUA
// ==========================================
var chkTodoAgua = document.getElementById('chk-todo-agua');
if (chkTodoAgua) {
    chkTodoAgua.addEventListener('change', function() {
        var estado = this.checked;
        var aguaIds = ['chk-bebed-fuente', 'chk-bebed-llenador', 'chk-bebed-nuevo', 'chk-bebed-deterioro', 'chk-bebed-baja'];
        aguaIds.forEach(function(id) {
            var cb = document.getElementById(id);
            if (cb && cb.checked !== estado) {
                cb.checked = estado;
                cb.dispatchEvent(new Event('change')); 
            }
        });
    });
}

// ==========================================
// 9. CONTROLADOR DE ARRANQUE Y CARGA DE ARCHIVOS (Limpio y Único)
// ==========================================

// Lanzar cargas iniciales
cargarPuntosEcologicos(); 
cargarPuntosFlora(); 
// ✨ CAMBIO 4: Lanzar la carga del nuevo CSV de cafetos al inicializar el mapa
cargarPuntosCafetos(); 
cargarDatos();
cargarJardinesReserva();

fetch('jefe_de_grupo.json')
    .then(response => {
        if (!response.ok) throw new Error("No se pudo cargar el archivo");
        return response.json();
    })
    .then(data => {

        // =========================================================
        // PASO 1: PRIMERA PASADA - Calcular los totales globales por Jefe
        // =========================================================
        data.features.forEach(function(feature) {
            if (feature.properties) {
                var props = feature.properties;
                
                // Extraemos el nombre del jefe de la fila actual
                var jefeOriginal = props.jefes || props.jefe || "No especificado";
                var textoJefe = jefeOriginal.toString().toLowerCase().trim();
                
                // Determinamos a qué categoría pertenece
                var categoriaDestino = "otros";
                if (textoJefe.includes("oscar") || textoJefe.includes("óscar")) categoriaDestino = "oscar";
                else if (textoJefe.includes("andres") || textoJefe.includes("andrés")) categoriaDestino = "andres";
                else if (textoJefe.includes("alfonso")) categoriaDestino = "alfonso";
                else if (textoJefe.includes("campo") || textoJefe.includes("depo")) categoriaDestino = "campos";
                else if (textoJefe.includes("bosque") || textoJefe.includes("húme") || textoJefe.includes("hume")) categoriaDestino = "bosque";

                // Buscamos e identificamos las llaves de área y perímetro (limpiando tildes)
                var areaOriginal = 0;
                var perimOriginal = 0;
                for (var llave in props) {
                    if (props.hasOwnProperty(llave)) {
                        var llaveLimpia = llave.toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "");
                        if (llaveLimpia === "area") areaOriginal = parseFloat(props[llave]) || 0;
                        if (llaveLimpia === "perimetro" || llaveLimpia === "perimeter") perimOriginal = parseFloat(props[llave]) || 0;
                    }
                }

                // Acumulamos directamente en nuestro objeto global 'totalesPorJefe'
                totalesPorJefe[categoriaDestino].area += areaOriginal;
                totalesPorJefe[categoriaDestino].perimetro += perimOriginal;
            }
        });

        // =========================================================
        // PASO 2: SEGUNDA PASADA - Construir las capas y los Popups
        // =========================================================
        var capaBaseGeoJSON = L.geoJSON(data, {
            style: function(feature) {
                var valorJefe = "";
                if (feature.properties) {
                    valorJefe = feature.properties.jefes || feature.properties.jefe || feature.properties.name || "";
                }
                return obtenerEstiloPoligono(valorJefe); // Aplica tus colores
            },
            onEachFeature: function(feature, layer) {
                if (feature.properties) {
                    var props = feature.properties;

                    var jefeOriginal = props.jefes || props.jefe || "No especificado";
                    var usoZona = props.Uso || "Área Verde";
                    var nombreSector = props.Nombre || "Sector sin nombre";

                    // Identificamos de nuevo la categoría para ir a buscar su total ya calculado
                    var textoJefe = jefeOriginal.toString().toLowerCase().trim();
                    var categoriaDestino = "otros";

                    if (textoJefe.includes("oscar") || textoJefe.includes("óscar")) categoriaDestino = "oscar";
                    else if (textoJefe.includes("andres") || textoJefe.includes("andrés")) categoriaDestino = "andres";
                    else if (textoJefe.includes("alfonso")) categoriaDestino = "alfonso";
                    else if (textoJefe.includes("campo") || textoJefe.includes("depo")) categoriaDestino = "campos";
                    else if (textoJefe.includes("bosque") || textoJefe.includes("húme") || textoJefe.includes("hume")) categoriaDestino = "bosque";

                    // 🎯 Extraemos el Gran Total Acumulado de este Jefe
                    var datosJefeGlobal = totalesPorJefe[categoriaDestino];
                    var areaTotalDelJefe = datosJefeGlobal.area;
                    var perimetroTotalDelJefe = datosJefeGlobal.perimetro;
                    var nombreOficialJefe = datosJefeGlobal.nombreOficial;

                    // Clasificamos en los LayerGroups visuales correspondientes
                    if (categoriaDestino === "oscar" && typeof capasresponsable !== 'undefined') capasresponsable.addLayer(layer);
                    if (categoriaDestino === "andres" && typeof capasresponsable !== 'undefined') capasresponsable.addLayer(layer);
                    if (categoriaDestino === "alfonso" && typeof capasresponsable !== 'undefined') capasresponsable.addLayer(layer);
                    if (categoriaDestino === "campos" && typeof capasCamposDepo !== 'undefined') capasCamposDepo.addLayer(layer);
                    if (categoriaDestino === "bosque" && typeof capasBosqueHumedo !== 'undefined') capasBosqueHumedo.addLayer(layer);

                    // Formateamos los números globales para el diseño del Popup
                    var txtAreaGlobal = areaTotalDelJefe > 0 ? `${areaTotalDelJefe.toLocaleString('es-PE', {maximumFractionDigits: 2})} m²` : "No registrado";
                    var txtPerimGlobal = perimetroTotalDelJefe > 0 ? `${perimetroTotalDelJefe.toLocaleString('es-PE', {maximumFractionDigits: 2})} m` : "No registrado";

                    // Popup Dinámico: Muestra datos del sector, pero la métrica es la SUMA TOTAL del Jefe
                    var popupContenido = `
                        <div style="padding:6px; font-size:12px; width:240px; line-height:1.4; color:#2d3748;">
                            <span style="font-size:9px; text-transform:uppercase; color:#8e44ad; font-weight:bold; display:block; margin-bottom:2px;">👨‍💼 Resumen por Responsable</span>
                            <h3 style="margin:0 0 6px 0; font-size:14px; color:#2c3e50; border-bottom:2px solid #8e44ad; padding-bottom:4px; font-weight:bold;">🟢 Cargo de: ${nombreOficialJefe}</h3>
                            
                            <p style="margin:3px 0;"><b>📍 Sector Actual:</b> ${nombreSector}</p>
                            <p style="margin:3px 0;"><b>🏛️ Uso de Zona:</b> ${usoZona}</p>
                            
                            <hr style="border:0; border-top:1px dashed #b2bec3; margin:6px 0;">
                            
                            <p style="margin:2px 0;"><b>📐 Área Total Asignada:</b> <br><span style="color:#27ae60; font-size:13px; font-weight:bold;">${txtAreaGlobal}</span></p>
                            <p style="margin:2px 0;"><b>📏 Perímetro Total Asignada:</b> <br><span style="color:#2980b9; font-size:13px; font-weight:bold;">${txtPerimGlobal}</span></p>
                        </div>
                    `;
                    layer.bindPopup(popupContenido);
                }
            }
        });

        // Opcional: También imprimimos el reporte ordenado en la consola del navegador
        mostrarResumenMétricas();
    })
    .catch(error => {
        console.error("Error al procesar y agrupar las áreas por jefe:", error);
    });

// Carga única de Fauna (.geojson)
fetch('fauna.geojson')
    .then(r => { if(r.ok) return r.json(); throw new Error(); })
    .then(d => capaFaunaGeoJSON.addData(d))
    .catch(e => console.log("fauna.geojson aún cargando o no encontrado."));

// ==========================================================================
// 12. NUEVAS CAPAS CON POPUPS: ÁREAS VERDES, PUERTAS, ESTACIONAMIENTOS Y XEROFÍTICAS
// ==========================================================================

// --- 1. DECLARACIÓN DE CAPAS ---

var capaPuertasEntradas = L.layerGroup();
var capaPlayasEstacionamiento = L.layerGroup();
var capaXerofitica = L.layerGroup(); // Nueva capa Xerofítica

// Subcapas independientes según Uso de Suelo
var subcapasAreasVerdesUso = {
    institucional: L.layerGroup(),
    administrativo: L.layerGroup(),
    recreativo: L.layerGroup(),
    sostenible: L.layerGroup(),
    otros: L.layerGroup()
};

// Subcapas independientes según Riego Actual
var subcapasAreasVerdesRiego = {
    gotero: L.layerGroup(),
    aspersion: L.layerGroup(),
    cisterna: L.layerGroup(),
    manual: L.layerGroup(),
    sin_riego: L.layerGroup(),
    otros: L.layerGroup()
};


// --- 2. FUNCIONES AUXILIARES DE CLASIFICACIÓN Y COLOR ---

// A. Clasificación por Uso de Suelo
function obtenerCategoriaUso(uso) {
    if (!uso) return "otros";
    const val = uso.toString().toLowerCase().trim();
    
    if (val.includes("institucional")) return "institucional";
    if (val.includes("administrativo")) return "administrativo";
    if (val.includes("recreativo") || val.includes("descanso")) return "recreativo";
    if (val.includes("sostenible") || val.includes("reducción") || val.includes("consumo")) return "sostenible";
    return "otros";
}

function obtenerColorPorUso(cat) {
    switch (cat) {
        case "institucional": return "#16a085"; // Turquesa / Verde Oscuro
        case "administrativo": return "#e67e22"; // Naranja
        case "recreativo":     return "#2ecc71"; // Verde Esmeralda
        case "sostenible":     return "#3498db"; // Azul
        default:               return "#95a5a6"; // Gris (Otros)
    }
}

// B. Clasificación por Riego Actual
function obtenerCategoriaRiego(riego) {
    if (!riego) return "sin_riego";
    const val = riego.toString().toLowerCase().trim();

    if (val.includes("goteo") || val.includes("gotero")) return "gotero";
    if (val.includes("aspersi") || val.includes("aspersor")) return "aspersion";
    if (val.includes("cisterna") || val.includes("camion")) return "cisterna";
    if (val.includes("manual") || val.includes("manguera")) return "manual";
    if (val.includes("sin") || val.includes("ninguno") || val.includes("no")) return "sin_riego";
    return "otros";
}

function obtenerColorPorRiego(cat) {
    switch (cat) {
        case "gotero":    return "#00cec9"; // Turquesa / Cian
        case "aspersion": return "#0984e3"; // Azul
        case "cisterna":  return "#6c5ce7"; // Púrpura
        case "manual":    return "#fdcb6e"; // Amarillo / Ámbar
        case "sin_riego": return "#d63031"; // Rojo
        default:          return "#b2bec3"; // Gris (Otros)
    }
}


// --- 3. CARGA DE GEOJSONS ---

// A. ÁREAS VERDES (CLASIFICADAS POR USO DE SUELO Y RIEGO ACTUAL)
fetch('areas_verdes.geojson')
    .then(r => { if(r.ok) return r.json(); throw new Error("No se pudo cargar areas_verdes.geojson"); })
    .then(d => {
        
        // Carga para Vista de Uso de Suelo
        L.geoJSON(d, {
            style: function(feature) {
                const uso = feature.properties ? feature.properties.Uso : '';
                const cat = obtenerCategoriaUso(uso);
                const color = obtenerColorPorUso(cat);

                return { 
                    color: color, 
                    fillColor: color, 
                    fillOpacity: 0.6, 
                    weight: 1.5 
                };
            },
            onEachFeature: function(feature, layer) {
                if (feature.properties) {
                    vincularPopupAreaVerde(feature, layer);

                    const catUso = obtenerCategoriaUso(feature.properties.Uso);
                    if (subcapasAreasVerdesUso[catUso]) {
                        subcapasAreasVerdesUso[catUso].addLayer(layer);
                    }
                }
            }
        });

        // Carga para Vista de Riego Actual
        L.geoJSON(d, {
            style: function(feature) {
                const riego = feature.properties ? feature.properties["Riego act"] : '';
                const cat = obtenerCategoriaRiego(riego);
                const color = obtenerColorPorRiego(cat);

                return { 
                    color: color, 
                    fillColor: color, 
                    fillOpacity: 0.6, 
                    weight: 1.5 
                };
            },
            onEachFeature: function(feature, layer) {
                if (feature.properties) {
                    vincularPopupAreaVerde(feature, layer);

                    const catRiego = obtenerCategoriaRiego(feature.properties["Riego act"]);
                    if (subcapasAreasVerdesRiego[catRiego]) {
                        subcapasAreasVerdesRiego[catRiego].addLayer(layer);
                    }
                }
            }
        });
    })
    .catch(e => console.log(e.message));

// Función para crear Popups de Áreas Verdes
function vincularPopupAreaVerde(feature, layer) {
    var props = feature.properties;
    var nombre = props.Nombre || "Zona sin nombre";
    var codigo = props["código"] || "S/C";
    var area = props["Área"] ? parseFloat(props["Área"]).toFixed(2) : "0.00";
    var perimetro = props["Perimetro"] ? parseFloat(props["Perimetro"]).toFixed(2) : "0.00";
    var uso = props.Uso || "No especificado";
    var riegoAct = props["Riego act"] || "No registrado";
    var proyRiego = props["Proy riego"] || "No registrado";

    var popupContenido = `
        <div style="padding: 4px; font-size: 11px; width: 260px; line-height: 1.4; color: #2d3748; font-family: Arial, sans-serif; box-sizing: border-box;">
            <span style="font-size: 9px; text-transform: uppercase; color: #27ae60; font-weight: bold; display: block; margin-bottom: 2px;">🗺️ Área Verde Campus (Código: ${codigo})</span>
            <h3 style="margin: 0 0 8px 0; font-size: 13px; color: #1a365d; border-bottom: 2px solid #27ae60; padding-bottom: 4px; font-weight: bold;">
                ${nombre}
            </h3>
            <table style="width: 100%; border-collapse: collapse;">
                <tr style="background-color: #f7fafc;"><td style="padding: 3px; font-weight: bold; width: 45%;">📐 Área:</td><td style="padding: 3px;">${area} m²</td></tr>
                <tr><td style="padding: 3px; font-weight: bold;">🔄 Perímetro:</td><td style="padding: 3px;">${perimetro} m</td></tr>
                <tr style="background-color: #f7fafc;"><td style="padding: 3px; font-weight: bold;">🏛️ Uso Suelo:</td><td style="padding: 3px;">${uso}</td></tr>
                <tr><td colspan="2" style="padding: 6px 3px 2px 3px; font-weight: bold; color: #4a5568; font-size: 9px; text-transform: uppercase; border-bottom: 1px solid #edf2f7;">💧 Sistema de Riego</td></tr>
                <tr style="background-color: #f7fafc;"><td style="padding: 3px; font-weight: bold;">Riego Actual:</td><td style="padding: 3px; color: #2c3e50;">${riegoAct}</td></tr>
                <tr><td style="padding: 3px; font-weight: bold;">Proy. Riego:</td><td style="padding: 3px; color: #7f8c8d; font-style: italic;">${proyRiego}</td></tr>
            </table>
        </div>`;

    layer.bindPopup(popupContenido, { maxWidth: 280, minWidth: 250 });
}


// B. PUERTAS Y ENTRADAS
fetch('puertas_entradas.geojson')
    .then(r => { if(r.ok) return r.json(); throw new Error(); })
    .then(d => {
        L.geoJSON(d, {
            pointToLayer: function(feature, latlng) {
                return L.marker(latlng, {
                    icon: L.divIcon({
                        className: 'puerta-marker',
                        html: `<div style="font-size:20px;">🚪</div>`,
                        iconSize: [24, 24]
                    })
                });
            },
            onEachFeature: function(feature, layer) {
                if (feature.properties) {
                    var props = feature.properties;
                    var popupContenido = `
                        <div style="padding:4px; font-size:12px; line-height:1.4; color:#2d3748;">
                            <strong style="color:#2980b9; font-size:13px;">🚪 Control de Acceso</strong><br>
                            <b>📍 Entrada:</b> ${props.Nombre || "Sin nombre"}<br>
                            <b>🛠️ Tipo:</b> ${props.Tipo || "No especificado"}<br>
                            <b>🛡️ Seguridad:</b> ${props.Control || "No registrado"}
                        </div>`;
                    layer.bindPopup(popupContenido);
                }
            }
        }).addTo(capaPuertasEntradas);
    })
    .catch(e => console.log("puertas_entradas.geojson no encontrado."));


// C. PLAYAS DE ESTACIONAMIENTO
fetch('playas_de_estacionamiento.geojson')
    .then(r => { if(r.ok) return r.json(); throw new Error(); })
    .then(d => {
        L.geoJSON(d, {
            style: { color: "#34495e", fillColor: "#7f8c8d", fillOpacity: 0.6, weight: 2 },
            onEachFeature: function(feature, layer) {
                if (feature.properties) {
                    var props = feature.properties;
                    var popupContenido = `
                        <div style="padding:4px; font-size:12px; line-height:1.4; color:#2d3748;">
                            <strong style="color:#34495e; font-size:13px;">🚗 Estacionamiento</strong><br>
                            <b>📍 Nombre:</b> ${props.Nombre || "Sin nombre"}<br>
                            <b>📦 Capacidad:</b> ${props.Capacidad || "No especificada"}<br>
                            <b>🗺️ Zona:</b> ${props.Zona || "No registrada"}
                        </div>`;
                    layer.bindPopup(popupContenido);
                }
            }
        }).addTo(capaPlayasEstacionamiento);
    })
    .catch(e => console.log("playas_de_estacionamiento.geojson no encontrado."));


// D. XEROFÍTICAS (NUEVA CAPA)
fetch('xerofitica.geojson')
    .then(r => { if(r.ok) return r.json(); throw new Error("No se pudo cargar xerofitica.geojson"); })
    .then(d => {
        L.geoJSON(d, {
            style: function(feature) {
                return {
                    color: "#d35400",       // Naranja cobrizo / Terracota
                    fillColor: "#e67e22",   // Tono cálido para xerófitas
                    fillOpacity: 0.65,
                    weight: 1.8
                };
            },
            onEachFeature: function(feature, layer) {
                if (feature.properties) {
                    vincularPopupXerofitica(feature, layer);
                }
            }
        }).addTo(capaXerofitica);
    })
    .catch(e => console.log("xerofitica.geojson no encontrado o error al cargar."));

// Función para crear Popups de Xerofíticas
function vincularPopupXerofitica(feature, layer) {
    var props = feature.properties;
    
    var nombre = props.Nombre || props.nombre || "Zona Xerofítica";
    var codigo = props["código"] || props.Codigo || props.codigo || "S/C";
    var rawArea = props["Área"] || props.Area || props.area || props["ÁREA"] || 0;
    var rawPerimetro = props["Perimetro"] || props.perimetro || props.Perímetro || props["PERIMETRO"] || 0;
    
    var area = parseFloat(rawArea) ? parseFloat(rawArea).toFixed(2) : "0.00";
    var perimetro = parseFloat(rawPerimetro) ? parseFloat(rawPerimetro).toFixed(2) : "0.00";
    var uso = props.Uso || props.uso || "Vegetación de bajo riego";
    var tipo = props.Tipo || props.tipo || "Xerófitas / Cactáceas";

    var popupContenido = `
        <div style="padding: 4px; font-size: 11px; width: 260px; line-height: 1.4; color: #2d3748; font-family: Arial, sans-serif; box-sizing: border-box;">
            <span style="font-size: 9px; text-transform: uppercase; color: #d35400; font-weight: bold; display: block; margin-bottom: 2px;">🌵 Zona Xerofítica (Código: ${codigo})</span>
            <h3 style="margin: 0 0 8px 0; font-size: 13px; color: #a04000; border-bottom: 2px solid #e67e22; padding-bottom: 4px; font-weight: bold;">
                ${nombre}
            </h3>
            <table style="width: 100%; border-collapse: collapse;">
                <tr style="background-color: #fef5e7;"><td style="padding: 3px; font-weight: bold; width: 45%;">📐 Área:</td><td style="padding: 3px;">${area} m²</td></tr>
                <tr><td style="padding: 3px; font-weight: bold;">🔄 Perímetro:</td><td style="padding: 3px;">${perimetro} m</td></tr>
                <tr style="background-color: #fef5e7;"><td style="padding: 3px; font-weight: bold;">🏛️ Uso Suelo:</td><td style="padding: 3px;">${uso}</td></tr>
                <tr><td style="padding: 3px; font-weight: bold;">🌵 Tipo/Especie:</td><td style="padding: 3px; color: #d35400; font-weight: bold;">${tipo}</td></tr>
            </table>
        </div>`;

    layer.bindPopup(popupContenido, { maxWidth: 280, minWidth: 250 });
}


// --- 4. LÓGICA DE CONTROL DE INTERRUPTORES (EVENTOS DE CHECKBOXES) ---

// Auxiliar para sincronizar checkbox maestro según hijos
function actualizarEstadoMaestro(selectorSub, idMaestro) {
    let subs = document.querySelectorAll(selectorSub);
    let maestro = document.getElementById(idMaestro);
    if (!maestro) return;
    let anyChecked = Array.from(subs).some(c => c.checked);
    maestro.checked = anyChecked;
}

// A. Control Maestro - Uso de Suelo
var chkAreaMaestroUso = document.getElementById('chk-nuevas-areas');
var subMenuAreasUso = document.getElementById('sub-menu-areas-verdes');

if (chkAreaMaestroUso) {
    chkAreaMaestroUso.addEventListener('change', function() {
        var activo = this.checked;
        if (subMenuAreasUso) subMenuAreasUso.style.display = activo ? 'block' : 'none';

        document.querySelectorAll('.chk-sub-area').forEach(function(subChk) {
            subChk.checked = activo;
            var cat = subChk.getAttribute('data-uso');
            
            if (subcapasAreasVerdesUso[cat]) {
                if (activo) {
                    if (!map.hasLayer(subcapasAreasVerdesUso[cat])) subcapasAreasVerdesUso[cat].addTo(map);
                } else {
                    if (map.hasLayer(subcapasAreasVerdesUso[cat])) map.removeLayer(subcapasAreasVerdesUso[cat]);
                }
            }
        });
    });
}

// Sub-Botones Individuales - Uso de Suelo
document.querySelectorAll('.chk-sub-area').forEach(function(subChk) {
    subChk.addEventListener('change', function() {
        var cat = this.getAttribute('data-uso');
        var activo = this.checked;

        if (subcapasAreasVerdesUso[cat]) {
            if (activo) {
                if (!map.hasLayer(subcapasAreasVerdesUso[cat])) subcapasAreasVerdesUso[cat].addTo(map);
            } else {
                if (map.hasLayer(subcapasAreasVerdesUso[cat])) map.removeLayer(subcapasAreasVerdesUso[cat]);
            }
        }
        actualizarEstadoMaestro('.chk-sub-area', 'chk-nuevas-areas');
    });
});

// B. Control Maestro - Riego Actual
var chkRiegoMaestro = document.getElementById('chk-riego-actual');
var subMenuRiego = document.getElementById('sub-menu-riego-actual');

if (chkRiegoMaestro) {
    chkRiegoMaestro.addEventListener('change', function() {
        var activo = this.checked;
        if (subMenuRiego) subMenuRiego.style.display = activo ? 'block' : 'none';

        document.querySelectorAll('.chk-sub-riego').forEach(function(subChk) {
            subChk.checked = activo;
            var cat = subChk.getAttribute('data-riego');
            
            if (subcapasAreasVerdesRiego[cat]) {
                if (activo) {
                    if (!map.hasLayer(subcapasAreasVerdesRiego[cat])) subcapasAreasVerdesRiego[cat].addTo(map);
                } else {
                    if (map.hasLayer(subcapasAreasVerdesRiego[cat])) map.removeLayer(subcapasAreasVerdesRiego[cat]);
                }
            }
        });
    });
}

// Sub-Botones Individuales - Riego Actual
document.querySelectorAll('.chk-sub-riego').forEach(function(subChk) {
    subChk.addEventListener('change', function() {
        var cat = this.getAttribute('data-riego');
        var activo = this.checked;

        if (subcapasAreasVerdesRiego[cat]) {
            if (activo) {
                if (!map.hasLayer(subcapasAreasVerdesRiego[cat])) subcapasAreasVerdesRiego[cat].addTo(map);
            } else {
                if (map.hasLayer(subcapasAreasVerdesRiego[cat])) map.removeLayer(subcapasAreasVerdesRiego[cat]);
            }
        }
        actualizarEstadoMaestro('.chk-sub-riego', 'chk-riego-actual');
    });
});

// C. Vinculación de Puertas, Estacionamientos y Xerofítica
if (typeof vincularCapaPro === "function") {
    vincularCapaPro('chk-puertas', capaPuertasEntradas);
    vincularCapaPro('chk-estacionamientos', capaPlayasEstacionamiento);
    vincularCapaPro('chk-xerofitica', capaXerofitica); // Vinculación Xerofítica igual a las demás
}

// ==========================================================================
// 13. BOTÓN MAESTRO INDEPENDIENTE PARA ADICIONALES CAMPUS
// ==========================================================================
var chkTodoAdicionales = document.getElementById('chk-todo-adicionales');
if (chkTodoAdicionales) {
    chkTodoAdicionales.addEventListener('change', function() {
        var estado = this.checked;
        var adicionalesIds = ['chk-nuevas-areas', 'chk-puertas', 'chk-estacionamientos', 'chk-xerofitica'];
        
        adicionalesIds.forEach(function(id) {
            var cb = document.getElementById(id);
            if (cb && cb.checked !== estado) {
                cb.checked = estado;
                cb.dispatchEvent(new Event('change')); 
            }
        });
    });
}

// ==========================================================================
// 🔍 14. MOTOR INTERACTIVO PUCP (BUSCADOR DINÁMICO EN TIEMPO REAL)
// ==========================================================================
var inputBusqueda = document.getElementById('input-busqueda-pucp') || document.getElementById('buscar') || document.getElementById('inputBuscador');
var listaSugerencias = document.getElementById('sugerencias-busqueda') || document.getElementById('sugerencias');

// Función que dibuja en el mapa SOLO los puntos coincidentes
function renderizarPuntosFiltrados(listaLugares) {
    if (typeof capaPuntosPUCP === 'undefined' || typeof map === 'undefined') return;

    // 1. Limpiamos cualquier marcador dibujado previamente
    capaPuntosPUCP.clearLayers();

    // Si la lista está vacía, no pintamos nada y devolvemos la vista general al campus
    if (!listaLugares || listaLugares.length === 0) {
        map.setView([-12.0685, -77.0815], 16);
        return;
    }

    var puntosEncontrados = [];

    // 2. Renderizamos solo los coincidentes
    listaLugares.forEach(function(item) {
        var lat = item.lat;
        var lng = item.lng;
        var titulo = item.titulo;
        var telefono = item.telefono;
        var urlSitio = item.urlSitio;
        var fotoUrl = item.fotoUrl;

        // Estilo del icono (marcador rojo)
        var iconoHtml = L.divIcon({
            className: 'marcador-forzado-visible', 
            html: `<div style="
                width: 20px !important; 
                height: 20px !important; 
                background-color: #ff2a2a !important; 
                border: 3px solid #ffffff !important; 
                border-radius: 50% !important; 
                box-shadow: 0 0 8px rgba(0,0,0,0.6) !important;
                cursor: pointer !important;
            "></div>`,
            iconSize: [20, 20],
            iconAnchor: [10, 10] 
        });

        var marcador = L.marker([lat, lng], { icon: iconoHtml });

        // Diseño del Popup
        var contenidoPopup = `
            <div style="width: 240px; font-family: sans-serif; font-size: 12px; line-height: 1.4;">
                <h4 style="margin: 0 0 6px 0; color: #1a73e8; font-size: 14px; font-weight: bold; border-bottom: 1px solid #eee; padding-bottom: 4px;">${titulo}</h4>
        `;
        
        if (fotoUrl && fotoUrl.startsWith("http")) {
            contenidoPopup += `
                <div style="width: 100%; height: 120px; overflow: hidden; border-radius: 4px; background-color: #f0f0f0; margin-bottom: 6px;">
                    <img src="${fotoUrl}" style="width: 100%; height: 100%; object-fit: cover;" alt="${titulo}">
                </div>
            `;
        }
        
        contenidoPopup += `<p style="margin: 4px 0;"><strong>📞 Teléfono:</strong> ${telefono}</p>`;
        
        if (urlSitio) {
            contenidoPopup += `
                <p style="margin: 6px 0 0 0; text-align: right;">
                    <a href="${urlSitio}" target="_blank" style="color: #1a73e8; text-decoration: none; font-weight: bold; font-size: 11px;">Ver en Google Maps ➔</a>
                </p>
            `;
        }
        contenidoPopup += `</div>`;

        marcador.bindPopup(contenidoPopup, { maxWidth: 260 });
        
        // Guardamos la referencia del marcador dentro del item para abrir su popup al hacer clic en sugerencias
        item.instanciaMarcador = marcador;

        capaPuntosPUCP.addLayer(marcador);
        puntosEncontrados.push([lat, lng]);
    });

    if (!map.hasLayer(capaPuntosPUCP)) {
        capaPuntosPUCP.addTo(map);
    }

    // 3. Ajustamos la vista según los marcadores resultantes
    if (puntosEncontrados.length === 1) {
        map.setView(puntosEncontrados[0], 18);
    } else if (puntosEncontrados.length > 1) {
        var bounds = L.latLngBounds(puntosEncontrados);
        map.fitBounds(bounds, { padding: [50, 50] });
    }
}

// 🎯 EVENTO PRINCIPAL DEL BUSCADOR
if (inputBusqueda) {
    inputBusqueda.addEventListener('input', function() {
        var query = this.value.trim().toLowerCase();
        
        if (listaSugerencias) {
            listaSugerencias.innerHTML = '';
        }

        // Si se limpia el buscador o hay menos de 2 letras
        if (query.length < 2) {
            if (listaSugerencias) listaSugerencias.style.display = 'none';
            renderizarPuntosFiltrados([]); // Limpia el mapa de inmediato
            return;
        }

        var resultadosCoincidentes = [];

        // Filtramos buscando en la memoria (bancoDatosPUCP)
        if (typeof bancoDatosPUCP !== 'undefined') {
            bancoDatosPUCP.forEach(function(fila) {
                var titulo = fila["title"] || fila["searchString"] || "Lugar PUCP";
                
                if (titulo.toLowerCase().includes(query)) {
                    var latRaw = fila["location/lat"] || fila["latitud"] || fila["lat"] || "";
                    var lngRaw = fila["location/lng"] || fila["longitud"] || fila["lng"] || "";

                    if (latRaw && lngRaw) {
                        var lat = parseFloat(latRaw.toString().trim().replace(/,/g, "."));
                        var lng = parseFloat(lngRaw.toString().trim().replace(/,/g, "."));

                        if (!isNaN(lat) && !isNaN(lng)) {
                            var fotoRaw = fila["imageOfPlace"] || fila["image"] || "";
                            
                            resultadosCoincidentes.push({
                                titulo: titulo,
                                lat: lat,
                                lng: lng,
                                telefono: fila["phone"] || "No disponible",
                                urlSitio: fila["url"] || fila["searchPageUrl"] || "",
                                fotoUrl: typeof transformarEnlaceDrive === 'function' ? transformarEnlaceDrive(fotoRaw) : fotoRaw
                            });
                        }
                    }
                }
            });
        }

        // Dibuja en el mapa SOLO los que hicieron match con la búsqueda
        renderizarPuntosFiltrados(resultadosCoincidentes);

        // Si tienes una lista desplegable HTML (<ul id="sugerencias">) la alimentamos
        if (listaSugerencias) {
            if (resultadosCoincidentes.length === 0) {
                var li = document.createElement('li');
                li.textContent = "No se encontraron resultados en el campus";
                li.style.cssText = "padding: 10px; color: #888; list-style: none;";
                listaSugerencias.appendChild(li);
            } else {
                resultadosCoincidentes.forEach(function(item) {
                    var li = document.createElement('li');
                    li.style.cssText = "padding: 10px 15px; cursor: pointer; border-bottom: 1px solid #eee; list-style: none; background: white;";
                    li.innerHTML = `<strong>${item.titulo}</strong> <span style="font-size:11px; color:#ff2a2a; float:right;">📍 Ir al lugar</span>`;

                    li.addEventListener('click', function() {
                        map.setView([item.lat, item.lng], 19);
                        if (item.instanciaMarcador) {
                            item.instanciaMarcador.openPopup();
                        }
                        listaSugerencias.style.display = 'none';
                        inputBusqueda.value = item.titulo;
                    });
                    listaSugerencias.appendChild(li);
                });
            }
            listaSugerencias.style.display = 'block';
        }
    });
}

// 🛠️ HERRAMIENTA DE RECALIBRACIÓN EN VIVO POR CLIC
map.on('click', function(e) {
    var coords = e.latlng;
    L.popup()
        .setLatLng(coords)
        .setContent(`
            <div style="font-family: monospace; font-size: 11px;">
                <strong>📍 Coordenada exacta en tu pantalla:</strong><br>
                lat: ${coords.lat.toFixed(5)}<br>lng: ${coords.lng.toFixed(5)}
            </div>
        `)
        .openOn(map);
});

// ==========================================
// 15. CAPA GEOJSON: VEREDAS DE MAYOR TRÁNSITO / PELIGRO
// ==========================================

// Variable global y grupo de capas para las veredas
var capaVeredasPeligro = L.layerGroup();

// Estilo visual predefinido para las áreas de veredas
const estiloVeredasPeligro = {
    color: "#dc2626",        // Borde rojo intenso
    weight: 3,               // Ancho de línea
    opacity: 0.9,
    fillColor: "#ef4444",    // Relleno rojo translúcido
    fillOpacity: 0.45,
    dashArray: '5, 5'        // Línea punteada/discontinua para indicar precaución
};

// --- Función para cargar y dibujar el GeoJSON ---
async function cargarVeredasPeligro() {
    capaVeredasPeligro.clearLayers();

    try {
        const response = await fetch('area_vereda_peligro.geojson');
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        
        const data = await response.json();

        const geojsonLayer = L.geoJSON(data, {
            style: estiloVeredasPeligro,
            onEachFeature: function (feature, layer) {
                // Generar popup automático según las propiedades disponibles
                let contenido = `
                    <div style="font-family: Arial, sans-serif; font-size: 12px; min-width: 180px; padding: 2px;">
                        <div style="background: #dc2626; color: white; padding: 6px; border-radius: 4px; font-weight: bold; margin-bottom: 6px;">
                            ⚠️ Zona de Alto Tránsito
                        </div>
                `;

                if (feature.properties) {
                    const props = feature.properties;
                    const nombre = props.nombre || props.Name || props.sector || 'Vereda de Concurrencia Alta';
                    const obs = props.observacio || props.descripcion || props.description || 'Zona crítica de tránsito peatonal.';

                    contenido += `
                        <strong style="color: #0f172a; font-size: 13px;">${nombre}</strong>
                        <p style="margin: 4px 0 0 0; color: #475569;">${obs}</p>
                    `;
                }

                contenido += `</div>`;
                layer.bindPopup(contenido);

                // Efecto de resalte al pasar el cursor (Hover)
                layer.on({
                    mouseover: function (e) {
                        var l = e.target;
                        l.setStyle({ fillOpacity: 0.7, weight: 4 });
                    },
                    mouseout: function (e) {
                        geojsonLayer.resetStyle(e.target);
                    }
                });
            }
        });

        capaVeredasPeligro.addLayer(geojsonLayer);
        console.log("🚶 Layer 'Veredas de Alto Tránsito' cargada con éxito.");

    } catch (error) {
        console.error("Error al cargar el archivo GeoJSON de veredas:", error);
    }
}

// --- Listener para el Checkbox / Interruptor en la interfaz ---
var chkVeredasPeligro = document.getElementById('chk-veredas-peligro');

if (chkVeredasPeligro) {
    chkVeredasPeligro.addEventListener('change', function () {
        if (this.checked) {
            if (!map.hasLayer(capaVeredasPeligro)) {
                capaVeredasPeligro.addTo(map);
            }
            cargarVeredasPeligro();
        } else {
            if (map.hasLayer(capaVeredasPeligro)) {
                map.removeLayer(capaVeredasPeligro);
            }
            capaVeredasPeligro.clearLayers();
        }
    });
}

// Alternar visibilidad del Panel de Capas Derecho con Flechas
function togglePanelCapas() {
    const panel = document.getElementById('panel-capas-derecho');
    const flecha = document.getElementById('flecha-capas');

    if (!panel || !flecha) return;

    panel.classList.toggle('panel-colapsado');

    // Si está colapsado, mostrar flecha para abrir (❮); de lo contrario, mostrar flecha para replegar (❯)
    if (panel.classList.contains('panel-colapsado')) {
        flecha.innerHTML = '❮';
    } else {
        flecha.innerHTML = '❯';
    }
}

// Función para replegar/desplegar el panel derecho
function togglePanelCapas() {
    var panel = document.getElementById('panel-capas-derecho');
    var flecha = document.getElementById('flecha-capas');

    panel.classList.toggle('panel-colapsado');

    if (panel.classList.contains('panel-colapsado')) {
        flecha.style.transform = 'rotate(180deg)'; // Gira la flecha para indicar reapertura
    } else {
        flecha.style.transform = 'rotate(0deg)';
    }
}

// Bloquea los clics e interactividad del mapa al hacer clic dentro del panel
document.addEventListener('DOMContentLoaded', function () {
    var panel = document.getElementById('panel-capas-derecho');
    if (panel && typeof L !== 'undefined') {
        L.DomEvent.disableClickPropagation(panel);
        L.DomEvent.disableScrollPropagation(panel);
    }
});

// ==========================================================================
// 🧹 LÓGICA Y ANIMACIÓN PARA LA BARREDORA N° 1
// ==========================================================================

// Coordenadas completas actualizadas (Puntos 1 al 25) para la Barredora N° 1
const rutaBarredora1 = [
    [-12.06853, -77.07868], // Punto 1
    [-12.06859, -77.07846], // Punto 2
    [-12.06880, -77.07859], // Punto 3
    [-12.06930, -77.07863], // Punto 4
    [-12.06936, -77.07864], // Punto 5
    [-12.06937, -77.07860], // Punto 6
    [-12.06957, -77.07860], // Punto 7
    [-12.06997, -77.07859], // Punto 8
    [-12.07008, -77.07858], // Punto 9
    [-12.07017, -77.07859], // Punto 10
    [-12.07105, -77.07879], // Punto 11
    [-12.07100, -77.07907], // Punto 12
    [-12.07097, -77.07923], // Punto 13
    [-12.07100, -77.07907], // Punto 14
    [-12.07105, -77.07879], // Punto 15
    [-12.07145, -77.07888], // Punto 16
    [-12.07139, -77.07917], // Punto 17
    [-12.07134, -77.07965], // Punto 18
    [-12.07125, -77.07996], // Punto 19
    [-12.07125, -77.07997], // Punto 20
    [-12.07105, -77.07993], // Punto 21
    [-12.07059, -77.07988], // Punto 22
    [-12.06959, -77.07978], // Punto 23
    [-12.06868, -77.07968], // Punto 24
    [-12.06869, -77.07955]  // Punto 25
];

// Variables de control de la animación
let marcadorBarredora = null;
let lineaRutaBarredora = null;
let poligonoBarrido = null;
let animacionIntervalo = null;
let pasoActual = 0;
let enSimulacion = false;

// Convertir las coordenadas para Turf.js (Turf usa [Longitud, Latitud])
const rutaTurf = rutaBarredora1.map(coord => [coord[1], coord[0]]);
const lineaTurf = turf.lineString(rutaTurf);
const distanciaTotal = turf.length(lineaTurf, { units: 'kilometers' });

// Crear ícono personalizado para la barredora
const iconoBarredora = L.divIcon({
    className: 'custom-icono-barredora',
    html: '<div style="font-size: 24px; filter: drop-shadow(0px 2px 3px rgba(0,0,0,0.4));">🚜</div>',
    iconSize: [30, 30],
    iconAnchor: [15, 15]
});

// Referencias a los elementos del DOM
const chkBarredora1 = document.getElementById('chk-barredora-1');
const panelBarredora = document.getElementById('panel-control-barredora');
const btnIniciar = document.getElementById('btn-iniciar-barredora');
const btnReiniciar = document.getElementById('btn-reiniciar-barredora');
const btnCerrarPanel = document.getElementById('btn-cerrar-panel-barredora');
const txtEstado = document.getElementById('txt-estado-barredora');

// Evento al marcar / desmarcar el checkbox "Barredora N° 1"
chkBarredora1.addEventListener('change', function() {
    if (this.checked) {
        mostrarSimulacionBarredora();
    } else {
        ocultarSimulacionBarredora();
    }
});

function mostrarSimulacionBarredora() {
    // 1. Mostrar panel flotante de control
    panelBarredora.classList.remove('panel-resumen-oculto');
    panelBarredora.classList.add('panel-resumen-visible');

    // 2. Dibujar la línea de ruta proyectada en el mapa
    lineaRutaBarredora = L.polyline(rutaBarredora1, {
        color: '#16a34a',
        weight: 3,
        dashArray: '6, 8',
        opacity: 0.8
    }).addTo(map);

    // 3. Crear el marcador de la barredora en la primera posición
    marcadorBarredora = L.marker(rutaBarredora1[0], { icon: iconoBarredora }).addTo(map);
    marcadorBarredora.bindPopup("<b>🚜 Barredora N° 1</b><br>Estado: En espera").openPopup();

    // Enfocar el mapa suavemente en la barredora
    map.flyTo(rutaBarredora1[0], 17);
}

function ocultarSimulacionBarredora() {
    detenerSimulacion();

    // Remover elementos visuales del mapa
    if (marcadorBarredora) map.removeLayer(marcadorBarredora);
    if (lineaRutaBarredora) map.removeLayer(lineaRutaBarredora);
    if (poligonoBarrido) map.removeLayer(poligonoBarrido);

    marcadorBarredora = null;
    lineaRutaBarredora = null;
    poligonoBarrido = null;

    // Ocultar panel flotante
    panelBarredora.classList.remove('panel-resumen-visible');
    panelBarredora.classList.add('panel-resumen-oculto');
}

// Botón Reproducir / Pausar
btnIniciar.addEventListener('click', function() {
    if (!enSimulacion) {
        iniciarSimulacion();
    } else {
        pausarSimulacion();
    }
});

// Botón Reiniciar
btnReiniciar.addEventListener('click', function() {
    reiniciarSimulacion();
});

// Botón Cerrar (X) del panel flotante
if (btnCerrarPanel) {
    btnCerrarPanel.addEventListener('click', function() {
        chkBarredora1.checked = false;
        ocultarSimulacionBarredora();
    });
}

function iniciarSimulacion() {
    enSimulacion = true;
    btnIniciar.textContent = '⏸ Pausar Simulación';
    btnIniciar.style.backgroundColor = '#d97706'; // Naranja al pausar
    txtEstado.textContent = 'Estado: Limpiando campus...';

    const pasosTotales = 300; // Cuanto más alto, más suave la animación
    const intervaloMs = 50;   // Velocidad de refresco (ms)

    animacionIntervalo = setInterval(() => {
        pasoActual++;

        if (pasoActual > pasosTotales) {
            pasoActual = pasosTotales;
            pausarSimulacion();
            txtEstado.textContent = 'Estado: ✅ Barrido completado';
            btnIniciar.textContent = '▶ Iniciar Simulación';
            return;
        }

        // Calcular posición actual a lo largo de la ruta usando Turf.js
        const avanceKm = (pasoActual / pasosTotales) * distanciaTotal;
        const puntoActual = turf.along(lineaTurf, avanceKm, { units: 'kilometers' });
        const coordsPunto = [puntoActual.geometry.coordinates[1], puntoActual.geometry.coordinates[0]];

        // Mover el marcador del vehículo
        marcadorBarredora.setLatLng(coordsPunto);

        // Generar área "barrida" acumulada (Buffer de 12 metros a los lados)
        const subRutaCoord = [rutaTurf[0]];
        const segmentoPasado = turf.lineSliceAlong(lineaTurf, 0, avanceKm, { units: 'kilometers' });
        
        if (segmentoPasado) {
            const areaBarridaTurf = turf.buffer(segmentoPasado, 0.006, { units: 'kilometers' }); // 6 metros a cada lado
            
            if (poligonoBarrido) map.removeLayer(poligonoBarrido);
            
            poligonoBarrido = L.geoJSON(areaBarridaTurf, {
                style: {
                    color: '#22c55e',
                    fillColor: '#86efac',
                    fillOpacity: 0.5,
                    weight: 1
                }
            }).addTo(map);
        }

    }, intervaloMs);
}

function pausarSimulacion() {
    enSimulacion = false;
    clearInterval(animacionIntervalo);
    btnIniciar.textContent = '▶ Continuar Simulación';
    btnIniciar.style.backgroundColor = '#2e7d32';
    txtEstado.textContent = 'Estado: En pausa';
}

function detenerSimulacion() {
    enSimulacion = false;
    clearInterval(animacionIntervalo);
    pasoActual = 0;
    btnIniciar.textContent = '▶ Iniciar Simulación';
    btnIniciar.style.backgroundColor = '#2e7d32';
    txtEstado.textContent = 'Estado: En espera';
}

function reiniciarSimulacion() {
    detenerSimulacion();
    if (poligonoBarrido) map.removeLayer(poligonoBarrido);
    poligonoBarrido = null;
    if (marcadorBarredora) {
        marcadorBarredora.setLatLng(rutaBarredora1[0]);
    }
}

// ==========================================================================
// 🧹 LÓGICA Y ANIMACIÓN PARA LA BARREDORA N° 2
// ==========================================================================

// Coordenadas para la Barredora N° 2 (Ruta ajustada / alternativa)
const rutaBarredora2 = [
    [-12.06754, -77.08021], // Punto 1
    [-12.06750, -77.08000], // Punto 2
    [-12.06748, -77.07983], // Punto 3
    [-12.06749, -77.07973], // Punto 4
    [-12.06767, -77.07971], // Punto 5
    [-12.06794, -77.07970], // Punto 6
    [-12.06823, -77.07970], // Punto 7
    [-12.06846, -77.07969], // Punto 8
    [-12.06864, -77.07969], // Punto 9
    [-12.06846, -77.07969], // Punto 10
    [-12.06823, -77.07970], // Punto 11
    [-12.06794, -77.07970], // Punto 12
    [-12.06708, -77.07973], // Punto 13
    [-12.06664, -77.07977], // Punto 14
    [-12.06649, -77.07979]  // Punto 15
];

// Variables de control de la animación N° 2
let marcadorBarredora2 = null;
let lineaRutaBarredora2 = null;
let poligonoBarrido2 = null;
let animacionIntervalo2 = null;
let pasoActual2 = 0;
let enSimulacion2 = false;

// Convertir las coordenadas para Turf.js (Turf usa [Longitud, Latitud])
const rutaTurf2 = rutaBarredora2.map(coord => [coord[1], coord[0]]);
const lineaTurf2 = turf.lineString(rutaTurf2);
const distanciaTotal2 = turf.length(lineaTurf2, { units: 'kilometers' });

// Crear ícono personalizado para la Barredora N° 2 (diferenciado en color/estilo)
const iconoBarredora2 = L.divIcon({
    className: 'custom-icono-barredora-2',
    html: '<div style="font-size: 24px; filter: drop-shadow(0px 2px 3px rgba(0,0,0,0.4));">🚜</div>',
    iconSize: [30, 30],
    iconAnchor: [15, 15]
});

// Referencias a los elementos del DOM (Barredora N° 2)
const chkBarredora2 = document.getElementById('chk-barredora-2');
const panelBarredora2 = document.getElementById('panel-control-barredora-2');
const btnIniciar2 = document.getElementById('btn-iniciar-barredora-2');
const btnReiniciar2 = document.getElementById('btn-reiniciar-barredora-2');
const btnCerrarPanel2 = document.getElementById('btn-cerrar-panel-barredora-2');
const txtEstado2 = document.getElementById('txt-estado-barredora-2');

// Evento al marcar / desmarcar el checkbox "Barredora N° 2"
if (chkBarredora2) {
    chkBarredora2.addEventListener('change', function() {
        if (this.checked) {
            mostrarSimulacionBarredora2();
        } else {
            ocultarSimulacionBarredora2();
        }
    });
}

function mostrarSimulacionBarredora2() {
    // 1. Mostrar panel flotante de control
    if (panelBarredora2) {
        panelBarredora2.classList.remove('panel-resumen-oculto');
        panelBarredora2.classList.add('panel-resumen-visible');
    }

    // 2. Dibujar la línea de ruta proyectada en el mapa (Azul/Cian para diferenciarla de la N° 1)
    lineaRutaBarredora2 = L.polyline(rutaBarredora2, {
        color: '#0284c7',
        weight: 3,
        dashArray: '6, 8',
        opacity: 0.8
    }).addTo(map);

    // 3. Crear el marcador de la barredora N° 2 en la primera posición
    marcadorBarredora2 = L.marker(rutaBarredora2[0], { icon: iconoBarredora2 }).addTo(map);
    marcadorBarredora2.bindPopup("<b>🚜 Barredora N° 2</b><br>Estado: En espera").openPopup();

    // Enfocar el mapa suavemente en la barredora N° 2
    map.flyTo(rutaBarredora2[0], 17);
}

function ocultarSimulacionBarredora2() {
    detenerSimulacion2();

    // Remover elementos visuales del mapa
    if (marcadorBarredora2) map.removeLayer(marcadorBarredora2);
    if (lineaRutaBarredora2) map.removeLayer(lineaRutaBarredora2);
    if (poligonoBarrido2) map.removeLayer(poligonoBarrido2);

    marcadorBarredora2 = null;
    lineaRutaBarredora2 = null;
    poligonoBarrido2 = null;

    // Ocultar panel flotante
    if (panelBarredora2) {
        panelBarredora2.classList.remove('panel-resumen-visible');
        panelBarredora2.classList.add('panel-resumen-oculto');
    }
}

// Botón Reproducir / Pausar
if (btnIniciar2) {
    btnIniciar2.addEventListener('click', function() {
        if (!enSimulacion2) {
            iniciarSimulacion2();
        } else {
            pausarSimulacion2();
        }
    });
}

// Botón Reiniciar
if (btnReiniciar2) {
    btnReiniciar2.addEventListener('click', function() {
        reiniciarSimulacion2();
    });
}

// Botón Cerrar (X) del panel flotante
if (btnCerrarPanel2) {
    btnCerrarPanel2.addEventListener('click', function() {
        if (chkBarredora2) chkBarredora2.checked = false;
        ocultarSimulacionBarredora2();
    });
}

function iniciarSimulacion2() {
    enSimulacion2 = true;
    btnIniciar2.textContent = '⏸ Pausar Simulación';
    btnIniciar2.style.backgroundColor = '#d97706';
    if (txtEstado2) txtEstado2.textContent = 'Estado: Limpiando campus...';

    const pasosTotales = 300;
    const intervaloMs = 50;

    animacionIntervalo2 = setInterval(() => {
        pasoActual2++;

        if (pasoActual2 > pasosTotales) {
            pasoActual2 = pasosTotales;
            pausarSimulacion2();
            if (txtEstado2) txtEstado2.textContent = 'Estado: ✅ Barrido completado';
            btnIniciar2.textContent = '▶ Iniciar Simulación';
            return;
        }

        // Calcular posición actual a lo largo de la ruta N° 2
        const avanceKm = (pasoActual2 / pasosTotales) * distanciaTotal2;
        const puntoActual = turf.along(lineaTurf2, avanceKm, { units: 'kilometers' });
        const coordsPunto = [puntoActual.geometry.coordinates[1], puntoActual.geometry.coordinates[0]];

        // Mover el marcador del vehículo
        marcadorBarredora2.setLatLng(coordsPunto);

        // Generar área "barrida" acumulada (Buffer de 12 metros a los lados)
        const segmentoPasado = turf.lineSliceAlong(lineaTurf2, 0, avanceKm, { units: 'kilometers' });
        
        if (segmentoPasado) {
            const areaBarridaTurf = turf.buffer(segmentoPasado, 0.006, { units: 'kilometers' });
            
            if (poligonoBarrido2) map.removeLayer(poligonoBarrido2);
            
            poligonoBarrido2 = L.geoJSON(areaBarridaTurf, {
                style: {
                    color: '#0284c7',
                    fillColor: '#7dd3fc',
                    fillOpacity: 0.5,
                    weight: 1
                }
            }).addTo(map);
        }

    }, intervaloMs);
}

function pausarSimulacion2() {
    enSimulacion2 = false;
    clearInterval(animacionIntervalo2);
    btnIniciar2.textContent = '▶ Continuar Simulación';
    btnIniciar2.style.backgroundColor = '#0284c7';
    if (txtEstado2) txtEstado2.textContent = 'Estado: En pausa';
}

function detenerSimulacion2() {
    enSimulacion2 = false;
    clearInterval(animacionIntervalo2);
    pasoActual2 = 0;
    btnIniciar2.textContent = '▶ Iniciar Simulación';
    btnIniciar2.style.backgroundColor = '#0284c7';
    if (txtEstado2) txtEstado2.textContent = 'Estado: En espera';
}

function reiniciarSimulacion2() {
    detenerSimulacion2();
    if (poligonoBarrido2) map.removeLayer(poligonoBarrido2);
    poligonoBarrido2 = null;
    if (marcadorBarredora2) {
        marcadorBarredora2.setLatLng(rutaBarredora2[0]);
    }
}
