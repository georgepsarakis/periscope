![Periscope Logo](images/logo-2.png)

# Periscope

Periscope is a self-contained error aggregator that uses in-memory pub/sub for asynchronous
event ingestion. Apart from the application server, the only required component is a relational database.
Postgres is the preferred database backend, however SQLite can work for small workloads.

Incoming events are validated and routed to projects. Each new event is assigned to an event group,
in an in-memory background process. Each new event group generates an alert and each alert
generates notifications for the preconfigured alert destination channels.
