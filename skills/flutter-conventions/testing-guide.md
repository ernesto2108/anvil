# Guía de Testing Flutter

## Pirámide de Testing

| Capa | Qué testear | Velocidad |
|---|---|---|
| **Unit Tests** | Lógica de negocio, ViewModels, Repositories, BLoCs | Rápido |
| **Widget Tests** | Comportamiento e interacción de widgets individuales | Medio |
| **Golden Tests** | Regresión visual (comparación de píxeles) | Medio |
| **Integration Tests** | Flujos completos de la app, interacción con plataforma | Lento |

---

## Unit Tests

Testear lógica de negocio, BLoCs, Cubits y repositorios.

### Test de BLoC

```dart
import 'package:bloc_test/bloc_test.dart';
import 'package:test/test.dart';

void main() {
  late AuthBloc authBloc;
  late MockAuthRepository mockRepo;

  setUp(() {
    mockRepo = MockAuthRepository();
    authBloc = AuthBloc(mockRepo);
  });

  tearDown(() => authBloc.close());

  group('AuthBloc', () => {
    blocTest<AuthBloc, AuthState>(
      'emits [loading, success] when login succeeds',
      build: () {
        when(() => mockRepo.login(any(), any()))
            .thenAnswer((_) async => Result.ok(testUser));
        return authBloc;
      },
      act: (bloc) => bloc.add(LoginRequested(email: 'a@b.com', password: '123')),
      expect: () => [isA<AuthLoading>(), isA<AuthSuccess>()],
    );

    blocTest<AuthBloc, AuthState>(
      'emits [loading, failure] when login fails',
      build: () {
        when(() => mockRepo.login(any(), any()))
            .thenAnswer((_) async => Result.error(Exception('Invalid')));
        return authBloc;
      },
      act: (bloc) => bloc.add(LoginRequested(email: 'a@b.com', password: 'wrong')),
      expect: () => [isA<AuthLoading>(), isA<AuthFailure>()],
    );
  });
}
```

### Test de Repositorio

```dart
void main() {
  late UserRepositoryImpl repo;
  late MockUserService mockService;

  setUp(() {
    mockService = MockUserService();
    repo = UserRepositoryImpl(mockService);
  });

  test('returns Ok with user when service succeeds', () async {
    when(() => mockService.fetchUser('1'))
        .thenAnswer((_) async => UserDto(id: '1', name: 'John', email: 'j@t.com', isVerified: true));

    final result = await repo.getUser('1');

    expect(result, isA<Ok<User>>());
    expect((result as Ok<User>).value.name, equals('John'));
  });

  test('returns Error when service throws', () async {
    when(() => mockService.fetchUser('1'))
        .thenThrow(HttpException('Not found'));

    final result = await repo.getUser('1');

    expect(result, isA<Error<User>>());
  });
}
```

### Reglas

- Usar `mocktail` para mocking (sin necesidad de generación de código)
- Testear el patrón Result — tanto rutas `Ok` como `Error`
- Agrupar tests relacionados con `group()`
- `setUp`/`tearDown` para inicialización consistente

---

## Widget Tests

Testear el comportamiento individual de widgets e interacciones del usuario.

### Test Básico de Widget

```dart
void main() {
  testWidgets('shows error on empty submit', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: LoginScreen()));

    await tester.tap(find.byType(ElevatedButton));
    await tester.pumpAndSettle();

    expect(find.text('Email is required'), findsOneWidget);
  });

  testWidgets('calls onSubmit with valid data', (tester) async {
    final onSubmit = MockCallback<LoginData>();

    await tester.pumpWidget(MaterialApp(
      home: LoginScreen(onSubmit: onSubmit),
    ));

    await tester.enterText(find.byKey(const Key('email')), 'test@test.com');
    await tester.enterText(find.byKey(const Key('password')), 'password123');
    await tester.tap(find.byType(ElevatedButton));
    await tester.pumpAndSettle();

    verify(() => onSubmit(LoginData(email: 'test@test.com', password: 'password123'))).called(1);
  });
}
```

### Test de Widget con BLoC

```dart
testWidgets('shows user name when loaded', (tester) async {
  final bloc = MockAuthBloc();
  whenListen(bloc, Stream.value(AuthSuccess(testUser)), initialState: AuthInitial());

  await tester.pumpWidget(
    MaterialApp(
      home: BlocProvider<AuthBloc>.value(
        value: bloc,
        child: const ProfileScreen(),
      ),
    ),
  );

  await tester.pumpAndSettle();
  expect(find.text('John Doe'), findsOneWidget);
});
```

### Reglas

- Siempre llamar `pumpAndSettle()` después de interacciones (tap, enterText)
- Usar `find.byKey` para elementos específicos, `find.byType` para tipos de widget
- Envolver en `MaterialApp` para contexto de tema/navegación
- Mockear BLoCs/providers — no testear gestión de estado en tests de widget

---

## Golden Tests

Comparación píxel a píxel contra imágenes de referencia. Detecta regresiones visuales.

### Test Golden Básico

```dart
testWidgets('UserCard matches golden', (tester) async {
  await tester.pumpWidget(
    MaterialApp(
      home: Scaffold(
        body: UserCard(user: User(id: '1', name: 'John', email: 'j@t.com')),
      ),
    ),
  );

  await expectLater(
    find.byType(UserCard),
    matchesGoldenFile('goldens/user_card.png'),
  );
});
```

### Golden Tests Avanzados con Alchemist

```dart
void main() {
  goldenTest(
    'UserCard variants',
    fileName: 'user_card_variants',
    builder: () => GoldenTestGroup(
      children: [
        GoldenTestScenario(
          name: 'default',
          child: UserCard(user: testUser),
        ),
        GoldenTestScenario(
          name: 'verified',
          child: UserCard(user: testUser.copyWith(isVerified: true)),
        ),
        GoldenTestScenario(
          name: 'long name',
          child: UserCard(user: testUser.copyWith(name: 'Very Long Name That Might Overflow')),
        ),
      ],
    ),
  );
}
```

### Reglas

- Ejecutar en entorno CI controlado (fuentes, locale y DPR deterministas)
- Actualizar goldens con `flutter test --update-goldens`
- Commitear los archivos golden al control de versiones
- Solo golden-testear componentes visuales, no lógica

---

## Integration Tests

Flujos completos de la app con interacción de plataforma.

```dart
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('login flow end-to-end', (tester) async {
    app.main();
    await tester.pumpAndSettle();

    // login
    await tester.enterText(find.byKey(const Key('email')), 'test@test.com');
    await tester.enterText(find.byKey(const Key('password')), 'password');
    await tester.tap(find.text('Login'));
    await tester.pumpAndSettle();

    // verify navigation to home
    expect(find.text('Welcome'), findsOneWidget);

    // navigate to profile
    await tester.tap(find.byIcon(Icons.person));
    await tester.pumpAndSettle();

    expect(find.text('test@test.com'), findsOneWidget);
  });
}
```

### Reglas

- Testear 3-5 journeys críticos de usuario (login, checkout, onboarding)
- Ejecutar en dispositivos/emuladores reales
- No duplicar lo que cubren los widget tests
- Usar widgets `Key` para encontrar elementos de forma estable

---

## Organización de Archivos de Test

```
test/
├── features/
│   ├── auth/
│   │   ├── application/
│   │   │   └── auth_bloc_test.dart
│   │   ├── data/
│   │   │   └── auth_repository_test.dart
│   │   └── presentation/
│   │       └── login_screen_test.dart
│   └── cart/
│       └── ...
├── goldens/                    # golden image files
├── fixtures/                   # test data
│   ├── user_fixture.dart
│   └── json/
│       └── user_response.json
└── helpers/
    ├── pump_app.dart           # custom pumpWidget with providers
    └── mocks.dart              # shared mocks
integration_test/
└── app_test.dart
```

### Helpers Compartidos de Test

```dart
// test/helpers/pump_app.dart
extension PumpApp on WidgetTester {
  Future<void> pumpApp(Widget widget) {
    return pumpWidget(
      MaterialApp(
        home: Scaffold(body: widget),
      ),
    );
  }
}

// test/fixtures/user_fixture.dart
final testUser = User(id: '1', name: 'John Doe', email: 'john@test.com');
final testUsers = [testUser, User(id: '2', name: 'Jane', email: 'jane@test.com')];
```

---

## Anti-Patrones

| Anti-patrón | Corrección |
|---|---|
| Testear la implementación (internos del estado) | Testear la salida y comportamiento del widget |
| Falta `pumpAndSettle` después de interacciones | Siempre hacer pump después de tap/enterText |
| Sin mockear dependencias externas | Usar `mocktail` para repos/services |
| Tests de snapshot/golden para lógica | Los golden tests son solo visuales |
| Integration tests para todo | Unit → Widget → Golden → Integration (pirámide) |
| Tests inestables por problemas de timing | Usar `pumpAndSettle`, no `pump` con duraciones arbitrarias |
