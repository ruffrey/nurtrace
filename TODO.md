# TODO

## Priority I
- [ ] write unit tests
  - [ ] test to ensure that adding the inhibitory cells are working
  - [ ] apply backtrace and the various supporting it
  - [ ] processBatch, particularly when doing backtraces
- [x] only inhibit the beginnings of bad paths
- [ ] grow paths using fewer new synapses
- [ ] the number of synapses grows hugely and hangs the network around 11%
- [ ] never reuses existing inhibitory synapses

## Priority II
- bleh/potential.backwardTraceFirings.func1.1 17.98s(6.46%) of 99.53s(35.76%)
- [ ] Periodically save back the threaded training to original
    - currently we lose all training on a crash
    - saving can also cause concurrent map read/writes and fail
- [ ] faster backtracing

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
- [ ] Genericize training and data so it can be driven by a UI
