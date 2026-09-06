# Guía de Arquitectura Flutter

## MVVM + Clean Architecture (Oficial de Google)

Google recomienda oficialmente MVVM con arquitectura por capas. Probado a escala por BMW (300 devs), Nubank (90M+ usuarios), ByteDance (700+ devs).

### Capas

```
┌─────────────────────────────────┐
│   UI Layer (Presentation)       │  Widgets + ViewModels/BLoCs
├─────────────────────────────────┤
│   Domain Layer (optional)       │  Use cases, entities, repo interfaces
├─────────────────────────────────┤
│   Data Layer                    │  Repositories + Services (API/DB)
└─────────────────────────────────┘
```

### Relaciones

- **Views ↔ ViewModels**: uno-a-uno por feature
- **ViewModels ↔ Repositories**: muchos-a-muchos
- **Services**: NO tienen estado — solo wrappers de carga de datos
- **Repositories**: nunca interactúan entre sí — combinar datos en ViewModels o capa de dominio
- Las importaciones son **direccionales hacia adentro** — UI → Domain → Data, nunca al revés

### Estructura de Carpetas — Feature-First

```
lib/
├── src/
│   ├── features/
│   │   ├── auth/
│   │   │   ├── presentation/     # widgets, pages, view_models
│   │   │   ├── application/      # use cases, BLoCs/Cubits
│   │   │   ├── domain/           # entities, repository interfaces
│   │   │   └── data/             # repo implementations, DTOs, services
│   │   ├── cart/
│   │   ├── products/
│   │   └── orders/
│   ├── common_widgets/           # shared UI components
│   ├── constants/
│   ├── exceptions/
│   ├── localization/
│   ├── routing/
│   └── utils/
├── main.dart
test/                             # mirrors lib/ structure
```

No todos los features necesitan todas las carpetas — incluir solo lo necesario.

---

## Manejo de Errores — Patrón Result (Oficial de Google)

Los repositorios retornan `Result<T>`, nunca lanzan excepciones. Los ViewModels/BLoCs hacen switch sobre Result.

```dart
sealed class Result<T> {
  const Result();
  const factory Result.ok(T value) = Ok._;
  const factory Result.error(Exception error) = Error._;
}

final class Ok<T> extends Result<T> {
  const Ok._(this.value);
  final T value;
}

final class Error<T> extends Result<T> {
  const Error._(this.error);
  final Exception error;
}
```

### Uso con Pattern Matching

```dart
final result = await userRepository.getProfile(id);
switch (result) {
  case Ok<UserProfile>():
    state = ProfileLoaded(result.value);
  case Error<UserProfile>():
    state = ProfileError(result.error.toString());
}
```

### Flujo de Errores por Capa

| Capa | Manejo de errores |
|---|---|
| **Service** | Puede lanzar (errores HTTP, parseo) |
| **Repository** | Captura excepciones del service, retorna `Result.error()` |
| **ViewModel/BLoC** | Hace switch sobre `Result`, nunca try/catch |
| **Widget** | Renderiza según el estado (loading/success/error) |

```dart
// repository
class UserRepositoryImpl implements UserRepository {
  final UserService _service;

  @override
  Future<Result<User>> getUser(String id) async {
    try {
      final dto = await _service.fetchUser(id);
      return Result.ok(dto.toDomain());
    } on HttpException catch (e) {
      return Result.error(e);
    } on FormatException catch (e) {
      return Result.error(e);
    }
  }
}

// viewmodel/bloc — no try/catch
Future<void> loadUser(String id) async {
  state = UserLoading();
  final result = await _repository.getUser(id);
  switch (result) {
    case Ok<User>():
      state = UserLoaded(result.value);
    case Error<User>():
      state = UserError(result.error.toString());
  }
}
```

---

## Stack de Generación de Código

| Paquete | Propósito |
|---------|---------|
| **freezed** | Clases de datos inmutables, copyWith, igualdad, sealed unions |
| **json_serializable** | Serialización/deserialización JSON |
| **injectable** | Generación de configuración de DI |
| **auto_route** | Generación de rutas (si no se usa GoRouter) |
| **build_runner** | Orquesta toda la generación de código |

### Entidad de Dominio con Freezed

```dart
@freezed
class User with _$User {
  const factory User({
    required String id,
    required String name,
    required String email,
    @Default(false) bool isVerified,
  }) = _User;
}
```

### DTO con json_serializable

```dart
@JsonSerializable()
class UserDto {
  final String id;
  final String name;
  final String email;
  @JsonKey(name: 'is_verified')
  final bool isVerified;

  const UserDto({
    required this.id,
    required this.name,
    required this.email,
    required this.isVerified,
  });

  factory UserDto.fromJson(Map<String, dynamic> json) => _$UserDtoFromJson(json);
  Map<String, dynamic> toJson() => _$UserDtoToJson(this);

  User toDomain() => User(id: id, name: name, email: email, isVerified: isVerified);
}
```

### Dos Capas de DTO

- **Entidades de dominio** (`freezed`): inmutables, sin anotaciones de serialización
- **DTOs** (`json_serializable`): serialización, mapper `toDomain()`
- Nunca mezclar — las entidades de dominio no saben de JSON

Ejecutar `dart run build_runner watch` durante el desarrollo.

---

## Inyección de Dependencias

### get_it + injectable (Estándar Enterprise)

```dart
@module
abstract class AppModule {
  @lazySingleton
  Dio get dio => Dio(BaseOptions(baseUrl: Env.apiUrl));

  @lazySingleton
  AuthRepository get authRepo => AuthRepositoryImpl(getIt<Dio>());
}

@injectable
class AuthBloc extends Bloc<AuthEvent, AuthState> {
  AuthBloc(AuthRepository repo) : super(AuthInitial());
}

// register all in main
void main() {
  configureDependencies();
  runApp(const MyApp());
}

// usage
final bloc = getIt<AuthBloc>();
```

### Registro por Entorno

```dart
@Environment('dev')
@LazySingleton(as: AuthRepository)
class MockAuthRepository implements AuthRepository { ... }

@Environment('prod')
@LazySingleton(as: AuthRepository)
class AuthRepositoryImpl implements AuthRepository { ... }
```

### DI con Riverpod (Alternativa)

```dart
@riverpod
AuthRepository authRepository(Ref ref) {
  return AuthRepositoryImpl(ref.read(dioProvider));
}
```

---

## Navegación — GoRouter

```dart
final router = GoRouter(
  routes: [
    GoRoute(path: '/', builder: (_, __) => const HomeScreen()),
    GoRoute(
      path: '/product/:id',
      builder: (_, state) {
        final id = state.pathParameters['id']!;
        return ProductScreen(id: id);
      },
    ),
    StatefulShellRoute.indexedStack(
      builder: (_, __, navigationShell) => MainShell(navigationShell: navigationShell),
      branches: [
        StatefulShellBranch(routes: [
          GoRoute(path: '/home', builder: (_, __) => const HomeTab()),
        ]),
        StatefulShellBranch(routes: [
          GoRoute(path: '/search', builder: (_, __) => const SearchTab()),
        ]),
        StatefulShellBranch(routes: [
          GoRoute(path: '/profile', builder: (_, __) => const ProfileTab()),
        ]),
      ],
    ),
  ],
);
```

### Reglas

- `StatefulShellRoute` para navegación inferior con stacks independientes (preserva estado por tab)
- Declarativo con sincronización de URL
- Deep linking out of the box
- Guards de redirección para auth: `redirect: (context, state) => isLoggedIn ? null : '/login'`

---

## Código Específico de Plataforma

### Casos Simples — MethodChannel

```dart
const platform = MethodChannel('com.example/native');

Future<String> getBatteryLevel() async {
  try {
    final result = await platform.invokeMethod<String>('getBatteryLevel');
    return result ?? 'Unknown';
  } on PlatformException catch (e) {
    return 'Error: ${e.message}';
  }
}
```

### Casos Complejos — Arquitectura de Plugin Federado

1. **Platform Interface Package**: interfaz abstracta
2. **App-Facing Package**: API para la app Flutter
3. **Platform Implementations**: iOS, Android, Web en paquetes separados

---

## Patrones por Empresa

| Empresa | Escala | Patrón | Lección clave |
|---------|-------|---------|------------|
| **BMW** | 300 devs, 47 países | MVVM basado en dominio | Equipos de dominio, no de plataforma. Patrón BFF desacopla features de releases de la app |
| **Nubank** | 90M+ usuarios | BLoC + Clean Arch | Separación estricta para cumplimiento financiero. PRs se mergean en promedio en 9.9 min |
| **Alibaba** | 100M+ usuarios | Fish Redux | Patrón Adapter para FPS en ListView (40→53). Prefetch de datos -300ms en dispositivos de gama baja |
| **ByteDance** | 700+ devs | Motor personalizado | Eliminar librerías no usadas para reducir tamaño. Aumento de productividad del 33% vs nativo |
| **Toyota** | Embebido | AOT + Embedder API | Flutter más allá del móvil — infotainment en autos |
