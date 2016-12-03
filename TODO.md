# TODO

## Near term

- [x] only grow characters that were trained by a given - round
- [ ] prune seems to have a bug where it leaves dangling - synapses
- [ ] add tests for synapse/cell integrity testing - before and after growth, pruning, and diffing
- [ ] move charrnn training from shake.go into its repo
- [ ] NewCellID and NewSynapseID should be an instance method on a network to ensure no ID collisions occur
- [x] build charrnn sampling code
- [ ] more aggressive pruning
- [ ] Important to save vocab with the network

## Optimizations
[x] use int32 for IDs
[ ] RandomCellKey method is pretty slow at scale

## Once working
[ ] look for properties of types that can probably be private (lower case them)
[ ] use [SIMD](https://github.com/bjwbell/gensimd) instructions
    - once it works, generate assembly for the low level math, to make it even faster and efficient.
[ ] distributed computing methods
