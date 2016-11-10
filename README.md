# Name Pending

This is a real neural network, modeled after real neuring firing patterns.

It uses the concept of the Action Potential and Membrane Potential from neuroscience to
train a neural network.

## Goals
- simple math resembling millivolts in neural network firings
- extremely fast and works cross platform
- parallelized learning and inputs - instead of only training on one thing at a time, can train
on more than one stimuli simultaneously
- focus is on the pathway, rather than the neural network solving only one problem

## Stretch Goals
- generalized model can be used for any kind of data that can be put in memory
- same network can be trained on more than one kind of stimuli to solve different problems

## TODO:
- figure out how to serialize and store memory
- figure out how to sample / receive output
    - to receive a sample, specify an input point and output point
- figure out how to train
    - to train, specify an input point and output point. get the output value. if good,
    backpropagate it by reinforcing the connections in that path. if bad, decrement the activated
    pathways.
