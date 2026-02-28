# Changelog

## [1.3.0](https://github.com/migalsp/kubex-operator/compare/v1.2.0...v1.3.0) (2026-02-28)


### Features

* Introduce NamespaceOptimization CRD and enhance ScalingGroup with staged scaling sequences and new status fields, alongside corresponding UI and controller updates. ([56d4f03](https://github.com/migalsp/kubex-operator/commit/56d4f03dc6d4df921667cea6ba1237bdd5c6399c))

## [1.2.0](https://github.com/migalsp/kubex-operator/compare/v1.1.0...v1.2.0) (2026-02-27)


### Features

* Implement optimistic UI updates, loading states, and fast polling for manual scaling actions across scaling pages. ([d1f5fbe](https://github.com/migalsp/kubex-operator/commit/d1f5fbeb767b6ebf4ed5f1381b13df94a85f9b50))
* implement silent background data refreshing for namespace details and add an auto-scroll toggle for operator logs. ([0d6d2d4](https://github.com/migalsp/kubex-operator/commit/0d6d2d41077d220c5e5d7a7bdcdbc48ac9e53454))
* Lower CPU safety floor from 100m to 20m, refactor resource floor application logic, and ensure limits are always greater than or equal to requests. ([7c3d441](https://github.com/migalsp/kubex-operator/commit/7c3d44134456fc3bd3243fffb8dfbcca633595d4))
* sanitize and validate scaling group names for Kubernetes RFC 1123 compliance. ([a0646c8](https://github.com/migalsp/kubex-operator/commit/a0646c84ab6b45c2612938c837af50e1c496b9f8))

## [1.1.0](https://github.com/migalsp/kubex-operator/compare/v1.0.1...v1.1.0) (2026-02-27)


### Features

* Display requested resource allocation alongside actual usage for cluster and node metrics on the dashboard. ([2d05a82](https://github.com/migalsp/kubex-operator/commit/2d05a82ca66418fc906b54f5402af390c89b144a))


### Bug Fixes

* init release issues ([cc1ea33](https://github.com/migalsp/kubex-operator/commit/cc1ea339bec416a40eb0271620deefeb37e888ae))

## [1.0.1](https://github.com/migalsp/kubex-operator/compare/v1.0.0...v1.0.1) (2026-02-27)


### Bug Fixes

* split docker and helm registry paths ([27bda6a](https://github.com/migalsp/kubex-operator/commit/27bda6a62788ae42c3b5edc70ea02e341ca7913c))

## 1.0.0 (2026-02-27)


### Features

* Implement initial Kubex Kubernetes operator and UI for FinOps resource optimization. ([661661a](https://github.com/migalsp/kubex-operator/commit/661661a8dfea8614d5d973fa98fe365747048a76))
* Update Go version to 1.25 and dynamically report Go and Node versions in CI. ([27d976a](https://github.com/migalsp/kubex-operator/commit/27d976a179279bed44a93a7b703ab9d8ea273e02))
