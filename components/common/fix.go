package main

//sentry
//go:generate go run cmd/generate-enum/main.go BootstrapInfraType $PWD/proto/types/sentry
//go:generate go run cmd/generate-enum/main.go BootstrapAgentType $PWD/proto/types/sentry
//go:generate go run cmd/generate-enum/main.go BootstrapAgentMode $PWD/proto/types/sentry
//go:generate go run cmd/generate-enum/main.go BootstrapAgentState $PWD/proto/types/sentry
//go:generate go run cmd/generate-enum/main.go BootstrapAgentTemplateType $PWD/proto/types/sentry
//go:generate go run cmd/generate-enum/main.go BootstrapTemplateHostType $PWD/proto/types/sentry
//go:generate go run cmd/generate-enum/main.go ClusterTokenState $PWD/proto/types/infrapb/v3
//go:generate go run cmd/generate-enum/main.go ClusterTokenType $PWD/proto/types/infrapb/v3
//go:generate go run cmd/generate-enum/main.go ClusterNodeState $PWD/proto/types/infrapb/v3
//go:generate go run cmd/generate-enum/main.go ClusterConditionType $PWD/proto/types/infrapb/v3
//go:generate go run cmd/generate-enum/main.go ClusterNamespaceConditionType $PWD/proto/types/infrapb/v3
//go:generate go run cmd/generate-enum/main.go ClusterTaskConditionType $PWD/proto/types/infrapb/v3
//go:generate go run cmd/generate-enum/main.go ClusterShareMode $PWD/proto/types/infrapb/v3
