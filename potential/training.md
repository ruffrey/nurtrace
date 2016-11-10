# Training

A single cell can be an *output* and another cell can be an *input*.

For every input there is a corresponding output.

There must be a path from input to output.

A path is not stored anywhere, it just exists in the network, and can change.

The network is this big unorganized mass of cells and synapses at first.

We decide we want to put something in, and get something out.

So we randomly choose some input cells. We randomly choose an output cell.

It is not enough that the input and output connect; a rapid series of action potential firings
on the input cell must result in the output cell firing.

The input will be a coded value that has meaning inside the network, and the output will be
a value coded on the same map/scale as the input.

What does it matter that a signal passes from input and comes out on an output? How are the
possible values stored in the network? It can work like the brain.

Specialized perception cells receive the data (like rods and cones in the eye). These are connected
to an intermediary consolidation network which always fire and represent the data as a pattern
of responses. In the case of speech generation based on a primer:
- vocab perception
    - one character per receptor (layer)
    - feed X characters in a row into the network via the input cells, which creates a firing pattern
    - the firing pattern makes synapses stronger automatically (fire together, wire together)
    - the firing path which reaches the expected output will get a reverse flow boost

Explained a slightly different way:
- the input layer has one receptor representing each character
- the output layer has one receptor representing each character
- for a single trial you know:
    - input character
    - expected output character
- feed the input character into the network.
    - whichever output character gets activated first, you reinforce the synapses on
    the path it took, back to the input.
    - this is possible because the whole time, we were tracking the activation path.

Do this for all the characters in, say, one sentence. This is a trial.

This enables multithreaded training:
- you want to do 8x threads
- take a snapshot of the network
- copy it 8x
- grab 8 difference sentences
- retain a master original snaptshot of the network
- run your training trial on a sentence, backpropagating the whole time
- diff the master original and each freshly trained copy network
- apply the diff to the master original
- now the network has been trained on 8 sentences at once
- we may want to do way more than 8x threads of training, because getting from input to output
may take a while at first, but should get faster and faster over time.

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
