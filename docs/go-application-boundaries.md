# Go Application Boundaries

Go Lambda `main` functions are composition roots. They read and validate
environment configuration once, construct concrete AWS adapters, then inject
only required dependencies into application operations.

- Commands perform one business mutation. They validate input, invoke pure
  domain decisions, and orchestrate persistence, audit, and publication.
- Queries return data only. They do not migrate, repair, audit, notify, or
  otherwise create external side effects.
- Handlers adapt transport: authorize/decode, invoke one command or query,
  and map the existing response envelope.
- Ports are small interfaces declared at the consuming operation. Pure values,
  mappers, and domain helpers remain concrete functions.
- Clocks, ID generators, senders, and emitters are injected whenever they
  change an observable command outcome.

## Code Conventions

- Name an application mutation `<noun>Command` and expose one `execute` method.
  Name a read `<noun>Query` and expose one `execute` method; query code must not
  write, migrate, audit, publish, or notify.
- Keep each HTTP, SQS, stream, or scheduled-event handler as an adapter. It
  authorizes or decodes input, invokes the application operation, and maps the
  established response or event contract.
- Define a port next to the command or query that consumes it. A port includes
  only the persistence or external methods needed by that operation.
- Keep a Lambda's production assembly in its `main.go` root helper. The helper
  receives typed config and concrete AWS facades, then explicitly supplies
  observable collaborators. Constructors used by tests may provide convenient
  defaults, but production roots must not rely on those defaults.
