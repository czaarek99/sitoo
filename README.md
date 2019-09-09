# Sitoo test assignment

This repository contains the assignment and everything that is needed
to run it from scratch.

### Running the project

The project has been configured using docker which spins up both the API
and the database that it talks with. If you don't already have docker you
can [download it here](https://www.docker.com/).

**One caveat with this setup is that the mysql database has not been set up to
handle any kind of persistance. So if you start the project and then shut it down
all the data in the database will dissapear**

Once docker is installed just cd into the main directory and run
`docker-compose up`

Once the docker containers are running the api can be reached at http://127.0.0.1/api/products

### Libraries

Other than the built in standard library the project uses two external
libraries. The first one is just the basic MySQL driver because we have to
talk to a database. The second library is [squirrel](https://github.com/Masterminds/squirrel).
Squirrel is a query builder for go. I could've written all my SQL queries by hand but
there was quite the need for dynamic queries throughout the codebase so i opted to just
use a library that will make my queries for me.

### Architechture

The architechture is very modular and makes it possible for one of the parts
to be replaced without having to modify the other parts.

#### server.go

The server layer is what takes the http requests supplied by go and parses
them into a format that is easier to work with in go. Things like marshalling/unmarshalling
json, reading query string arguments and deciding what kind of request we actually received
are all responsibilities of the server.

#### product_service.go

Once the server has parsed the http requests it talks to the service layer. The service
layer has a few responsibilities. The first responsibility of the service layer is to log
what happens to the request like if it gets rejected, accepted or something went wrong
during processing. The second responsibility of the service layer is to validate
that the data we receive is not breaking any specification rules.

#### product_repository.go

Once the service is done with all of its work it will talk to the repository
which is our database layer. The repository will make all the queries to the
database and return whatever data it received.

When we get a response from the repository it will propogate backwards through
all the layers and the server will marshall it to json and respond to the client.

### Advantages & Disadvantages

The biggest advantage of this approach is that we could one day in the future
decide that we want to use MongoDB as our database layer instead of MySQL.
In that case we would only have to rewrite the repository layer and make it
follow the same interface. Nothing else would have been changed.

Another example is if we suddenly decide we do not want to use json for
communication but protobufs. We could just rewrite the server layer
and not touch the other layers.

A disadvantage with this model is that it adds a bit of complexity to
the code if you haven't seen this kind of approach from the beginning.
So if you just need a quick server then this route is probably not ideal.
But if you have a microservice architechture that is still growing this
approach is great because you can easily swap things out.