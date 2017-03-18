# TODO

## Priority I
- [x] increase synapse weights to int16
  - 5 vs 125 or 5 vs 32,000 is way more efficient and provides greater contrast
- [ ] update laws for int16
- [ ] add spike-time dependent plasticity (reinforce cells that result in firing that round)
- [ ] Figure out how the brain consolidates and prunes during sleep
- [ ] Re-implement and add pruning cycle
  - removal of unreachable cells / no synapses
  - degrading or removal of less-firing cells
  - adding dendrites to cells that fired a lot / strong pathways
  - do some kind of cyclical or regular firing pattern which does activations
  and filters out unnecessary pathways.
- [ ] ensure that adding the inhibitory synapses is working, and inhibiting the right thing.

## Priority II
- [ ] Pull things that change less into sub files
- [ ] change terminology to pre- and post- synaptic neurons where it'll make it easier to follow
- [ ] Periodically save back the threaded training to original
    - currently we lose all training on a crash
    - saving can also cause concurrent map read/writes and fail

## Priority III
- [ ] add logging with glog
- [ ] RandomCellKey method is pretty slow at scale
- [ ] return errors instead of logging or doing a panic
- [ ] add methods for making and removing connections between synapses and cells on a network
    - the dual relationship assignment, which also requires mutexes, are error prone
    - [x] adding
    - [ ] removing

## Later - once it works
- [ ] Design distributed training architecture: desktop UI, CLI/services, server, cloud?
- [ ] Add word-level and phrase-level neural networks
- [ ] look for properties of types that can probably be private (lower case them)
