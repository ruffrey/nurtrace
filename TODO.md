# TODO

## Top of the Stack
- [ ] issue with integrity
    - When we add a cell during ApplyDiff, occasionally it is missing one of its synapses
    - This ONLY occurs after pruning.
    - Pruning is leaving cells with synapses that are listed on the cell, but not on the network.
    - However, the network passes an integrity check after being pruned.
    - Applying the diff of a network that was pruned is breaking.
- [ ] synapse diffs are never greater than zero now (might be ok?)
- [ ] infrequently, an added cell has one dendrite connection where the synapse
is not on the network after applying a diff
- [x] got this issue - with no pruning:
```
$ go run shake.go -train=4
Loaded network from disk, 1630 cells 18186 synapses
Reading test data file shake.txt
Loaded training text for shake.txt samples= 32777
Beginning training 4 simultaneous sessions
charrnn data setup complete
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x28 pc=0x60537]

goroutine 1304 [running]:
panic(0x14a600, 0xc420014080)
	/usr/local/go/src/runtime/panic.go:500 +0x1a1
bleh/potential.(*Cell).FireActionPotential(0x0)
	/Users/jeffparrish/go/src/bleh/potential/cell.go:101 +0x37
bleh/potential.processBatch(0xc4200c7400, 0x2d, 0x40, 0xc422a86f90, 0xc4238cbc20, 0xc4207f6000, 0x1, 0x0, 0x0, 0x0, ...)
	/Users/jeffparrish/go/src/bleh/potential/trainer.go:172 +0xcc
bleh/potential.Train.func1(0xc4200103f0, 0xc4205cea50, 0xc42315af60, 0xc423883ec0, 0x2)
	/Users/jeffparrish/go/src/bleh/potential/trainer.go:112 +0x102
created by bleh/potential.Train
	/Users/jeffparrish/go/src/bleh/potential/trainer.go:131 +0x181
exit status 2
```
- [ ] More refined learning techniques:
    - [ ] ensure everything in learning-mechanisms.md is done
    - [ ] larger sets of pathways are OK - more synapses between start and end
    - [ ] more aggressive pruning and optimize learning more granularly to reduce randomness:
    - [ ] if it fails, do a prune on the network copy, then GrowPathBetween, and apply it to the original (?)
- [ ] Periodically save back the threaded training to original
    - currently we lose all training on a crash
- [ ] add logging with glog

## Optimizations and Refactoring
- [ ] If vocab is saved, network must be also
- [ ] RandomCellKey method is pretty slow at scale
- [ ] occasionally fail diff when using more workers than threads

## Later - once it works
- [ ] code for word level and phrase level neural networks
- [ ] look for properties of types that can probably be private (lower case them)
- [ ] distributed computing methods
- [ ] parallelize and use [SIMD](https://github.com/bjwbell/gensimd) instructions?
