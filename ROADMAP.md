# Development Roadmap

Our aim is to make Paralus a robust, easy to use and a feature loaded application. There is a long list of features we wish to add and improvements that we wish to make. Most of this will be tracked via [GitHub issues](https://github.com/paralus/paralus/issues), however this road map provides a gist of what you can expect from Paralus in the future.

## Planned Features

Below is a list of features that are planned for Paralus:

- **SAML Support:** Adding SAML based login support in addition to OIDC that will provide users/organisations with an additional way to login to Paralus using multiple applications.
- **Resource Specific Access:** Ability to provide fine grained access to individual resources like a workloads, pods, services etc.
- **Support For System Users:** Currently Paralus allows you to add & manage normal users. We want to allow Paralus to be able to add & manage System users/Service accounts that will be used by automation scripts, applications to interact with Paralus.
- **Paralus Access Plane:** Enable Paralus to provide zero trust access to resources outside of Kubernetes like Virtual Machines, Servers, Databases etc.
- **Multi Factor Authentication:** Make Paralus more robust and secured by implementing multi factor authentication for users.
- **Unified Error Handling:** We want to improve the way we handle errors. Make them more streamlined and standardized across all the APIs in Paralus.
- **Postgres For Audit Logs:** Make Postgres as a default database for storing audit logs to improve overall performance. Currently we are using Elasticsearch. Post the proposed change, users will be able to choose between Elasticsearch and Postgres.
- **Easier CLI Download:** Currently the end user has to choose the CLI binary based on their system which means they can download an incompatible binary. The goal is to automatically identify user's system and provide the correct binary for download.
- **Update Group Flows From OIDC Provider:** We want to add the ability to automatically configure groups in Paralus based on changes made to a user's associated group in the OIDC provider. Currently, the org admin has to manually update the groups in Paralus if there's any changes made in the OIDC provider.
- **Fix Buf Lint Issues:** Paralus makes extensive use of [Protobufs](https://github.com/protocolbuffers/protobuf) across the application. The code currently isn't as per the standards and hence there are linting issues that we want to fix in the near future.
- **Implement Soft Delete Mechanism:** Paralus uses [Bun](https://github.com/uptrace/bun) to interact with database. We'd like to improve the way we handle delete operation in Paralus by implementing Bun's soft delete mechanism.
- **CLI Usage Without Dashboard:** Currently the user needs to setup Paralus dashboard to use the CLI. We want to remove that and allow the users to use the CLI tool as a standalone application.

While these are the planned features and enhancements, we definitely welcome suggestions and ideas from everyone. Feel free to [open an issue](https://github.com/paralus/paralus/issues).
