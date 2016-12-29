# TODO

## Priority I
- [ ] backtracing
  - [x] training batch path tracing output->input and marking good synapses
  - [x] training batch path tracing output->input and marking bad synapses
  - [x] duplicate a synapse when it goes over the int8 limit
  - [x] inhibit bad synapse paths
  - [x] reinforce good synapses
  - [ ] backtracing or training with backtracing seems to hang

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
- [ ] Generic training and data that can be driven by a UI
