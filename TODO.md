# TODO

## Top of the Stack
- [ ] issues with integrity
    - [x] when the synapse ID changes during ApplyDiff, then Prune is called, and it gets pruned,
    we end up with that synapse being listed still on its axon and dendrite, however it does
    not exist on the network.
    - [ ] issue with integrity of ApplyDiff
- [ ] More refined learning techniques:
    - [ ] ensure everything in learning-mechanisms.md is done
    - [ ] larger sets of pathways are OK - more synapses between start and end
    - [ ] more aggressive pruning and optimize learning more granularly to reduce randomness:
    - [ ] if it fails, do a prune on the network copy, then GrowPathBetween, and apply it to the original (?)
- [ ] Periodically save back the threaded training to original
    - currently we lose all training on a crash
    - saving can also cause concurrent map read/writes and fail

## Optimizations and Refactoring
- [ ] If vocab is saved, network must be also
- [ ] RandomCellKey method is pretty slow at scale
- [ ] occasionally fail diff when using more workers than threads
- [ ] add logging with glog
- [ ] occasionally log progress percentage

## Later - once it works
- [ ] code for word level and phrase level neural networks
- [ ] look for properties of types that can probably be private (lower case them)
- [ ] distributed computing methods
- [ ] parallelize and use [SIMD](https://github.com/bjwbell/gensimd) instructions?
