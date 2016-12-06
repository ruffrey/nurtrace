# Distributed computing

This document outlines how to train the same network across
multiple physical devices.

## Instance types

### Master Server

Keeps the original network and persists it to disk or a location
reachable via the internet.

### Worker

Any computer, such as a desktop PC, a server, or a smartphone,
that trains a network.

## Training

### Strategy 1, for lots of training data:

Workers are weighed using a "unit of computation" which indicates how much to allocate to the
worker, relative to other workers. For example:

- old iPhone - 1
- 2 year old core i5 - 4
- large older server with 16 cores - 5

Training flow:

- Training data should be chunked into equal units
- workers get their allocation
- workers pull the latest network and training data and send back the resulting network
- master diffs and merges it
- workers request a new set of data

### Strategy 2, for smaller sets of data:

- workers pull latest network and training data and send back the resulting network
- master diffs and merges new network
