# Flutter Performance Guide

## Performance Checklist (Priority Order)

From Alibaba (100M+ users), ByteDance (700+ devs), BMW (300 devs).

### 1. `const` Widgets — Up to 30% Rendering Improvement

```dart
// bad: recreated every build
child: Text('Hello')

// good: compile-time constant, never rebuilt
child: const Text('Hello')
```

**Rules:**
- Mark every stateless widget constructor as `const`
- Use `const` for any widget with no dynamic data
- Dart analyzer warns on missing `const` — fix all warnings

### 2. `ListView.builder` — 50%+ Memory Reduction

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

Also use `ListView.separated` for lists with dividers.

### 3. `RepaintBoundary` — Isolate Expensive Repaints

```dart
// wrap around animations or frequently-updating UI
RepaintBoundary(
  child: AnimatedWidget(...),
)
```

Use when: animations, timers, progress indicators, or any widget that updates independently of its parent.

### 4. Keep `build()` Pure

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

### 5. Granular Rebuilds

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

### 6. Extract Widgets, Don't Use Helper Methods

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

### 7. Image Optimization (Alibaba: -300ms on Low-End Android)

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

## Alibaba Patterns (100M+ Users)

- **Data prefetching**: Start loading data before navigating to the next screen
- **Template preloading**: Pre-build widget templates for common screens
- **Adapter pattern for ListView**: Custom adapter instead of component model. FPS improvement: 40 → 53 on Android
- **Native image rendering**: Direct TextureID rendering, bypassing PixelBuffer copy

---

## ByteDance Patterns (700+ Devs)

- **Strip unused native libraries**: Removed unused parts of Skia, BoringSSL, ICU, libwebp from Flutter engine
- **iOS data section compression**: Reduced app package size
- **Custom rendering pipeline**: Optimized for their specific use cases
- **Result: ~33% productivity increase** over native development

---

## Profiling

### Flutter DevTools

```bash
flutter run --profile  # profile mode (release performance + debugging)
```

Use DevTools to:
- **Widget rebuild tracker**: Find unnecessary rebuilds
- **Timeline view**: Identify jank (frames >16ms)
- **Memory tab**: Detect leaks and excessive allocations
- **CPU profiler**: Find hot functions

### Performance Overlay

```dart
MaterialApp(
  showPerformanceOverlay: true,  // shows GPU/UI thread graphs
)
```

### Rules

- **Profile before optimizing** — don't guess
- Profile in **release mode** — debug mode is 10-100x slower
- Target **60fps** (16ms per frame) or **120fps** (8ms) on high-refresh displays

---

## Common Bottlenecks

| Bottleneck | Symptom | Fix |
|---|---|---|
| Rebuilding entire tree | Janky UI | `BlocSelector`, `Consumer`, `Selector` |
| Large list rendering | OOM, slow scroll | `ListView.builder`, pagination |
| Oversized images | Slow loading, memory | CDN resize, `CachedNetworkImage` |
| Heavy computation in build | Dropped frames | Move to isolate or BLoC |
| Animation without RepaintBoundary | Repainting unrelated widgets | Wrap with `RepaintBoundary` |
| Widget helper methods | No rebuild isolation | Extract to separate widget classes |
| `MediaQuery.of` in nested widgets | Excessive rebuilds | Pass values down or use `LayoutBuilder` |
| Missing `const` constructors | Unnecessary object creation | Add `const` everywhere possible |

---

## Anti-Patterns

| Anti-Pattern | Why It's Bad | Fix |
|---|---|---|
| `Widget _buildX()` helper methods | No rebuild boundary, rebuilds with parent | Extract to separate `StatelessWidget` class |
| `setState` in build-adjacent code | Triggers full widget rebuild | Use `BlocSelector`/`Consumer` for granular updates |
| Loading full-resolution images | Memory pressure, slow render | Resize server-side, use CDN params |
| No pagination for lists | Loading thousands of items | Paginate with `ListView.builder` + load more |
| Profiling in debug mode | 10-100x slower than release | Always profile in `--profile` or `--release` |
