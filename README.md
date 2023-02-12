# Teapot

Teapot is a template repository with boilerplate code and a project structure for starting a new Go http-based service project.

## Overview

Try to keep this project as simple as possible so that it can be used broadly for a variety of projects.

* CLI
  * Supports subcommand-style CLI program options ie. `myapp version`, `myapp serve`, etc.
  * Supports parameters to be passed to the program via command line flags (both POSIX `--` and standard `-` are supported), environment variables, and a configuration file.
  * Supports separate grouping of the global options and subcommand options. 
  * Binds the runtime parameter values from cli flags, environment variables, and config file to the `Config` object.
* SQL
  * Includes an adapter/wrapper-based data access layer that connects to a SQL database. Currently, supports MySQL, PostgreSQL, and SQLite.
  * Includes simple migration code that creates the tables and metadata needed in the database in order for the program to run.
* HTTP Server
  * Supports Prometheus telemetry collection.
  * Supports profiling. 
* Testing
  * Contains unit tests
* Others
  * Most of the key components follow a simple dependency injection that balances simplicity and (some) modularity.
  * Uses Go Generics

## License

Teapot is released under the MIT license. See [LICENSE](LICENSE).
