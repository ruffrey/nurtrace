# Diffing Networks

Only diff one copied network with the original at a time.

The diff object stores the version of original that was used, to encourage one-diff-at-a-time.
A diff can only be applied when it's version matches original. We don't want to be applying
a diff that was between a newer network and the wrong version of an older network. That
requires sophisticated diffing logic, like in git. It is a lot simpler to just enforce one
merge at a time and not have outstanding diffs.

This will still allow combining any two networks, regardless of history or if they were ever
even from a similar ancestor.
