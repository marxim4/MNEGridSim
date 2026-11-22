# ADR 001 — Deterministic core

**Status** — Accepted

## Context

The first simulation step allocates available power among consumers. Every permutation of the same requests must produce the same result. The power balance must be exact even when total demand exceeds the maximum `int64` value.

## Decision

The `sim` package exposes allocation through the `Allocate` function. It operates solely on its arguments and uses the Go standard library. It copies the requests and sorts them by ascending priority, then lexicographically by ID. It validates every request before allocating power. Each allocation reduces the remaining power budget in kW, so the function can compute the balance without summing demand.

## Alternatives

- Sorting the input slice in place saves a copy at the cost of changing the caller's request order.
- A priority queue adds a data structure, and producing ordered output for all consumers can still take `O(n log n)` time.
- A three-layer or hexagonal architecture can isolate the core from external systems when integrations are needed. At this stage, those approaches would introduce interfaces and boundaries for dependencies that do not yet exist.

## Consequences

Callers can repeat a simulation step with the same available power and requests and get the same power balance and output order. The request copy, ID tracking, and output use `O(n)` space. Future callers can perform I/O outside the package and use the same function contract.
