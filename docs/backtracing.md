# Backtracing

This could probably be called backpropagation, but it differs in several ways,
so a new term seemed like a good idea.

We have a network, and it is currently in a state where it has fired, but
maybe there was some noise during the firing. Yes, it fired the output cells
we wanted, but some other output cells fired that we didn't want.

To reduce the noise, we traverse the tree backward from the *expected* output
cells back to the input cells. We follow the "good" happy path. We save all
those good cells or synapses on the happy path.

Then we traverse the tree forward again, from the input cells to the output
cells. We want to find where things went wrong. Step through the network,
and upon finding a fired path that was not part of the "good" happy path,
stop and save those "bad" pathways.

Finally, go through the bad pathways and add inhibitory synapses to the
bad pathways so they hopefully will not fire next time.
