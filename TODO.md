# TODO

## Priority I
- [ ] More refined learning techniques:
    - [ ] Fix overpruning that results in no synapses at the end of training
    - [ ] ensure everything in learning-mechanisms.md is done
    - [ ] larger sets of pathways are OK - more synapses between start and end
    - [ ] more aggressive pruning and optimize learning more granularly to reduce randomness
    - [x] if it fails, do a prune on the network copy, then GrowPathBetween, and apply it to the original (?)
- [ ] NewNetwork should return a pointer

## Priority II
- [ ] profile CPU again
- [ ] Periodically save back the threaded training to original
    - currently we lose all training on a crash
    - saving can also cause concurrent map read/writes and fail

## Priority III
- [ ] add logging with glog
- [ ] RandomCellKey method is pretty slow at scale
- [ ] return errors instead of logging or doing a panic
- [ ] add methods for making and removing connections between synapses and cells on a network
    - the dual relationship assignment, which also requires mutexes, are error prone

## Later - once it works
- [ ] Design distributed training architecture: desktop UI, CLI/services, server, cloud?
- [ ] Add word-level and phrase-level neural networks
- [ ] look for properties of types that can probably be private (lower case them)
- [ ] distributed computing methods
- [ ] parallelize and use [SIMD](https://github.com/bjwbell/gensimd) instructions?
