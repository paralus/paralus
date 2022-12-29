## Understanding Bootstrapping and Sentry CA flow

``` mermaid
sequenceDiagram
    participant sentry
    participant bootstrapinfra
    participant bootstrapagent
    participant agent
    sentry->>sentry: create selfsign ca cert / private-key
    sentry->>bootstrapinfra: create bootstrap-infra entries
    sentry->>bootstrapagent: create bootstrap-agent entries
    agent->>agent: create csr ( with private key )
    agent->>sentry: register request ( infra-ref, agent-token )
    sentry->>bootstrapagent: check agent-token
    sentry->>bootstrapinfra: get ca cert/private-key to sign csr
    sentry->>sentry: sign agent csr
    sentry->>agent: respond signed cert
```