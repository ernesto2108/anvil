# Flutter Testing Guide

## Testing Pyramid

| Layer | What to Test | Speed |
|---|---|---|
| **Unit Tests** | Business logic, ViewModels, Repositories, BLoCs | Fast |
| **Widget Tests** | Individual widget behavior and interaction | Medium |
| **Golden Tests** | Visual regression (pixel comparison) | Medium |
| **Integration Tests** | Full app flows, platform interaction | Slow |

---

## Unit Tests

Test business logic, BLoCs, Cubits, and repositories.

### BLoC Test

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

### Repository Test

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

### Rules

- Use `mocktail` for mocking (no code generation needed)
- Test Result pattern — both `Ok` and `Error` paths
- Group related tests with `group()`
- `setUp`/`tearDown` for consistent initialization

---

## Widget Tests

Test individual widget behavior and user interaction.

### Basic Widget Test

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

### Widget Test with BLoC

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

### Rules

- Always call `pumpAndSettle()` after interactions (tap, enter text)
- Use `find.byKey` for specific elements, `find.byType` for widget types
- Wrap in `MaterialApp` for theme/navigation context
- Mock BLoCs/providers — don't test state management in widget tests

---

## Golden Tests

Pixel-by-pixel comparison against baseline images. Catches visual regressions.

### Basic Golden Test

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

### Advanced Golden Tests with Alchemist

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

### Rules

- Run in controlled CI environment (deterministic fonts, locale, DPR)
- Update goldens with `flutter test --update-goldens`
- Commit golden files to version control
- Only golden-test visual components, not logic

---

## Integration Tests

Full app flows with platform interaction.

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

### Rules

- Test 3-5 critical user journeys (login, checkout, onboarding)
- Run on real devices/emulators
- Don't duplicate what widget tests cover
- Use `Key` widgets for stable element finding

---

## Test File Organization

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

### Shared Test Helpers

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

## Anti-Patterns

| Anti-Pattern | Fix |
|---|---|
| Testing implementation (state internals) | Test widget output and behavior |
| Missing `pumpAndSettle` after interactions | Always pump after tap/enterText |
| No mocking of external dependencies | Use `mocktail` for repos/services |
| Snapshot/golden tests for logic | Golden tests are visual only |
| Integration tests for everything | Unit → Widget → Golden → Integration (pyramid) |
| Flaky tests from timing issues | Use `pumpAndSettle`, not `pump` with arbitrary durations |
