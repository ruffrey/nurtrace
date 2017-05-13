# Pattern Firing

A low level value will be defined by a VocabUnit. The value might be a single
letter. It will be represented by many input cells.

When a VocabUnit's input cells fire many times together, they will produce
a firing pattern in the network. That firing pattern during training will
represent an expected value, in other words an OutputCollection. That is a
collection of fired cells that signify a value. The OutputCollection is also
like a "decoder" or "definition" because it is the thing that represents
a predicted value, or the "knowledge" in the network.

Example: the word "jeff"

So the letter "j" might get fired 25 times in a row, which produces a pattern.
"j" predicts "e" so the firing pattern for e's OutputCollection, in this case,
is the firing pattern of "j". Then "f" might get fired, and be a predictor
for "e" - so if the firing pattern for "f" does not look enough like the
previous "e" pattern, we need to merge those patterns together somehow.

Every OutputCollection has a string value that it represents.

After each training session, see if any cells on OutputCollections did not
fire any more, and remove them from the OutputCollections.

An OutputCollection's firing pattern evolves during training, and only
exists as a result of training.
