# TODO

## Near term
[ ] finish diffing
[x] use IDs between synapses and cells, instead of pointers, for portability during diffs
[x] switch all lists to be maps
[ ] switch Cell arrays to be maps
[ ] look for properties of types that can probably be private (lower case them)
[ ] implement: serialize the network and deserialize it
    - assign the pointer to the network on each cell and synapse
[ ] attach receptors/perceptors cleanly to the network
[ ] add training methods
[ ] NewCellID and NewSynapseID should be an instance method on a network to ensure
no ID collisions occur

## Once working

[ ] use [SIMD](https://github.com/bjwbell/gensimd) instructions
    - once it works, generate assembly for the low level math, to make it even faster and efficient.
