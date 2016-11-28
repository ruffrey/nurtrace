# Training

A single cell can be an *output* and another cell can be an *input*.

For every input there is a corresponding output.

There must be a path from input to output.

A path is not stored anywhere, it just exists in the network, and can change.

The network is this big unorganized mass of cells and synapses at first.

We decide we want to put something in, and get something out.

It is not enough that the input and output connect; a rapid series of action potential firings
on the input cell must result in the output cell firing *first*.

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
- the output layer has one perceptor representing each character
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

# No backpropagation.

Instead, run a few rapid trials, inputting one letter and seeing if it results in the output firing. If it does, do not discard the learning. If it doesn't, discard this trial. Then grow the network to try to get a connection between the input and desired output. Growing can be done by adding neurons, adding synapses, and maybe pruning. Not sure about pruning. Pruning may be better to reduce noise later. Or, perhaps after a successful series of trials, we prune synapses.

This avoids the expensive and complicated path tracing. It is better than brute forcing because we change the network in between trials.

# Timing

Timing is an interesting problem to solve during training and sampling. While a brain is always
working with the same hardware, a neural network may be on vastly different hardware. Additionally
there may be slightly different sets of processes running in the background which degrade training
or sampling speed in small ways. While small, it is a challenge because a learning and thinking
brain does not experience vast variations in hardware on det

- the timing solution is to always add a millisecond sleep before applying the voltage
