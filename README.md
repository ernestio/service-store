# Service Store

master:  [![CircleCI](https://circleci.com/gh/ernestio/service-store/tree/master.svg?style=shield)](https://circleci.com/gh/ernestio/service-store/tree/master)  
develop: [![CircleCI](https://circleci.com/gh/ernestio/service-store/tree/develop.svg?style=shield)](https://circleci.com/gh/ernestio/service-store/tree/develop)

It manages all ernest service storage through a public Nats API

## Installation

```
make deps
make install
```

## Running Tests

```
make deps
make test
```

## Endpoints

You have available the nats endpoints:

###service.get
It receives as input a valid service with only the id or name as required fields. It returns a valid service.

###service.del
It receives as input a valid service with only the id as required field. And it deletes the row if it can find it.

###service.set
It receives as input a valid service with id or not, and it will create or update the service with the given fields.

###service.find
It receives as input a valid service, and it will do a search on the database with the given fields.

###service.get.mapping
It receives as input a valid service with only the id or name as required fields. It returns a valid service.

###service.set.mapping
It receives as input a valid service with id, and it will update the service with the mapping field.

###service.get.definition
It receives as input a valid service with only the id or name as required fields. It returns a valid service definition.

###service.set.definition
It receives as input a valid service with id, and it will update the service with the definition field.

## Contributing

Please read through our
[contributing guidelines](CONTRIBUTING.md).
Included are directions for opening issues, coding standards, and notes on
development.

Moreover, if your pull request contains patches or features, you must include
relevant unit tests.

## Versioning

For transparency into our release cycle and in striving to maintain backward
compatibility, this project is maintained under [the Semantic Versioning guidelines](http://semver.org/).

## Copyright and License

Code and documentation copyright since 2015 ernest.io authors.

Code released under
[the Mozilla Public License Version 2.0](LICENSE).
