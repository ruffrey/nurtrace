# Timing

Timing is an interesting problem to solve during training and sampling.

While a brain is always working with the same hardware, a neural network may be on vastly
different hardware. Additionally there may be slightly different sets of processes running
in the background on the operating system, which degrade training or sampling speed in
small ways.

Thus, differences in processing power present a challenge in taking this kind of network and
training or using it on different hardware. It would be useless. Or more likely, it would never
work and get trained because the computer would always do things a little differently.

Most machine learning works around this by having fixed layers where one layer is computed at
a time. Gates are put inside the cells, using complex math to decide when to open or close
the gates and forget recent memory. This is an effective yet all-around flawed approach. It
is not how brains work.

In the brain, patterns of firing produce complexity and results.

No brain cell can fire action potentials more than 120 or so times per second. Inputs are hitting
different cells at the same time and produce different patterns across the brain, resulting
in understanding and perception.

This library purposely slows down firing so the patterns of firing are more brain-like. That
lets the pathways and cells with positive or negative effects act as gates.

It is simple to do: add some delay before a synapse applies voltage to the dendrite of the
next cell.

This has the additional benefit of being more reproducable across hardware.
