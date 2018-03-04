# startstopper

StartStopper emulates the idea of a channel (used only for signalling) being
"reopened". This facilitates some edge cases where we want to preserve some state
between starting and stopping a process.

It is quite possible that usages of this library indicate a design smell that could
be solved another way. This was a solution to some data races picked up by the race
detector in another project. I thought I would make it a library to see if anyone
finds it useful and/or to get feedback on the idea.

See [the GoDoc documentation](https://godoc.org/github.com/samsalisbury/startstopper) for
more.
