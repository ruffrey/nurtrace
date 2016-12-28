# Diffing Networks

Only diff one copied network with the original at a time.

To do otherwise would create undefined behavior.

Otherwise we would require sophisticated diffing logic, like in git. It is a lot simpler to
just enforce one merge at a time and not have outstanding diffs.

This will still allow combining any two networks, regardless of history or if they were ever
even from a similar ancestor.
