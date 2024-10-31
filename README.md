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
   --template-version value                       Set the version of the template response (default: "3")
   --template-url value [ --template-url value ]  URL to a template file
   --help, -h                                     show help
```

## Example

```bash
$ portainer-template-server \
   --template-url=https://raw.githubusercontent.com/portainer/templates/v3/templates.json \
   --template-url=https://raw.githubusercontent.com/swarmlibs/portainer-templates/refs/heads/main/templates.json
```

## Docker Compose

```yaml
services:
  server:
    image: ghcr.io/swarmlibs/portainer-template-server
    command:
      - --template-url=https://raw.githubusercontent.com/portainer/templates/v3/templates.json
      - --template-url=https://raw.githubusercontent.com/swarmlibs/portainer-templates/refs/heads/main/templates.json
    ports:
      - "4242:4242"
```

## License
Licensed under the [MIT License](LICENSE).
