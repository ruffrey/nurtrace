# Name Pending (codename nurtrace)

v0.12.0

This is a real neural network, modeled after real neuring firing patterns.

It uses the concept of the Action Potential and Membrane Potential from neuroscience to
train a neural network.

## Todos and Project Management

See the github issues for this project.

## Core lib architected to support:

- [x] simple math resembling millivolts in neural network firings
- [x] fast and works cross platform
- [x] can train in parallel on the same dataset
- [x] parallelized learning and inputs
      - instead of only training on one thing at a time, can train on more than one stimuli simultaneously
- [x] focus is on the pathway, rather than the neural network solving only one problem
- [x] generalized model can be used for any kind of data that can be put in memory
- [x] same network can be trained on more than one kind of stimuli to solve different problems
- [ ] methods for adaptive learning - training evaluates its learning speed and adjusts
- [ ] (maybe) parallelize and use [SIMD](https://github.com/bjwbell/gensimd) instructions

## Stretch Goals

- [x] Provides tools for easy distributed training across computers
- [x] Visualize network changing over time
    - [ ] visualize with built-in tools
- [ ] Prebuilt network training for:
    - [ ] character and word recurrent neural networks
    - [ ] image classification
    - [ ] binary decision
- [ ] Tools for measuring training speed
- [ ] Cross platform GUI app for all of it
