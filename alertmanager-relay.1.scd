alertmanager-relay(1)

# NAME

alertmanager-relay - Relay alerts from one alertmanager to another

# SYNOPSIS

*alertmanager-relay* [options]

# OPTIONS

## ENVIRONMENT

*POLL_INTERVAL*=60s
	Interval at which to pull and push alerts

*SRC_ALERTMANAGER_URL*=http://src-alertmanager:9093
	The URI of the alertmanager to pull alerts from

*DST_ALERTMANAGER_URL*=http://dst-alertmanager:9093
	The URI of the alertmanager to push alerts to

*ALERTMANAGER_API_VERSION*=v1
	alertmanager API version

*HTTP_PORT*=8080
	The port of the HTTP endpoint, only for healthcheck

*SRC_AUTH_USERNAME*
	Username for basic authentication to source alertmanager

*SRC_AUTH_PASSWORD*
	Password for basic authentication to source alertmanager

*DST_AUTH_USERNAME*
	Username for basic authentication to destination alertmanager

*DST_AUTH_PASSWORD*
	Password for basic authentication to destination alertmanager

# EXAMPLES

alertmanager-relay
	Run executable with defaults

SRC_ALERTMANAGER_URL=http://localhost:9093 DST_ALERTMANAGER_URL=http://localhost:9095 alertmanager-relay
	Run executable with custom source and destination alertmanagers

# DESCRIPTION

*alertmanager-relay* relays alerts from one alertmanager to another,
e.g. for Prometheus monitoring stacks where you  don't have a direct say in the configuration.

# AUTHORS

Maintained by randomgoods and contributors.

# SEE ALSO

https://github.com/randomgoods/alertmanager-relay
