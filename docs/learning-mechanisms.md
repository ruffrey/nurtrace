# Mechanisms of learning

- Start with a fully random set of cells and synapses.
- Assign inputs and outputs to cells.
- Forge a path of neurons/synapses from each input to its output.
    - one input firing should trigger the output to fire.
- Measure learning at different kinds of success:
    - firing one input unit produces output unit firing
    - firing a group of input units produces a group of output unit firing
    - firing a group of input units produces a group of out units *in the right order*
- When firing was successful:
    - don't lose it
    - reinforce all synapses that fired on the path from the input to the output cell
- When firing was not successful:
    - forge a strong path through the network from the input to the output cell
    - reinforce all synapses along the way
- Reducing noise:
    - traverse the network from each expected output back to each input cell
    - mark each synapse/cell on the happy path
    - traverse the network from each unexpected output back to each input cell
    - mark each synapse/cell on the bad path
    - to make the network learn:
        - add inhibitory synapses to bad path cells
        - reinfoce good synapses
        - when synapses hit the max voltage, add another synapse in the same direction
- Pruning:
    - pruning's purpose is *to remove noise*
    - pruning sessions should occur only on the main network while no clones are training
    - always prune cells with no synapses that are not immortal
    - perhaps only prune non-fired synapses at the end of a large training session or several training sessions
    - During sleep in the brain, weaker and more plastic synapses are pruned, while stronger
      synapses are ignored or spared and retained. Additionally, some dendrites grow on certain
      cells, seemingly making them more susceptable to receiving new connections.
    - at least one study makes it seem like, during sleep, networks fire backward, and regular waves of firing occur, and somehow this leads to removal of unused synapses and/or neurons.
    - Prune at regular intervals? Do a regular fire-and-prune result pathways cycle. When firing random stuff, maybe remove things that do, or do not, fire. The brain kind of does that. Something about the brain waves and regular (non-real-life?) firing patterns helps reduce noise and improve learning.

## Ideas about brain wave emulation

Implement all the kinds of brain waves (SWS, Theta, etc).

Make a distinction between the two cycles:
- learning (awake)
- consolidation/pruning (sleep)

The learning cycle, for now, will roughly stay the same. Data will be fed into
the network. Synapses will grow and try to inhibit each other.

The consolidation cycle will be added. It will happen at regular intervals. The
network will get hit with the brain waves, maybe in order of what the brain does.
That will fire the network with different patterns. Then..somehow..certain cells
will get removed. Duplicate pathways or synapses will get consolidated into single
stronger ones.
