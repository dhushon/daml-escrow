# service_templates.md

This document provides **reference templates** for backend services.

Agents should follow these templates when generating new services.

------------------------------------------------------------------------

# Standard Service Structure

cmd/service-name/main.go

internal/ api/ services/ ledger/ config/

pkg/

------------------------------------------------------------------------

# Example Main Entry

``` go
package main

import (
    "context"
    "log"
)

func main() {
    ctx := context.Background()

    server := NewServer()

    if err := server.Start(ctx); err != nil {
        log.Fatal(err)
    }
}
```

------------------------------------------------------------------------

# Ledger Client Pattern

Services interacting with DAML should isolate ledger logic.

    internal/ledger/client.go

Responsibilities:

-   contract submission
-   query execution
-   event subscriptions

------------------------------------------------------------------------

# API Layer

REST or gRPC endpoints should:

1.  Validate input
2.  Call service layer
3.  Return deterministic responses

Never embed business logic in handlers.
