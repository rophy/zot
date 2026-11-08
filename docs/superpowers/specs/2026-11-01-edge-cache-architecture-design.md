# zot as an On-Premises Edge Cache: Architecture

| | |
|---|---|
| Date | 2026-11-01 |
| Status | Design agreed; not deployed or tested |
| Scope | Corporate, on-premises: one origin registry, up to about 10 sites |
| zot version | Current `main`: v2.1.23 plus PR #4646 (merged 2026-11-08, not yet released) |

## 1. Requirements and decisions

| Topic | Decision |
|---|---|
| Origin | One registry, for example a Harbor cluster. |
| Sites | Up to about 10. Each site has a usable WAN link to the origin, so every site syncs directly from the origin (one tier, no regional hubs). |
| Edge per site | Two zot nodes, each with a local disk. On Kubernetes when the site has it; otherwise two VMs with HA through DNS. |
| Image references | Clients pull pinned versions or digests only. A tag never changes after it is pushed. |
| Access control | A configuration choice per deployment (see section 6). It does not change the architecture. |
| Site failure | No failover to other sites. New pulls at a failed site fail until it is repaired; running containers are not affected. |
| Cache fill | On demand. Prewarming by an external trigger is a later add-on (section 7). |

## 2. Topology

```
                 Origin (Harbor)
                  ▲            ▲
        on-demand │ sync       │ (read-only robot account per site)
     ┌────────────┴──┐   ┌─────┴─────────┐
     │ Site X        │   │ Site Y        │   ... up to ~10 sites
     │  zot-a  zot-b │   │  zot-a  zot-b │   each node: own local disk, full cache
     │    ▲      ▲   │   │    ▲      ▲   │
     │    └─ LB ─┘   │   │    └─ LB ─┘   │   Kubernetes Service, or DNS over 2 VMs
     │      ▲        │   │      ▲        │
     │   clients     │   │   clients     │
     └───────────────┘   └───────────────┘
```

Each site has:

- **Two zot nodes with the same configuration.** Each node is a full on-demand cache of the origin on
  its own disk. The nodes share no state and do not talk to each other.
- **One site address** for clients:
  - Kubernetes: a Service in front of both pods.
  - VMs: a DNS name for both nodes. Plain DNS does not detect a dead node, so a VM site needs DNS
    with health checks (GSLB) or a floating IP (for example keepalived).

### Rejected alternatives

| Alternative | Why rejected |
|---|---|
| zot cluster mode (`http.cluster`) | It shards repositories: each repository belongs to one member, and the other member forwards requests to it. When the owner is down, the forward fails with HTTP 500 (`pkg/api/proxy.go`). With local disks this is not HA. |
| Primary and standby | Half of the capacity is idle, and the standby has a cold cache after a failover. |
| Chain (node B syncs from node A, then from the origin) | Asymmetric: a cold miss on A also slows B. A two-way chain can loop on a miss. The fallback from the first sync URL to the next was not verified. |

### Optional: route by repository

Where the site has an L7 load balancer (a Kubernetes ingress or HAProxy), it can route each request
by repository name with a consistent hash, and fall back to the other node when one fails. Each
repository then normally lives on one node, so the origin sends it once per site instead of twice.
Failover still works, because the other node fetches on demand. This is optional; plain round robin
also works.

## 3. Request flow and failure behavior

1. **Hit:** the node serves the image from its local disk. With `manifestCheckInterval: -1`, it does
   not contact the origin.
2. **Miss:** the node fetches the image from the origin and serves it. With `stream`, the client
   receives data while the node downloads. Concurrent pulls of the same image on one node share one
   fetch. The two nodes do not share fetches, so the origin sends an image at most twice per site.

| Failure | Effect |
|---|---|
| Origin down | Hits keep working. Misses fail. |
| One node down | The other node serves all pulls. Its cache may be colder, so some pulls become misses. |
| Whole site down | New pulls at that site fail until repair. Running containers are not affected. |

## 4. Edge node configuration

Both nodes of a site use the same configuration. Only the site name and the origin credential
differ between sites.

### Sync from the origin

| Setting | Value | Reason |
|---|---|---|
| `onDemand` | `true` | Fetch an image when a client first requests it. |
| `manifestCheckInterval` | `-1` | A stored tag is served without contacting the origin. Pinned tags never change, so the check has no value, and it fails or waits when the origin is slow or down. A negative value was added by PR #4646. |
| `content` | `**`, or only the origin projects the site needs | Limits what the site can cache. |
| `credentialsFile` | A read-only robot account on the origin | The edge syncs with its own credential; it never forwards the client's identity. |
| `tlsVerify` | `true` | Required by `stream`. |
| `stream` | `true` | Faster cold misses: the client gets data during the download, not after the whole image is stored (PR #4407, merged 2026-10-10). It requires HTTPS with `tlsVerify`, and it does not work together with `maxRetries`/`retryDelay` or `onDemandInBackground`. |

If `stream` is not used, set `maxRetries` low, so that a miss fails quickly when the origin is down
instead of after minutes of retries.

On a version without PR #4646 (v2.1.23 or older), a positive `manifestCheckInterval` (for example
`24h`) is the closest option: it checks the origin once per interval for each tag instead of on every
pull.

### HTTP server

| Setting | Value | Reason |
|---|---|---|
| `http.writeTimeout` | Longer than the pull of the largest image | With the default of 60 s, one run of our cold-pull benchmark (64 clients, one 1 GB image) cut off 15 clients. |
| `http.shutdownTimeout` | For example `30s` (the default) | On restart, zot finishes in-flight pulls before it stops (PR #4638). |

### Storage and eviction

- Local disk, `dedupe: true`, `gc: true`.
- `retention` with, for example, `pulledWithin: 720h` on `**`: tags that nobody pulled for 30 days
  are deleted, and GC then frees their blobs.
- This is eviction by age, not by size. Size the disk for the images pulled within one retention
  window, and alert on disk use (section 5). The GC API (PR #4522, open) can trigger a cleanup on
  demand.

### Access

Clients only read: no `create`, `update` or `delete` on the edge. Authentication is chosen per
deployment (section 6).

## 5. Operations

### Deployment

- **Kubernetes site:** a StatefulSet with 2 replicas, one PVC per pod on local storage, a Service for
  clients, and pod anti-affinity so that the two pods run on different nodes.
- **VM site:** 2 VMs with the same zot configuration, behind health-checked DNS or a floating IP.
- **Configuration:** one template for all sites.

### Upgrades and restarts

- Restart one node at a time. The other node serves meanwhile, and `shutdownTimeout` lets the
  restarting node finish its in-flight pulls.
- A restart keeps the cache, because it is on the local disk. A replaced disk starts empty and fills
  on demand.

### Monitoring

Use the zot `metrics` extension with Prometheus:

- Per node: up or down, request rate, errors, disk use against capacity.
- Sync errors toward the origin: they show an origin outage or a bad credential.
- Alerts: only one healthy node at a site; disk use above a threshold.

### Sizing (to be measured in each environment)

- **Disk per node:** the images pulled within one retention window.
- **CPU and memory:** in our benchmarks (zot v2.1.21, 10 MB images), serving from a warm cache cost
  about 0.9 vCPU per GB/s, with a zot RSS of 160–190 MB; spare RAM helps as page cache for the hot
  images. Filling the cache cost about 10 times more CPU per GB than serving it, and the VM used
  0.6–0.7 GB of memory.
- **WAN to the origin:** at most two fetches per image per site (one with the optional routing by
  repository).

## 6. Access control options

zot supports these; choose per deployment.

| Option | How | Works when the origin is down |
|---|---|---|
| Open read | `anonymousPolicy: ["read"]` on `**` | Yes |
| Groups at the edge | LDAP (`userGroupAttribute`) or OIDC (groups claim), with `accessControl` rules per repository pattern that mirror the origin's projects | Yes. The rules are kept in two places: the origin and the edge configuration. |
| Origin decides | zot `bearer` auth with the origin's token service as `realm` and the origin's token-signing certificate as `cert`; zot checks the signature and the `access` claim | No: clients need the origin for every new token. Read from code only (zot does not check the token's `service` claim); not tested. |

mTLS client certificates and API keys are also available for machine clients.

## 7. Future add-on: prewarming

Not designed in detail. Notes for later:

- An external trigger (an origin webhook on push, or a CI step) asks the edge nodes to fetch an image
  before a rollout pulls it.
- It must reach **each node directly**, not the site address, so that both caches fill.
- Without `stream`, a manifest GET makes the node store the whole image before it answers, so the
  trigger can send only manifest GETs (the index and each wanted platform manifest). Platform
  manifests are needed because on-demand sync stores a multi-arch index without its platform
  images, and syncs each platform only when it is requested.
- Open question: with `stream: true`, does a manifest-only GET still store the blobs? Not tested.

## 8. Validation before rollout

Run these on a test setup with the origin and one site of two nodes:

1. **Hit:** a stored pinned tag is served with no request to the origin, also while the origin is
   down.
2. **Miss with `stream`:** cold pulls of a small and a 1 GB image. Measure the time to the first byte
   and the total time. Then stop the origin and pull again: the image must be fully stored.
3. **Miss while the origin is down:** the client gets an error quickly, not after minutes.
4. **Node failure:** stop one node during pulls. New pulls go to the other node; pulls in progress on
   the stopped node finish or fail cleanly.
5. **Rolling upgrade:** restart both nodes one after the other while clients pull. No client
   failures expected.
6. **Retention and GC:** a tag not pulled within the window is deleted and its disk space comes back;
   a recently pulled tag stays.
7. **Load:** warm throughput, and a burst of cold pulls of the same image on both nodes (expect two
   fetches from the origin, one per node).
8. **Access control:** each option from section 6 that the deployment uses. The origin-token option
   needs a real test.

## 9. Risks

| Risk | Mitigation |
|---|---|
| The design depends on a negative `manifestCheckInterval` (PR #4646), which is merged but not yet released, and on `stream` and `shutdownTimeout`, which are new in v2.1.23. | Pin the build; repeat section 8 for each new version. |
| `stream` is new code; its behavior with interrupted clients is not known. | Validation tests 2 and 4. |
| No disk size limit: a burst of new images can fill the disk before retention runs. | Disk alerts; size for the retention window; the GC API (PR #4522) when available. |
| HA at VM sites depends on health-checked DNS or a floating IP, outside zot. | Choose one per site before rollout. |

## 10. References

- Warm-cache benchmark: rophy/zot-perf-test, `docs/results/edge-cache-warm-throughput-report.md`.
- Cold-pull benchmark: rophy/zot-perf-test, `docs/results/edge-cache-cold-pull-report.md`.
- zot PRs: #4407 (`stream`), #4638 (`shutdownTimeout`), #4646 (negative `manifestCheckInterval`), #4522
  (GC API), #4410 (sparse on-demand sync of multi-arch indexes).
- zot configuration reference: `examples/README.md` (sync, retention, authentication, access
  control).
