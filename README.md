# About
A simple Portainer template server, that serves templates specified via `--template-url` and combines them if multiple are specified.

## Usage

```
NAME:
   portainer-template-server - Portainer template server

USAGE:
   portainer-template-server [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value                                   Host to listen on (default: "0.0.0.0")
   --port value                                   Port to listen on (default: "4242")
   --template-version value                       Version of the template scheme (default: "3")
   --template-url value [ --template-url value ]  URL to a template file
   --help, -h                                     show help
```

## License
Licensed under the [MIT License](LICENSE).
