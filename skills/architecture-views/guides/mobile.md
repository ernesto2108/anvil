# Template: architecture-mobile.md

Inspirado en: Flutter architecture guide + Android App Architecture + iOS Human Interface Guidelines.

**Generar cuando:** hay trabajo de mobile involucrado (Flutter, React Native, Swift, Kotlin).

## Template

```markdown
# Arquitectura Mobile — <TASK-ID>

## Plataformas objetivo

- [ ] iOS
- [ ] Android
- [ ] Ambas (Flutter / React Native)

## Patrones de comunicación con backend

<!-- Marcar cuáles aplican. Incluir solo esas secciones abajo. -->
- [ ] REST / HTTP
- [ ] WebSockets / SSE (tiempo real)
- [ ] gRPC
- [ ] Push notifications (FCM / APNs)
- [ ] Deep links / universal links
- [ ] Estado local únicamente (sin backend)

---

## Navegación

<!-- Stack-based, tab-based, drawer, modal flows -->

```mermaid
graph TD
  App --> TabBar
  TabBar --> HomeStack
  TabBar --> ProfileStack
  HomeStack --> DetailScreen
  DetailScreen --> EditModal
```

| Ruta / Screen | Stack | Guard (auth, permisos) | Deep link path |
|---|---|---|---|

### Deep linking — incluir si aplica
- **Scheme:** `myapp://` o universal links `https://domain.com/path`
- **Mapeo:** qué deep link abre qué screen
- **Fallback:** qué pasa si el deep link llega sin auth o con datos inválidos

---

## Contratos de tipos

<!-- Modelos Dart/Kotlin/Swift. Deben coincidir con contratos backend exactamente. -->

```dart
// Derivado de contratos REST/gRPC del backend
class ResponseDTO {
  final String id;
  // ...
}

// Estado de la feature
class FeatureState {
  final List<ResponseDTO> items;
  final bool loading;
  final String? error;
}
```

---

## Gestión de estado

<!-- BLoC, Riverpod, Provider, Redux, MobX — cuál y por qué -->
- **Patrón:** ...
- **Scope:** global vs por-feature vs por-screen
- **Persistencia de estado:** qué estado sobrevive kill de app (y dónde: SQLite, shared prefs, secure storage)

### Máquina de estados — incluir si la feature tiene más de 2 estados

```mermaid
stateDiagram-v2
  [*] --> Initial
  Initial --> Loading : fetch()
  Loading --> Loaded : datos recibidos
  Loading --> Error : request falló
  Error --> Loading : retry
  Loaded --> Refreshing : pull-to-refresh
  Refreshing --> Loaded : datos actualizados
```

---

## Estrategia offline — incluir si aplica

- **Nivel de soporte:** online-only / offline-read / offline-first
- **Almacenamiento local:** SQLite, Hive, shared preferences, secure storage — cuál y para qué datos
- **Sincronización:** estrategia de sync (queue local → push al reconectar, merge de conflictos)
- **Conflictos:** last-write-wins / merge manual / server-wins
- **Indicador de estado:** cómo el UI comunica modo offline al usuario

---

## Ciclo de vida de la app

- **Background:** qué pasa cuando la app va a background (pausar streams, guardar estado)
- **Foreground resume:** qué se revalida al volver (tokens, datos stale)
- **Kill y restore:** qué estado se restaura vs qué se re-fetches
- **Session management:** expiración de tokens, refresh strategy

---

## Push notifications — incluir si aplica

- **Proveedor:** FCM / APNs / ambos
- **Tipos de notificación:** data-only, notification+data, topic-based, targeted
- **Manejo en foreground:** mostrar in-app banner / silenciar / custom UI
- **Manejo en background:** qué procesa sin abrir la app
- **Deep link desde notificación:** a qué screen navega y con qué datos

---

## Permisos del dispositivo — incluir si aplica

| Permiso | Cuándo se solicita | Fallback si deniegan | Plataforma |
|---|---|---|---|
| Cámara | ... | ... | iOS + Android |
| Ubicación | ... | ... | iOS + Android |
| Notificaciones | ... | ... | iOS + Android |

---

## Capa de integración con backend

### REST / gRPC
- **Cliente:** dio / http / retrofit — cuál y por qué
- **Manejo de errores:** qué hace el UI ante errores de red, timeouts, 4xx, 5xx
- **Retry:** estrategia (exponential backoff, límite de intentos)
- **Cache de respuestas:** estrategia (cache-first, network-first, stale-while-revalidate)

### WebSockets / SSE — incluir si aplica
- **Reconexión:** backoff, límite, fallback a polling
- **Estado de conexión:** indicador visual para el usuario

---

## Consideraciones de plataforma — incluir si aplica

### Diferencias iOS vs Android
- **Navigation patterns:** ...
- **Permisos:** diferencias en el flujo de solicitud
- **Background execution:** límites por plataforma

### Platform channels / FFI — incluir si aplica
- **Canal:** qué operación nativa se expone
- **Serialización:** formato de datos entre Dart/JS y native
```

## Reglas

- Las clases/interfaces de tipos DEBEN coincidir con contratos backend — mismos nombres de campo, mismos tipos
- Incluir SOLO secciones que apliquen — omitir secciones vacías completamente
- La sección offline es obligatoria si la app debe funcionar sin conexión — "siempre online" debe declararse explícitamente
- La sección de push notifications es obligatoria si el backend emite eventos al dispositivo
- Permisos del dispositivo deben documentar el fallback cuando el usuario deniega — no solo "solicitar permiso"
- La estrategia de gestión de estado debe justificar el patrón elegido — no solo nombrar la librería
- Deep linking debe mapear cada ruta a un screen con su guard de auth
- Si existe vista backend, los modelos mobile se DERIVAN de esos contratos — no se definen independientemente
- Ciclo de vida es obligatorio para features que usan streams, timers, o estado temporal
