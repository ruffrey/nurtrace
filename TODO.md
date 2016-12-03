# TODO

## Top of the Stack

- [ ] prune seems to have a bug where it leaves dangling - synapses
```
error: synapse 389969456 does not exist on cell 3245169848
error: synapse 2347244027 does not exist on cell 3245169848
error: synapse 389969456 does not exist on cell 3245169848
```
```
warn: synapse attempted to be accesed but it was already removed 3347374729
warn: synapse attempted to be accesed but it was already removed 3349803347
warn: synapse attempted to be accesed but it was already removed 3349781045
```
- [ ] add more tests for synapse/cell integrity testing - before and after growth, pruning, and diffing
- [ ] more aggressive pruning and optimize learning more granularly
- [ ] see if multiple threads work

## Optimizations and Refactoring

- [ ] Important to save vocab with the network
- [x] use int32 for IDs
- [ ] RandomCellKey method is pretty slow at scale
- [ ] recalculate shasum more often, and make it less computationally difficult
- [ ] move charrnn training from shake.go into its repo
- [ ] NewCellID and NewSynapseID should be an instance method on a network to ensure no ID collisions occur

## Later - once it works
- [ ] look for properties of types that can probably be private (lower case them)
- [ ] use [SIMD](https://github.com/bjwbell/gensimd) instructions
    - once it works, generate assembly for the low level math, to make it even faster and efficient.
- [ ] distributed computing methods
