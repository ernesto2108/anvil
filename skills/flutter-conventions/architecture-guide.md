# Flutter Architecture Guide

## MVVM + Clean Architecture (Google Official)

Google officially recommends MVVM with layered architecture. Proven at scale by BMW (300 devs), Nubank (90M+ users), ByteDance (700+ devs).

### Layers

```
┌─────────────────────────────────┐
│   UI Layer (Presentation)       │  Widgets + ViewModels/BLoCs
├─────────────────────────────────┤
│   Domain Layer (optional)       │  Use cases, entities, repo interfaces
├─────────────────────────────────┤
│   Data Layer                    │  Repositories + Services (API/DB)
└─────────────────────────────────┘
```

### Relationships

- **Views ↔ ViewModels**: one-to-one per feature
- **ViewModels ↔ Repositories**: many-to-many
- **Services**: hold NO state — pure data-loading wrappers
- **Repositories**: never interact with each other — combine data in ViewModels or domain layer
- Imports are **directional inward** — UI → Domain → Data, never reverse

### Folder Structure — Feature-First

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

Not every feature needs all folders — only include what's necessary.

---

## Error Handling — Result Pattern (Google Official)

Repositories return `Result<T>`, never throw. ViewModels/BLoCs switch on Result.

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

### Usage with Pattern Matching

```dart
final result = await userRepository.getProfile(id);
switch (result) {
  case Ok<UserProfile>():
    state = ProfileLoaded(result.value);
  case Error<UserProfile>():
    state = ProfileError(result.error.toString());
}
```

### Error Flow by Layer

| Layer | Error Handling |
|---|---|
| **Service** | May throw (HTTP errors, parsing) |
| **Repository** | Catches service exceptions, returns `Result.error()` |
| **ViewModel/BLoC** | Switches on `Result`, never try/catch |
| **Widget** | Renders based on state (loading/success/error) |

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

## Code Generation Stack

| Package | Purpose |
|---------|---------|
| **freezed** | Immutable data classes, copyWith, equality, sealed unions |
| **json_serializable** | JSON serialization/deserialization |
| **injectable** | DI configuration generation |
| **auto_route** | Route generation (if not using GoRouter) |
| **build_runner** | Orchestrates all code generation |

### Domain Entity with Freezed

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

### DTO with json_serializable

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

### Two DTO Layers

- **Domain entities** (`freezed`): immutable, no serialization annotations
- **DTOs** (`json_serializable`): serialization, `toDomain()` mapper
- Never mix — domain entities don't know about JSON

Run `dart run build_runner watch` during development.

---

## Dependency Injection

### get_it + injectable (Enterprise Standard)

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

### Environment-Specific Registration

```dart
@Environment('dev')
@LazySingleton(as: AuthRepository)
class MockAuthRepository implements AuthRepository { ... }

@Environment('prod')
@LazySingleton(as: AuthRepository)
class AuthRepositoryImpl implements AuthRepository { ... }
```

### Riverpod DI (Alternative)

```dart
@riverpod
AuthRepository authRepository(Ref ref) {
  return AuthRepositoryImpl(ref.read(dioProvider));
}
```

---

## Navigation — GoRouter

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

### Rules

- `StatefulShellRoute` for bottom nav with independent stacks (preserves state per tab)
- Declarative with URL synchronization
- Deep linking out of the box
- Redirect guards for auth: `redirect: (context, state) => isLoggedIn ? null : '/login'`

---

## Platform-Specific Code

### Simple Cases — MethodChannel

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

### Complex Cases — Federated Plugin Architecture

1. **Platform Interface Package**: abstract interface
2. **App-Facing Package**: API for Flutter app
3. **Platform Implementations**: iOS, Android, Web in separate packages

---

## Patterns by Company

| Company | Scale | Pattern | Key Lesson |
|---------|-------|---------|------------|
| **BMW** | 300 devs, 47 countries | Domain-based MVVM | Domain teams, not platform teams. BFF pattern decouples features from app releases |
| **Nubank** | 90M+ users | BLoC + Clean Arch | Strict separation for financial compliance. PRs merge in 9.9 min avg |
| **Alibaba** | 100M+ users | Fish Redux | Adapter pattern for ListView FPS (40→53). Data prefetch -300ms on low-end |
| **ByteDance** | 700+ devs | Custom engine | Strip unused libraries for size. 33% productivity increase over native |
| **Toyota** | Embedded | AOT + Embedder API | Flutter beyond mobile — in-car infotainment |
