# NATS Jetstream

### Installing locally:

It can be installed via docker by commands:

`docker network create nats`

`docker run -p 4222:4222 -p 8222:8222 -p 6222:6222 --entrypoint "" --name nats-server -ti nats:latest /nats-server -js`

This commands deviates slightly from command provided by official documentation because jetstream is not enabled by default. We need to enable it ourselves.

Client CLI installation can be downloaded from https://github.com/nats-io/natscli/releases

### Pros
* Client libraries available for several programming languages including Java and Go
* Very small and fast
* Persistence is tied to cursors
* Most options are intuitive
* No need to explicitly route messages. Create a stream and just make a subscription with desired filter.
* Support for asynchronous reading and publishing
* Ability to bulk fetch is supported

### Cons
* Jetstream not enabled at the start
* Lack of comments in connector code
* Many tutorials are outdated