# TODO

## Top of the Stack
- [ ] issue with integrity
    - most likely caused by synapse naming collisions during diff. it happens (?only?) when there
        are multiple networks
    - when there are multiple threads, pruning tends to fail on the next load from disk fresh training
    - diffing multiple networks that were independently trained, or vocab is getting out of sync
    - maybe it is because diffing does not currently handle pruning unused old cells or synapses,
        and we do not prune after the final merge. so a network is going to have pruned
        synapses and perhaps cells! but those won't be reflected on the original network.
    [ ] solve the problem of pruning cells and/or synapses and then having collisions and leaving
        them dangling, when merging networks back together.
- [ ] More refined learning techniques:
    - [ ] ensure everything in learning-mechanisms.md is done
    - [ ] larger sets of pathways are OK - more synapses between start and end
    - [ ] more aggressive pruning and optimize learning more granularly to reduce randomness:
    - [ ] if it fails, do a prune on the network copy, then GrowPathBetween, and apply it to the original (?)
- [ ] Periodically save back the threaded training to original
- [ ] add logging with glog

## Optimizations and Refactoring
- [ ] If vocab is saved, network must be also
- [ ] RandomCellKey method is pretty slow at scale
- [x] move charrnn training from shake.go into charrnn repo, adding as much as possible to the main lib

## Later - once it works
- [ ] code for word level and phrase level neural networks
- [ ] look for properties of types that can probably be private (lower case them)
- [ ] distributed computing methods
- [ ] parallelize and use [SIMD](https://github.com/bjwbell/gensimd) instructions?
