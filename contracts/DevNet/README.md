# DevNet Configuration (UNTESTED)

This directory contains configurations for a standard Canton multi-node topology, intended for use in distributed environments.

## Status: UNTESTED
**WARNING:** These configurations have not been verified against a live DevNet environment. They are structured according to Canton 3.x standards but require validation once VPN and credentials are available.

## Topology Design
- **Domain:** Explicit Sequencer and Mediator nodes.
- **Participant:** A standalone node that connects to the remote Domain.
- **Bootstrap:** `init-distributed.canton` handles identity initialization and cross-node connection.

## Files
- `domain.conf`: Configures the Synchronizer components.
- `participant.conf`: Configures the Ledger API and Participant node.
- `init-distributed.canton`: Remote console script for network establishment.
