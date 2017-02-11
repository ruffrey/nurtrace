# Possible design for consolidating synapses

It has reached the point where we grow so many synapses, most of the
processing power is spent:

- creating synapses
- traversing so many pathways

It would likely take far less resources to use larger integers instead
of duplicating synapses.

One change would be that while calculating a step, we need to apply
all would-be fired synapses at once, and if the result is higher
than the AP threshold, then it would fire the next round. Currently
the action potential is considered fired immediately. But we would
instead tally the voltage inhibition and excitation, then check
in a second step. Then add that to the list of cells to fire, fire them,
and that is it.
