# Guía de Rendimiento Flutter

## Checklist de Rendimiento (Orden de Prioridad)

De Alibaba (100M+ usuarios), ByteDance (700+ devs), BMW (300 devs).

### 1. Widgets `const` — Hasta 30% de Mejora en Renderizado

```dart
// bad: recreated every build
child: Text('Hello')

// good: compile-time constant, never rebuilt
child: const Text('Hello')
```

**Reglas:**
- Marcar cada constructor de widget stateless como `const`
- Usar `const` para cualquier widget sin datos dinámicos
- El analizador de Dart advierte sobre `const` faltante — corregir todos los warnings

### 2. `ListView.builder` — Reducción de Memoria 50%+

```dart
// bad: builds all items at once (OOM for large lists)
ListView(
  children: items.map((item) => ItemCard(item: item)).toList(),
)

// good: builds only visible items
ListView.builder(
  itemCount: items.length,
  itemBuilder: (context, index) => ItemCard(item: items[index]),
)
```

También usar `ListView.separated` para listas con separadores.

### 3. `RepaintBoundary` — Aislar Repintados Costosos

```dart
// wrap around animations or frequently-updating UI
RepaintBoundary(
  child: AnimatedWidget(...),
)
```

Usar cuando: animaciones, timers, indicadores de progreso, o cualquier widget que se actualice independientemente de su padre.

### 4. Mantener `build()` Puro

```dart
// bad: API call in build
@override
Widget build(BuildContext context) {
  final data = await api.fetchData(); // NEVER do this
  return Text(data.name);
}

// bad: heavy computation in build
@override
Widget build(BuildContext context) {
  final sorted = items.sort((a, b) => a.name.compareTo(b.name)); // expensive
  return ListView(...);
}

// good: move to BLoC/ViewModel, build only renders
@override
Widget build(BuildContext context) {
  return BlocBuilder<DataBloc, DataState>(
    builder: (context, state) => switch (state) {
      DataLoaded(:final items) => ListView.builder(...),
      DataLoading() => const CircularProgressIndicator(),
      _ => const SizedBox.shrink(),
    },
  );
}
```

### 5. Reconstrucciones Granulares

```dart
// bad: entire tree rebuilds when one value changes
BlocBuilder<CartBloc, CartState>(
  builder: (context, state) => Column(
    children: [
      CartHeader(count: state.itemCount),    // rebuilds
      CartList(items: state.items),           // rebuilds
      CartTotal(total: state.total),          // rebuilds
    ],
  ),
)

// good: each widget rebuilds independently
Column(
  children: [
    BlocSelector<CartBloc, CartState, int>(
      selector: (state) => state.itemCount,
      builder: (context, count) => CartHeader(count: count),
    ),
    BlocSelector<CartBloc, CartState, List<CartItem>>(
      selector: (state) => state.items,
      builder: (context, items) => CartList(items: items),
    ),
    BlocSelector<CartBloc, CartState, double>(
      selector: (state) => state.total,
      builder: (context, total) => CartTotal(total: total),
    ),
  ],
)
```

### 6. Extraer Widgets, No Usar Métodos Helper

```dart
// bad: helper method — no rebuild isolation
class MyScreen extends StatelessWidget {
  Widget _buildHeader() {
    return Container(...);
  }

  Widget _buildBody() {
    return Container(...); // rebuilds when header changes
  }

  @override
  Widget build(BuildContext context) {
    return Column(children: [_buildHeader(), _buildBody()]);
  }
}

// good: separate widget class — independent rebuild boundary
class MyScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return const Column(children: [_Header(), _Body()]);
  }
}

class _Header extends StatelessWidget {
  const _Header();
  @override
  Widget build(BuildContext context) => Container(...);
}

class _Body extends StatelessWidget {
  const _Body();
  @override
  Widget build(BuildContext context) => Container(...);
}
```

### 7. Optimización de Imágenes (Alibaba: -300ms en Android Gama Baja)

```dart
// cache images
CachedNetworkImage(
  imageUrl: url,
  placeholder: (_, __) => const Skeleton(),
  errorWidget: (_, __, ___) => const Icon(Icons.error),
)

// proper sizing — don't load 4K images for thumbnails
Image.network(
  '$url?w=200&h=200', // request correct size from CDN
  width: 200,
  height: 200,
  fit: BoxFit.cover,
)
```

---

## Patrones de Alibaba (100M+ Usuarios)

- **Prefetch de datos**: Iniciar la carga de datos antes de navegar a la siguiente pantalla
- **Precarga de plantillas**: Pre-construir plantillas de widgets para pantallas comunes
- **Patrón Adapter para ListView**: Adapter personalizado en vez del modelo de componentes. Mejora de FPS: 40 → 53 en Android
- **Renderizado nativo de imágenes**: Renderizado directo con TextureID, evitando la copia de PixelBuffer

---

## Patrones de ByteDance (700+ Devs)

- **Eliminar librerías nativas no usadas**: Se eliminaron partes no utilizadas de Skia, BoringSSL, ICU, libwebp del motor Flutter
- **Compresión de sección de datos en iOS**: Reducción del tamaño del paquete de la app
- **Pipeline de renderizado personalizado**: Optimizado para sus casos de uso específicos
- **Resultado: ~33% de aumento de productividad** sobre el desarrollo nativo

---

## Profiling

### Flutter DevTools

```bash
flutter run --profile  # profile mode (release performance + debugging)
```

Usar DevTools para:
- **Widget rebuild tracker**: Encontrar reconstrucciones innecesarias
- **Timeline view**: Identificar jank (frames >16ms)
- **Memory tab**: Detectar fugas y alocaciones excesivas
- **CPU profiler**: Encontrar funciones con alto costo

### Performance Overlay

```dart
MaterialApp(
  showPerformanceOverlay: true,  // shows GPU/UI thread graphs
)
```

### Reglas

- **Medir antes de optimizar** — no adivinar
- Medir en **modo release** — el modo debug es 10-100x más lento
- Objetivo: **60fps** (16ms por frame) o **120fps** (8ms) en pantallas de alta frecuencia

---

## Cuellos de Botella Comunes

| Cuello de botella | Síntoma | Corrección |
|---|---|---|
| Reconstruir todo el árbol | UI con jank | `BlocSelector`, `Consumer`, `Selector` |
| Renderizado de listas grandes | OOM, scroll lento | `ListView.builder`, paginación |
| Imágenes sobredimensionadas | Carga lenta, memoria | Redimensionar en CDN, `CachedNetworkImage` |
| Cómputo pesado en build | Frames perdidos | Mover a isolate o BLoC |
| Animación sin RepaintBoundary | Repintado de widgets no relacionados | Envolver con `RepaintBoundary` |
| Métodos helper de widget | Sin aislamiento de reconstrucción | Extraer a clases widget separadas |
| `MediaQuery.of` en widgets anidados | Reconstrucciones excesivas | Pasar valores hacia abajo o usar `LayoutBuilder` |
| Constructores `const` faltantes | Creación innecesaria de objetos | Agregar `const` donde sea posible |

---

## Anti-Patrones

| Anti-patrón | Por qué es malo | Corrección |
|---|---|---|
| Métodos helper `Widget _buildX()` | Sin barrera de reconstrucción, se reconstruye con el padre | Extraer a clase `StatelessWidget` separada |
| `setState` en código adyacente al build | Dispara reconstrucción completa del widget | Usar `BlocSelector`/`Consumer` para actualizaciones granulares |
| Cargar imágenes en resolución completa | Presión de memoria, renderizado lento | Redimensionar en servidor, usar parámetros de CDN |
| Sin paginación para listas | Cargar miles de items | Paginar con `ListView.builder` + cargar más |
| Medir en modo debug | 10-100x más lento que en release | Siempre medir en `--profile` o `--release` |
