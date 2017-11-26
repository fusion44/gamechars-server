# Game Character Viewer Backend

A simple backend for a game character database. This project is mainly a
learning project, but for someone else, it might be useful too. There is a
[simple frontend](https://github.com/fusion44/gamechars-frontend) written with
React and Apollo.

## Installation

Make sure you have golang working correctly on
[your system](https://golang.org/doc/install)

1. Clone the repository to your go src folder
2. run **go get**
3. run **go run server.go**
4. Run the frontend

## TODO's

* [x] Register Users
* [x] Login / Logout Users
* [x] Session Handling
* [x] Retrieve Character(s)
* [x] Logged in users can add new characters
* [x] Game characters can be marked as public or private to restrict access
* [x] Game characters can be deleted

## Tests

Currently, there are none.

## Contributing

Just fork the repository, commit your changes to a new feature branch and send
me the pull request.

If you found something to improve upon, please don't hesitate and open a pull
request or an issue.

## Libraries

* [Gorilla Toolkit - Sessions](http://www.gorillatoolkit.org/pkg/sessions) -
  Package sessions provides cookie and filesystem sessions and infrastructure
  for custom session backends.
* [bbolt](https://github.com/coreos/bbolt) - An embedded key/value database for
  Go.
* [graphql-go](https://github.com/neelance/graphql-go/) - GraphQL server with a
  focus on ease of use
* [batched-graphql-handler](https://github.com/nicksrandall/batched-graphql-handler) -
  An http handler to use graphql-go with a graphql client that supports batching
  like graphql-query-batcher or apollo client.

## Author

* **Stefan Stammberger**

## License

This project is licensed under the MIT License - see the
[LICENSE.md](LICENSE.md) file for details
