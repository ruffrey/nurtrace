# TODO

## Priority I
- [ ] the number of synapses grows hugely and hangs the network around 11%
- [ ] always 0 bad synapses
- [ ] sampling never yields results with new weights
- [ ] reuse existing inhibitory synapses during backtracing and only create new one when not existing
- [x] backtraced paths, particularly the "bad" ones, are not marking very many synapses
- [x] recheck to make sure distributed training works
- [ ] write unit tests
  - [x] backtrace good - tree search
  - [x] backtrace bad - tree search
  - [ ] apply backtrace
  - [ ] processBatch
- [ ] backtracing has possibility of hanging due to circular circuits, possibly

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
