# Flutter State Management Guide

## Decision Matrix

| Solution | Best For | Trade-off |
|----------|----------|-----------|
| **BLoC** | Enterprise, regulated industries, strict audit trails | More boilerplate, steeper curve |
| **Riverpod 3.0** | Most projects, fast iteration, compile-time safety | Less opinionated structure |
| **Provider** | Simple apps, small teams | Limited scalability |
| **GetX** | Rapid prototyping only | Poor testability, avoid in production |

### Escalation Path

```
setState (local) → Provider (simple shared) → Riverpod (most cases) → BLoC (enterprise/regulated)
```

### When to Use What

| Scope | Simple | Medium | Complex |
|---|---|---|---|
| Single Widget | `setState` | `setState` | Cubit |
| Feature (few widgets) | `ValueNotifier` | Cubit | BLoC |
| Cross-feature | Provider | Riverpod | BLoC |
| Global (app-wide) | Riverpod | Riverpod | BLoC |

---

## BLoC Pattern (Nubank — 90M+ users)

Event-driven architecture with strict separation. Chosen by Nubank for predictability, testability, and audit trails in financial applications.

### Events

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

### States

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

### Cubit (Simplified BLoC)

For simple state transitions without events:

```dart
class CounterCubit extends Cubit<int> {
  CounterCubit() : super(0);

  void increment() => emit(state + 1);
  void decrement() => emit(state - 1);
  void reset() => emit(0);
}
```

### Widget Integration

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

### BLoC Rules

- Events are sealed classes — one class per user action
- States are sealed classes — exhaustive switching in UI
- BLoC handles ONLY business logic — no UI code, no navigation
- Use `BlocListener` for side effects (navigation, toasts), `BlocBuilder` for UI
- One BLoC per feature. Share data via repositories, not BLoC-to-BLoC

---

## Riverpod Pattern

Compile-time safe, no BuildContext dependency, auto-disposal.

### Basic Provider

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

### Notifier (Stateful)

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

### Family (Parameterized)

```dart
@riverpod
Future<Product> product(Ref ref, String id) async {
  final repo = ref.read(productRepositoryProvider);
  return repo.getById(id);
}

// usage
final product = ref.watch(productProvider(productId));
```

### Riverpod Rules

- Use `ref.watch` in `build()` for reactive updates
- Use `ref.read` in callbacks/event handlers (non-reactive)
- Use `ref.listen` for side effects
- `autoDispose` is the default — providers clean up when no longer watched
- Prefer `@riverpod` annotation over manual provider creation

---

## Provider (Simple Cases Only)

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

### When Provider is Enough

- Theme switching
- Locale selection
- Simple feature flags
- Any single-value global state that changes infrequently

---

## Local State

### `setState` (Single Widget Only)

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

### `ValueNotifier` (Lightweight Shared)

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

### Rules

- `setState` only for state that no other widget needs
- If 2+ widgets need the same state → upgrade to Cubit, Riverpod, or Provider
- Never pass `setState` callbacks up the tree — that's prop drilling
