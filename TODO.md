# TODO

## Top of the Stack
- [ ] issue with diffing multiple networks that were independently trained, or vocab is getting out of sync.
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
