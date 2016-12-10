# Mechanisms of learning

- Start with a fully random set of cells and syanpses.
- Assign inputs and outputs to cells.
- Forge a path of neurons/synapses from each input to its output.
    - one input firing should trigger the output to fire.
- Measure learning at different levels of success:
    - firing one input unit produces output unit firing
    - firing a group of input units produces a group of output unit firing
    - firing a group of input units produces a group of out units *in the right order*

- When firing was successful:
    - don't lose it
    - reinforce all synapses that fired
- When firing was not successful:
    - forge a strong path through the network
    - reinforce all synapses along the way
- After several full chunks of learning:
    - degrade unused synapse by 1/2 millivolts until 0
    - remove once getting close to zero
