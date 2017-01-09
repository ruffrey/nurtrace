# TODO

## Priority I
- [ ] write unit tests
  - [ ] test to ensure that adding the inhibitory cells are working
  - [ ] apply backtrace and the various supporting it
  - [ ] processBatch, particularly when doing backtraces
- [ ] the number of synapses grows hugely and hangs
- [ ] never reuses existing inhibitory synapses
  - [ ] unit test run input/output, add inhibitory synapse, make sure it inhibits and all expected cells fire
- [ ] make the network deeper and let sampling or training run more steps to propagate through

## Priority II
- [ ] Periodically save back the threaded training to original
    - currently we lose all training on a crash
    - saving can also cause concurrent map read/writes and fail
- [ ] faster backtracing
  - bleh/potential.applyBacktrace 9.06s(4.61%) of 33.81s(17.21%)
  - bleh/potential.backwardTraceFirings.func1.1 6.70s(3.41%) of 34.71s(17.67%)

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
