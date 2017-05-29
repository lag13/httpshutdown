# httpshutdown
Provides a small wrapper function around golang's `ListenAndServe` and
`Shutdown` functions to start up a server which can be gracefully shut
down. Created because the code required to start and then shutdown a
server upon receiving some sort of signal was *just* enough that I
didn't feel like repeating it across multiple golang http server
repositories.
