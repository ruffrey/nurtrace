# Learning

The underlying concepts of learning are outlined below.

## Setup and grow

A network is grown from scratch, randomly adding lots of cells and synapses.

## Ensure paths between inputs and outputs

A perception data set has pairs of inputs with expected outputs. We make sure there is a
sufficiently strong path between the inputs and the outputs. Then we make sure that firing
the input cell results in firing the desired output cell.

## Feed the data

Since we know there are connections between the expected inputs and outputs, we can now
feed data into the network, saving the number of firings. We feed a decent amount of data
into the network at a time, and listen on the outputs to see when they fired.

## Reinforce the proper firing

When they did fire, we want to reinforce those paths (like backpropagation). How do we do that?
Synapses track their activation history. As they are activated, if they reach a certain threshold
where they have fired a lot, the cell will boost the number of millivolts it boots the cells
it fires.

## Internal cell and synapse mechanisms

A synapse's activation history is similar.

A synapse also can increase its millivolts. It tracks a second activation history counter, which,
upon reaching higher and higher thresholds depending on millivolts, will eventually reset
but also boost the millivolts.

Similarly, if at the end of a training session, a synapse did not fire very much, or at all,
it gets pruned, then its cells get pruned.

## How to know when a cell fired - getting data out of the system

When a cell is an output cell, it will have a golang `chan CellID`. A firing cell will send
its `CellID` into the channel. This acts as an indication of "perception." There would be a
construct where the data type that is being trained sets up the channel when it sets up inputs
and outputs, and attaches it to the outputs.
