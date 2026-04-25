# Gate Post-Implementación

Después de CUALQUIER cambio de código en archivos `.go`, invocar el skill `/lint` antes de considerar la tarea como terminada. El skill de lint ejecuta tanto `golangci-lint` como `go test` y bloquea ante nuevas violaciones. Ver el skill de lint para más detalles.
