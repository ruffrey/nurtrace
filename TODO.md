# TODO

## Priority I
- [ ] backtracing is a good start, but it needs to be more throught through.
  - [ ] rewrite backtracing article
  - [ ] reimplement or fix backtracing
  - [ ] adding inhibitory synapses on noise should just be done during traversal so we have context
  - [ ] in trainer.go, consider treating failed expected batches differently from noise
- [ ] never reuses existing inhibitory synapses
- [ ] the number of synapses grows hugely and hangs
- [ ] write unit tests
  - [ ] test to ensure that adding the inhibitory cells are working
  - [ ] apply backtrace and the various supporting it
  - [ ] processBatch, particularly when doing backtraces
  - [ ] unit test run input/output, add inhibitory synapse, make sure it inhibits and all expected cells fire

## Priority II
- [ ] Periodically save back the threaded training to original
    - currently we lose all training on a crash
    - saving can also cause concurrent map read/writes and fail

## Priority III
- [ ] faster backtracing
  - [ ] bleh/potential.backwardTraceFirings.func1.1 7.26s(4.77%) of 33.60s(22.06%)
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
