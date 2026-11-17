---
name: auxiliary-only-skill
description: A skill that uses only shell and config code blocks for testing contamination.
metadata:
  labelText: "Set up shell tools and configuration."
---
# Auxiliary Only Skill

This skill uses bash commands and config files but no application languages.

## Setup

Install the required tools:

```bash
brew install jq yq
```

## Configuration

Create a config file:

```yaml
server:
  host: localhost
  port: 8080
```

Alternatively use JSON:

```json
{
  "server": {
    "host": "localhost",
    "port": 8080
  }
}
```

## Running

Start the service:

```sh
./start.sh --config config.yaml
```
