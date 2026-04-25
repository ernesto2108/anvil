# Matriz de Decisión de Concurrencia

Cuándo usar cada primitiva de concurrencia:

| Escenario | Usar | Por qué |
|---|---|---|
| Proteger estado compartido (caché, contador, config) | `sync.Mutex` / `sync.RWMutex` | Lo más simple y rápido para resguardar datos |
| Coordinar goroutines, pasar datos entre etapas | Channels | Natural para productor-consumidor, pipelines |
| Tareas paralelas que pueden fallar | `errgroup.Group` | Combina WaitGroup + propagación de errores + cancelación de contexto |
| Esperar a que N goroutines terminen (sin errores) | `sync.WaitGroup` | Ligero, no requiere manejo de errores |
| Contador o flag atómico | `sync/atomic` | Sin locks, el más rápido para valores únicos |
| Inicialización única | `sync.Once` | Init lazy thread-safe |
| Acceso concurrente a mapas | `sync.Map` o `map` + `sync.RWMutex` | `sync.Map` para lectura intensiva con claves estables; `map` + mutex para todo lo demás |
| Rate limiting | `time.Ticker` + channel o `golang.org/x/time/rate` | Token bucket para límites de API |
| Propagación de timeout / cancelación | `context.Context` | Siempre -- es el mecanismo estándar de cancelación |

## Flujo de Decisión Rápido

```
¿Necesitas compartir datos entre goroutines?
  SÍ -> ¿Las goroutines transfieren ownership de los datos?
           SÍ -> Channel
           NO  -> ¿Es lectura intensiva, escritura poco frecuente?
                    SÍ -> sync.RWMutex
                    NO  -> sync.Mutex
  NO  -> ¿Las goroutines realizan trabajo paralelo?
           SÍ -> ¿Pueden fallar?
                    SÍ -> errgroup
                    NO  -> sync.WaitGroup
           NO  -> ¿Necesitas timeout/cancelación?
                    SÍ -> context.WithTimeout / context.WithCancel
```

Fuente: [Go Wiki: Mutex or Channel](https://go.dev/wiki/MutexOrChannel) -- "Usa el que sea más expresivo y simple para tu problema."
