# React Native Review Checklist

Incluye todo lo de React (JS/TS) checklist, mas lo siguiente especifico de mobile:

## Correctness

- [ ] Navigation params tipados correctamente
- [ ] No hay listeners de navigation sin cleanup en useEffect
- [ ] Permisos (camera, location, notifications) se solicitan antes de usar
- [ ] Deep links manejan rutas invalidas sin crash
- [ ] Keyboard avoiding view en forms que lo necesitan
- [ ] Safe area insets respetados (notch, home indicator)
- [ ] Back handler en Android implementado donde aplique

## Security

- [ ] Datos sensibles usan SecureStore o Keychain, no AsyncStorage
- [ ] No hay API keys hardcodeadas en el bundle
- [ ] Certificate pinning en requests a APIs sensibles
- [ ] No se loguean datos sensibles en produccion
- [ ] Biometric auth no bypaseable desde el JS layer

## Conventions

- [ ] Estilos en StyleSheet.create, no objetos inline
- [ ] Dimensiones responsivas (no hardcodear px para diferentes screens)
- [ ] Platform-specific code usa Platform.select o archivos .ios/.android
- [ ] No hay logica de negocio en componentes de presentacion

## Performance

- [ ] FlatList/SectionList en lugar de ScrollView para listas largas
- [ ] FlatList tiene keyExtractor, getItemLayout si es posible
- [ ] No hay re-renders innecesarios en listas (React.memo en items)
- [ ] Imagenes cacheadas correctamente (FastImage o equivalente)
- [ ] Animaciones usan Animated o Reanimated en el UI thread
- [ ] No hay bridge calls excesivos en hot paths
- [ ] Hermes habilitado para mejor performance

## Platform-Specific

- [ ] Comportamiento probado en iOS Y Android
- [ ] Fonts custom registrados para ambas plataformas
- [ ] StatusBar configurado correctamente por pantalla
- [ ] Splash screen con transicion limpia
