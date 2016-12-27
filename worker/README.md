# Remote worker / distributed training / pworker

This repository holds a library and program that can do distributed training
across servers.

You can build the program for all platforms using `make build`.
Then when you run it on you local machine, tell it to go out to a remote server
via SSH, detect the OS and architecture, and install the executable for that
platform.

Once the training program is setup on the remote worker machine, you can train
it by transferring training samples, training data, and the network to the
machine and running it there. When it is finished, we transfer back the updated
network.

When connecting via SSH we assume the remote server already has your public SSH
key installed.
