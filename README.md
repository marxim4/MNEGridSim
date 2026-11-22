# MNEGridSim

MNEGridSim is a deterministic power system simulator written in Go. It aggregates
available power from multiple producers and allocates it to consumers by priority,
using consumer IDs to break ties. Scenarios define consumers and their power demand
across a sequence of independent simulation ticks.

## Run the demo

```sh
go run ./cmd/mnegridsim
```

The CLI runs a built-in scenario with synthetic data and reports allocated power,
unmet demand, and surplus in kW for each tick.

## Run the tests

```sh
go test ./...
```

Concurrency, network communication, and more detailed power system models are
planned for later milestones.
