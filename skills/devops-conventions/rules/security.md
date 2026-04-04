# Infrastructure Security Best Practices

## Container Scanning
- Scan images in CI before push (Trivy, Grype, or Snyk)
- Generate SBOMs with Syft; track dependencies
- Enforce image signing (cosign) and digest pinning in production
- Block deployment of images with critical/high CVEs via admission controllers
- Enable continuous scanning for newly discovered CVEs

## Secrets Management
- Eliminate long-lived secrets: use OIDC workload identity for CI/CD-to-cloud auth
- Use cloud-native stores: AWS Secrets Manager, GCP Secret Manager, HashiCorp Vault
- Automate rotation; never store secrets in env files, Git, or images
- In K8s: external-secrets-operator to sync cloud secrets to K8s secrets
- Audit secret access logs regularly

## IAM Least Privilege
- Start with zero permissions; add only what's needed
- Prefer service-specific roles over broad roles
- Use IAM conditions (source IP, MFA, time) for additional restrictions
- Regular access reviews; remove unused permissions
- Use short-lived credentials and role assumption over static keys

## Network Security
- Default-deny ingress and egress per K8s namespace
- Allow only required pod-to-pod communication via NetworkPolicies
- Use service mesh (Istio/Linkerd) for mTLS between services
- Block access to cloud metadata endpoints (169.254.169.254) unless needed
- Restrict egress to known endpoints only

## Container Runtime
- Run as non-root; drop all capabilities, add back only what's needed
- Use read-only root filesystem where possible
- Apply seccomp profiles (minimum: `RuntimeDefault`)
- Never run privileged containers in production
- Monitor with Falco or similar runtime detection

## CI/CD Security
- Pin all action/plugin versions to SHA
- Use OIDC instead of long-lived cloud credentials
- Run SAST/SCA/secret scanning in pipeline (fail on critical)
- Require PR reviews before merge to main
- Sign commits and artifacts
- Use environment protection rules for production deploys
