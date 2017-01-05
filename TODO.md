# TODO

## Priority I
- [ ] write unit tests
  - [ ] test to ensure that adding the inhibitory cell is working
  - [ ] apply backtrace
  - [ ] processBatch
- [ ] the number of synapses grows hugely and hangs the network around 11%
- [ ] never finds existing inhibitory synapses
- [ ] huge numbers of good synapses with no bad synapses

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
- [ ] Genericize training and data so it can be driven by a UI
