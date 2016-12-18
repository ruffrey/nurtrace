# TODO

## Top of the Stack
- [ ] issue with integrity
    - When we add a cell during ApplyDiff, occasionally it is missing one of its synapses
    - This ONLY occurs after pruning.
    - Pruning is leaving cells with synapses that are listed on the cell, but not on the network.
    - However, the network passes an integrity check after being pruned.
    - Applying the diff of a network that was pruned is breaking.
- [ ] More refined learning techniques:
    - [ ] ensure everything in learning-mechanisms.md is done
    - [ ] larger sets of pathways are OK - more synapses between start and end
    - [ ] more aggressive pruning and optimize learning more granularly to reduce randomness:
    - [ ] if it fails, do a prune on the network copy, then GrowPathBetween, and apply it to the original (?)
- [ ] Periodically save back the threaded training to original
    - currently we lose all training on a crash
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
