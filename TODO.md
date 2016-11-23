# TODO

## Near term
[ ] finish diffing
[ ] consider using IDs between synapses and cells, instead of pointers, for portability during diffs
[x] switch all lists to be maps
[ ] figure out how to serialize and store memory (it is an interface{})
[ ] attach receptors/perceptors cleanly to the network
[ ] add training methods
[ ] NewCellID and NewSynapseID should be an instance method on a network

## Once working

[ ] use [SIMD](https://github.com/bjwbell/gensimd) instructions
    - once it works, generate assembly for the low level math, to make it even faster and efficient.
