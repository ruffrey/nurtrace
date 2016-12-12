# TODO

## Top of the Stack
- [ ] Vocab seems to have bad references when reloaded
    - [ ] If vocab is saved, network must be also
    - [x] Diffing might appear to have issues due to the code we added
    - [x] Poor integrity:
        - when diffing, a new network might have happened to add a new cell or synapse, and during the
            same time another network would have added one with the same ID. However their dendrites are
            different and may reference other cells. This needs smarter diffing.
        - `warn: cannot grow path because synapse axon does not exist 1381766920 from cell= 2193275927`
        - `error: cannot activate synapse 1278366234 from cell 2465421962 because it does not exist`
- [ ] Find all "laws of the universe" constants and collect in one place
- [ ] More refined learning techniques:
    - [ ] ensure everything in learning-mechanisms.md is done
    - [ ] larger sets of pathways are OK - more synapses between start and end
- [ ] do multiple threads differently so they wont block on each batch of lines
- [ ] more aggressive pruning and optimize learning more granularly to reduce randomness:
    - [x] Do a large Grow() before starting the entire session.
    - [x] only train each input and output once.
    - [x] wait only `synapseDelay * maxHops` for an to fire an output
    - [ ] if it fails, do a prune on the network copy, then GrowPathBetween, and apply it to the original
- [x] only grow network when it fails. otherwise it grows out of control
- [ ] try firing sweeps instead of timing, to avoid issues with timeout functions

## Optimizations and Refactoring
- [x] Important to save vocab with the network
- [ ] RandomCellKey method is pretty slow at scale
- [ ] move charrnn training from shake.go into charrnn repo, adding as much as possible to the main lib

## Later - once it works
- [ ] code for word level and phrase level neural networks
- [ ] look for properties of types that can probably be private (lower case them)
- [ ] distributed computing methods
- [ ] parallelize and use [SIMD](https://github.com/bjwbell/gensimd) instructions?
