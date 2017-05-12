# G8OS Hub Direct Server
This webservice is a http-gateway which allows you to syncronize a local database with our server database.

# Endpoints
This webservice only allows two endpoints:
- `/exists`: send a list of keys, and the server will response you which keys are not found on the server
- `/insert`: send keys contents via http post-data files

This makes syncronizing a little bit more efficient by requesting/inserting keys in batch.

# Arguments
Each endpoints can contains argument `rootkey=...` which tell which HSET to use to look for,
otherwise default keys are used
