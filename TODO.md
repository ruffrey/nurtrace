# TODO

## Top of the Stack
- [ ] try firing sweeps (`network.Step()`) instead of timing, to avoid issues with timeout functions
- [ ] More refined learning techniques:
    - [ ] ensure everything in learning-mechanisms.md is done
    - [ ] larger sets of pathways are OK - more synapses between start and end
- [ ] do multiple threads differently so they wont block on each batch of lines
- [ ] more aggressive pruning and optimize learning more granularly to reduce randomness:
    - [ ] if it fails, do a prune on the network copy, then GrowPathBetween, and apply it to the original
- [ ] Find all "laws of the universe" constants and collect in one place

## Optimizations and Refactoring
- [ ] If vocab is saved, network must be also
- [ ] RandomCellKey method is pretty slow at scale
- [ ] move charrnn training from shake.go into charrnn repo, adding as much as possible to the main lib

## Later - once it works
- [ ] code for word level and phrase level neural networks
- [ ] look for properties of types that can probably be private (lower case them)
- [ ] distributed computing methods
- [ ] parallelize and use [SIMD](https://github.com/bjwbell/gensimd) instructions?
