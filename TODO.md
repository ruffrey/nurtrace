# TODO

## Priority I
- [x] Crash due to accessing a cell's synapse list while it is being written to (diff.go:233, diff.go:306)
- [ ] crash - getting a dendrite cell where it does not exist
```
<local> local thread 2 done
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x6a49b]

goroutine 26 [running]:
panic(0x213260, 0xc4200120c0)
	/usr/local/go/src/runtime/panic.go:500 +0x1a1
bleh/potential.copySynapseToNetwork(0xc42363c690, 0xc438fac380, 0xc4aa18827f)
	/Users/jpx/go/src/bleh/potential/diff.go:298 +0x24b
bleh/potential.CloneNetwork(0xc420077840, 0xc4200e00e0)
	/Users/jpx/go/src/bleh/potential/diff.go:201 +0x2a1
bleh/potential.Train.func1(0x0, 0xc42014d4a0, 0xc4200d5c80, 0xc420164588, 0x3b, 0xc5, 0xc420086928, 0xc4200c8f60, 0xc4200e00e0, 0xc420197020, ...)
	/Users/jpx/go/src/bleh/potential/trainer.go:276 +0x1a3
created by bleh/potential.Train
	/Users/jpx/go/src/bleh/potential/trainer.go:287 +0xcab
```
- [x] we are adding similar or the same inhibitory cells over and over
- [ ] ensure that adding the inhibitory synapses is working, and inhibiting the right thing.
- [ ] Re-implement and add pruning cycle
  - removal of unreachable cells / no synapses
  - degrading or removal of less-firing cells
  - adding dendrites to cells that fired a lot / strong pathways
  - do some kind of cyclical or regular firing pattern which does activations
  and filters out unnecessary pathways.

## Priority II
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
