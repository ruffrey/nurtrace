# TODO

## Top of the Stack

- [ ] when diffing, a new network might have happened to add a new cell or synapse, and during the
same time another network would have added one with the same ID. However their dendrites are
different and may reference other cells. This needs smarter diffing.
- [x] add more tests for synapse/cell integrity testing - before and after growth, pruning, and diffing
- [ ] more aggressive pruning and optimize learning more granularly to reduce randomness:
    - [ ] Do a large Grow() before starting the entire session.
    - [ ] only train each input and output once.
    - [ ] wait only `synapseDelay * maxHops` for an to fire an output
    - [ ] if it fails, do a prune on the network copy, then GrowPathBetween, and apply it to the original
- [x] see if multiple threads work
- [ ] do multiple threads differently so they wont block on each batch of lines

## Optimizations and Refactoring

- [ ] Important to save vocab with the network
- [x] use int32 for IDs
- [ ] RandomCellKey method is pretty slow at scale
- [x] make recalculate version computationally less difficult
- [ ] move charrnn training from shake.go into charrnn repo, adding as much as possible to the main lib
- [x] NewCellID and NewSynapseID should be an instance method on a network to ensure no ID collisions occur

## Later - once it works
- [ ] look for properties of types that can probably be private (lower case them)
- [ ] distributed computing methods
- [ ] parallelize and use [SIMD](https://github.com/bjwbell/gensimd) instructions?
