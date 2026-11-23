# Guía de Gestión de Estado Flutter

## Matriz de Decisión

| Solución | Mejor para | Trade-off |
|----------|----------|-----------|
| **BLoC** | Enterprise, industrias reguladas, auditorías estrictas | Más boilerplate, curva más pronunciada |
| **Riverpod 3.0** | La mayoría de proyectos, iteración rápida, seguridad en tiempo de compilación | Estructura menos opinada |
| **Provider** | Apps simples, equipos pequeños | Escalabilidad limitada |
| **GetX** | Solo prototipado rápido | Poca testabilidad, evitar en producción |

### Ruta de Escalado

```
setState (local) → Provider (simple shared) → Riverpod (most cases) → BLoC (enterprise/regulated)
```

### Cuándo Usar Qué

| Alcance | Simple | Medio | Complejo |
|---|---|---|---|
| Widget único | `setState` | `setState` | Cubit |
| Feature (pocos widgets) | `ValueNotifier` | Cubit | BLoC |
| Cross-feature | Provider | Riverpod | BLoC |
| Global (toda la app) | Riverpod | Riverpod | BLoC |

---

## Patrón BLoC (Nubank — 90M+ usuarios)

Arquitectura orientada a eventos con separación estricta. Elegido por Nubank por su predictibilidad, testabilidad y trazas de auditoría en aplicaciones financieras.

### Eventos

```dart
sealed class AuthEvent {}

class LoginRequested extends AuthEvent {
  final String email;
  final String password;
  LoginRequested({required this.email, required this.password});
}

class LogoutRequested extends AuthEvent {}

class TokenRefreshRequested extends AuthEvent {}
```

### Estados

```dart
sealed class AuthState {}

class AuthInitial extends AuthState {}

class AuthLoading extends AuthState {}

class AuthSuccess extends AuthState {
  final User user;
  AuthSuccess(this.user);
}

class AuthFailure extends AuthState {
  final String message;
  AuthFailure(this.message);
}
```

### BLoC

```dart
class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final AuthRepository _repo;

  AuthBloc(this._repo) : super(AuthInitial()) {
    on<LoginRequested>(_onLoginRequested);
    on<LogoutRequested>(_onLogoutRequested);
    on<TokenRefreshRequested>(_onTokenRefresh);
  }

  Future<void> _onLoginRequested(LoginRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    final result = await _repo.login(event.email, event.password);
    switch (result) {
      case Ok<User>():
        emit(AuthSuccess(result.value));
      case Error<User>():
        emit(AuthFailure(result.error.toString()));
    }
  }

  Future<void> _onLogoutRequested(LogoutRequested event, Emitter<AuthState> emit) async {
    await _repo.logout();
    emit(AuthInitial());
  }

  Future<void> _onTokenRefresh(TokenRefreshRequested event, Emitter<AuthState> emit) async {
    final result = await _repo.refreshToken();
    switch (result) {
      case Ok<User>():
        emit(AuthSuccess(result.value));
      case Error<User>():
        emit(AuthInitial()); // force re-login
    }
  }
}
```

### Cubit (BLoC Simplificado)

Para transiciones de estado simples sin eventos:

```dart
class CounterCubit extends Cubit<int> {
  CounterCubit() : super(0);

  void increment() => emit(state + 1);
  void decrement() => emit(state - 1);
  void reset() => emit(0);
}
```

### Integración con Widgets

```dart
// provide
BlocProvider(
  create: (context) => getIt<AuthBloc>(),
  child: const LoginScreen(),
)

// consume with BlocBuilder
BlocBuilder<AuthBloc, AuthState>(
  builder: (context, state) {
    return switch (state) {
      AuthInitial() => const LoginForm(),
      AuthLoading() => const CircularProgressIndicator(),
      AuthSuccess(:final user) => ProfileScreen(user: user),
      AuthFailure(:final message) => ErrorWidget(message: message),
    };
  },
)

// listen for side effects (navigation, snackbars)
BlocListener<AuthBloc, AuthState>(
  listener: (context, state) {
    if (state is AuthSuccess) {
      context.go('/home');
    }
    if (state is AuthFailure) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(state.message)),
      );
    }
  },
  child: const LoginForm(),
)
```

### Reglas de BLoC

- Los eventos son clases sealed — una clase por acción del usuario
- Los estados son clases sealed — switching exhaustivo en la UI
- El BLoC maneja SOLO lógica de negocio — sin código UI, sin navegación
- Usar `BlocListener` para efectos secundarios (navegación, toasts), `BlocBuilder` para UI
- Un BLoC por feature. Compartir datos vía repositorios, no de BLoC a BLoC

---

## Patrón Riverpod

Seguro en tiempo de compilación, sin dependencia de BuildContext, auto-disposal.

### Provider Básico

```dart
@riverpod
Future<List<Product>> products(Ref ref) async {
  final repo = ref.read(productRepositoryProvider);
  return repo.getAll();
}

// widget
class ProductList extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = ref.watch(productsProvider);

    return products.when(
      data: (items) => ListView.builder(
        itemCount: items.length,
        itemBuilder: (_, index) => ProductCard(product: items[index]),
      ),
      loading: () => const CircularProgressIndicator(),
      error: (err, stack) => ErrorWidget(message: err.toString()),
    );
  }
}
```

### Notifier (Con Estado)

```dart
@riverpod
class CartNotifier extends _$CartNotifier {
  @override
  List<CartItem> build() => [];

  void addItem(CartItem item) {
    state = [...state, item];
  }

  void removeItem(String id) {
    state = state.where((item) => item.id != id).toList();
  }

  double get total => state.fold(0, (sum, item) => sum + item.price * item.quantity);
}
```

### Family (Parametrizado)

```dart
@riverpod
Future<Product> product(Ref ref, String id) async {
  final repo = ref.read(productRepositoryProvider);
  return repo.getById(id);
}

// usage
final product = ref.watch(productProvider(productId));
```

### Reglas de Riverpod

- Usar `ref.watch` en `build()` para actualizaciones reactivas
- Usar `ref.read` en callbacks/event handlers (no reactivo)
- Usar `ref.listen` para efectos secundarios
- `autoDispose` es el valor por defecto — los providers se limpian cuando dejan de observarse
- Preferir la anotación `@riverpod` sobre la creación manual de providers

---

## Provider (Solo Casos Simples)

```dart
class ThemeNotifier extends ChangeNotifier {
  ThemeMode _mode = ThemeMode.light;
  ThemeMode get mode => _mode;

  void toggle() {
    _mode = _mode == ThemeMode.light ? ThemeMode.dark : ThemeMode.light;
    notifyListeners();
  }
}

// provide
ChangeNotifierProvider(create: (_) => ThemeNotifier())

// consume
Consumer<ThemeNotifier>(
  builder: (_, theme, __) => Switch(
    value: theme.mode == ThemeMode.dark,
    onChanged: (_) => theme.toggle(),
  ),
)
```

### Cuándo Provider es Suficiente

- Cambio de tema
- Selección de idioma
- Feature flags simples
- Cualquier estado global de valor único que cambia con poca frecuencia

---

## Estado Local

### `setState` (Solo un Widget)

```dart
class _CounterState extends State<Counter> {
  int _count = 0;

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      onPressed: () => setState(() => _count++),
      child: Text('Count: $_count'),
    );
  }
}
```

### `ValueNotifier` (Compartido Ligero)

```dart
class CartBadge extends StatelessWidget {
  final ValueNotifier<int> itemCount;
  const CartBadge({required this.itemCount});

  @override
  Widget build(BuildContext context) {
    return ValueListenableBuilder<int>(
      valueListenable: itemCount,
      builder: (_, count, __) => Badge(
        label: Text('$count'),
        child: const Icon(Icons.shopping_cart),
      ),
    );
  }
}
```

### Reglas

- `setState` solo para estado que ningún otro widget necesita
- Si 2+ widgets necesitan el mismo estado → actualizar a Cubit, Riverpod o Provider
- Nunca pasar callbacks `setState` hacia arriba en el árbol — eso es prop drilling
