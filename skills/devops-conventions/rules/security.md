# Mejores Prácticas de Seguridad en Infraestructura

## Escaneo de Contenedores
- Escanea imágenes en CI antes del push (Trivy, Grype, o Snyk)
- Genera SBOMs con Syft; rastrea dependencias
- Aplica firma de imágenes (cosign) y pin de digest en producción
- Bloquea el deploy de imágenes con CVEs críticos/altos via admission controllers
- Habilita escaneo continuo para CVEs recién descubiertos

## Gestión de Secretos
- Elimina secretos de larga duración: usa identidad de carga de trabajo OIDC para auth CI/CD-a-cloud
- Usa almacenes nativos de nube: AWS Secrets Manager, GCP Secret Manager, HashiCorp Vault
- Automatiza la rotación; nunca almacenes secretos en archivos env, Git, o imágenes
- En K8s: external-secrets-operator para sincronizar secretos de nube a secretos K8s
- Audita los logs de acceso a secretos regularmente

## IAM con Menor Privilegio
- Empieza con cero permisos; agrega solo lo necesario
- Prefiere roles específicos de servicio sobre roles amplios
- Usa condiciones IAM (IP de origen, MFA, tiempo) para restricciones adicionales
- Revisiones de acceso regulares; elimina permisos no utilizados
- Usa credenciales de corta duración y asunción de roles sobre claves estáticas

## Seguridad de Red
- Default-deny ingress y egress por namespace de K8s
- Permite solo la comunicación pod-a-pod requerida via NetworkPolicies
- Usa service mesh (Istio/Linkerd) para mTLS entre servicios
- Bloquea el acceso a endpoints de metadatos de nube (169.254.169.254) a menos que sea necesario
- Restringe el egress solo a endpoints conocidos

## Runtime de Contenedor
- Ejecuta como non-root; elimina todas las capabilities, agrega de vuelta solo las necesarias
- Usa filesystem raíz de solo lectura donde sea posible
- Aplica perfiles seccomp (mínimo: `RuntimeDefault`)
- Nunca ejecutes contenedores privilegiados en producción
- Monitorea con Falco o detección de runtime similar

## Seguridad de CI/CD
- Fija todas las versiones de action/plugin a SHA
- Usa OIDC en lugar de credenciales de nube de larga duración
- Ejecuta SAST/SCA/escaneo de secretos en el pipeline (falla en críticos)
- Requiere revisiones de PR antes del merge a main
- Firma commits y artefactos
- Usa reglas de protección de entorno para deploys a producción
