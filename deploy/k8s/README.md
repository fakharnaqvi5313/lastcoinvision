# Public testnet: 3-sentry + validator topology

Reference Kubernetes manifests for running one LCV validator behind 3 sentry nodes — the
architecture documented in
[the validator handbook](../../docs/docs/validator-handbook.md#sentry-node-architecture).
**These manifests are unexecuted reference infrastructure**: this environment has no live
cluster to test `kubectl apply` against (`kubectl` client is present, no context is
configured), so treat this as a correct starting point to adapt to your cluster/cloud
provider, not a push-button deploy. YAML structure was checked with a local parser, not a
running API server.

## Topology

```
                  ┌─────────────────────────┐
   public P2P --> │   lcv-sentry-{0,1,2}     │ <-- public RPC/REST (lcv-rpc Service)
                   │   (StatefulSet, 3 pods)  │
                  └────────────┬─────────────┘
                               │ private, in-cluster only
                               │ (NetworkPolicy-enforced)
                  ┌────────────▼─────────────┐
                  │      lcv-validator-0      │
                  │  (StatefulSet, 1 pod —    │
                  │   never scaled up: a      │
                  │   2nd replica sharing the │
                  │   consensus key IS a      │
                  │   double-sign)            │
                  └───────────────────────────┘
```

The validator's `persistent_peers` is limited to the 3 sentries and `pex = false` (no peer
discovery beyond that fixed list) — set by `configmap-entrypoint.yaml`'s startup script. A
`NetworkPolicy` (`networkpolicy-validator.yaml`) enforces the same restriction at the cluster
network layer, so even a config mistake can't expose the validator's P2P port publicly.
Sentries peer with the validator and with each other, and are the only pods that take public
P2P traffic (`lcv-sentry-p2p`, a `LoadBalancer` Service) or serve public RPC/REST
(`lcv-rpc`) — matching standard Cosmos operational practice.

## Files

| File | Purpose |
|---|---|
| `namespace.yaml` | The `lcv-testnet` namespace everything else lives in |
| `configmap-entrypoint.yaml` | Shared startup script for both roles — first-boot init, `persistent_peers` wiring |
| `configmap-genesis.example.yaml` | Template — fill from a real `genesis.json` |
| `configmap-node-ids.example.yaml` | Template — fill with real node IDs after generating keys |
| `secret-node-keys.example.yaml` | Template — **never commit real key material**; see the in-file instructions |
| `statefulset-validator.yaml` | The validator StatefulSet + its headless Service |
| `statefulset-sentry.yaml` | The 3-replica sentry StatefulSet, its headless Service, public P2P/RPC Services, and a PodDisruptionBudget |
| `networkpolicy-validator.yaml` | Restricts the validator's P2P port to sentry pods only |
| `kustomization.yaml` | Ties the above together for `kubectl apply -k .` |

## Before deploying

1. **Generate real keys offline** — never let cluster tooling generate a validator's
   consensus key. Follow the instructions at the top of `secret-node-keys.example.yaml`, then
   `kubectl apply` the *real* secrets you create from `kubectl create secret ... --from-file`
   (rename the `.example.yaml` files after filling them in, or better, never write the real
   values to a file the manifests directory at all — create the Secrets imperatively).
2. Run `lcvd comet show-node-id --home <dir>` for each of the 4 homes and fill in
   `configmap-node-ids.example.yaml` (rename to `configmap-node-ids.yaml`).
3. Get the real, published genesis.json for the network you're joining and fill in
   `configmap-genesis.example.yaml` (or better, `kubectl create configmap` directly from the
   file — see the comment in that template).
4. Push `chain/Dockerfile`'s image to a registry your cluster can pull from, and update the
   `image:` field in both StatefulSets (currently a placeholder, `lcv-lcvd:latest`).
5. Size `resources:`/`storage:` for your actual validator's expected load — the requests in
   these manifests are the hardware baseline from the validator handbook, not tuned for any
   specific cloud provider's instance types.

## Applying

```bash
kubectl apply -k deploy/k8s/
```

## Not covered here

- **Upgrades**: see `../../scripts/upgrade/` for the cosmovisor + governance walkthrough —
  StatefulSet pods run `cosmovisor` under the hood once wired in, so a chain upgrade doesn't
  require touching these manifests at all.
- **Monitoring/alerting**: no Prometheus/Grafana manifests here — hook into the CometBFT RPC
  and lcvd's `/metrics` (if you enable `[telemetry]` in `app.toml`) with whatever monitoring
  stack your cluster already runs.
- **Non-validating full nodes / archive nodes**: this topology is validator-focused; a public
  RPC-only fleet would reuse the sentry StatefulSet's pattern without a validator behind it.
