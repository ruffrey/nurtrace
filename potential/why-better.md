## Why this is better

### It is faster

#### The vast majority of the math operations will be simple addition and subtraction

Normally there's all this matrix multiplication, sigmoids, logrithms, math loops within loops,
and other expensive things which must be computed on a GPU with hard-to-conceptualize GPU
programming. This leads to abstraction libraries that can be easy-ish, yet contain inflexible
concepts. The high barrier of entry to those who are not hardcore programmers remains below the
surface. There are fewer hardcore programmers than regular programmers want you to believe.

#### Parallelizable design by nature

Copy the network, train it on some data on each copy, capture the diff, then merge it back to
master at a given point.

#### Able to be distributed

You might want to train a network on two different datasets and merge them, then grow and
prune it, then train across two additional datasets for a total of four. That's ok. The merging is
the key part here. You fast forward each set of weights later when you reconvene.

What it means if you can take a new or partially trained network, copy it, split up your
computation onto multiple machines, split it again to take advantage of all cores, then
merge back to master.

#### Designed to get around overfitting/overtraining

If for some reason a network gets overfitted to a dataset, it can be grown sufficiently to
reduce its effectiveness due to having too many untrained new connections and cells. It will
still retain the training, and the training can continue using the original data and new data
to avoid overfitting in the future.

#### Keep growing the network

##### More neurons
If the network is not performing well on a new dataset, then you can add more neurons and synapses
and keep trying. No training is really lost.

Traditional neural networks seem to suffer from the ability to not be able to change their
size - number of layers and number of cells per layer.

##### New input types

This is where it gets really interesting.

An existing trained network can be applied to a related problem. Just as us humans leverage
existing concepts, this network architecture can use existing pathways to learn to solve new
classes of problems, provided they are somewhat similar. Heck you could even train it to do
something totally different, but you'd be training it from scratch. Importantly though would
be that we need to keep training it on the original dataset at the same time, too. Of course
this could happen in parallel.

## Limitations

### Memory

Keeping everything in memory is crucial to fast training. Having a huge network will be important,
and it will be hard to do that while keeping it all in memory.

Possible solutions are renting huge servers, buying slightly older but cheaper RAM, or using swap
on PCIe solid state drives. Scaling horizontally is only helpful so long as you can fit the
entire dataset in memory.
