# Paralus

![codeql](https://github.com/paralus/paralus/actions/workflows/codeql.yml/badge.svg)
![helm](https://img.shields.io/github/v/tag/paralus/helm-charts?label=Helm%20Chart%20Version&logo=helm&color=%230F1689&logoColor=%23f0f0f0)
![go](https://img.shields.io/github/go-mod/go-version/paralus/paralus?color=%2300ADD8&logo=go&logoColor=%2300ADD8)
![license](https://img.shields.io/github/license/paralus/paralus?color=%23D22128&label=License&logo=apache&logoColor=%23D22128)
[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/6823/badge)](https://bestpractices.coreinfrastructure.org/projects/6823)
<a href="https://join.slack.com/t/paralus/shared_invite/zt-1a9x6y729-ySmAq~I3tjclEG7nDoXB0A" target="_blank">
<img src="https://img.shields.io/badge/Community-%20Slack-blue.svg?logo=slack&&logoColor=%23FFA500&color=%23FFA500" />
</a>
<a href="https://twitter.com/paralus_" target="_blank">
<img src="https://img.shields.io/badge/Twitter-%20Follow-blue.svg?logo=slack&&logoColor=%231DA1F2&color=%231DA1F2" />
</a>

[Paralus](https://paralus.io) is a free, open source tool that enables controlled, audited access to Kubernetes infrastructure for your users, user groups, and services. Ships as a GUI, API, and CLI. We are a [**CNCF Sandbox project**](https://www.cncf.io/projects/paralus/)

Paralus can be easily integrated with your pre-existing RBAC configuration and your SSO providers, or Identity Providers (IdP) that support OIDC (OpenID Connect). Through just-in-time service account creation and fine-grained user credential management, Paralus provides teams with an adaptable system for guaranteeing secure access to resources when necessary, along with the ability to rapidly identify and respond to threats through dynamic permission revocation and real time audit logs.

<p align="center">
  <a href="https://paralus.io">
    <img alt="Kubernetes Goat" src="https://www.paralus.io/img/hero.svg" width="600" />
  </a>
</p>

## Features

- Creation of custom [roles, users, and groups](https://www.paralus.io/docs/usage/roles).
- Dynamic and immediate changing and revoking of permissions.
- Ability to control access via [pre-configured roles](https://www.paralus.io/docs/usage/) across clusters, namespaces, projects, and more.
- Seamless integration with [Identity Providers (IdPs)](https://www.paralus.io/docs/single-sign-on/) allowing the use of external authentication engines for users and group definitions, such as GitHub, Google, Azure AD, Okta, and others.
- [Automatic logging](https://www.paralus.io/docs/usage/audit-logs) of all user actions performed for audit and compliance purposes.
- Interact with Paralus either with a modern web GUI (default), a CLI tool called [pctl](https://www.paralus.io/docs/usage/cli), or [Paralus API](https://www.paralus.io/docs/references/api-reference).
  
<p align="center">
  <a href="https://paralus.io">
    <img alt="Kubernetes Goat" src="https://raw.githubusercontent.com/paralus/paralus/main/paralus.gif" width="600" />
  </a>
</p>

## Getting Started

Installing and setting up Paralus takes less time than it takes to brew a (good) cup of coffee! You'll find the instructions here:

- [Docs](https://www.paralus.io/docs/)
- [Installation](https://www.paralus.io/docs/installation/)

## 🤗 Community & Support

- Check out the [Paralus website](https://paralus.io/docs) for the complete documentation and helpful links.
- Join our [Slack workspace](https://join.slack.com/t/paralus/shared_invite/zt-1a9x6y729-ySmAq~I3tjclEG7nDoXB0A) to get help and to discuss features.
- Tweet [@paralus_](https://twitter.com/paralus_/) on Twitter.
- Create [GitHub Issues](https://github.com/paralus/paralus/issues) to report bugs or request features.
- Join our Paralus Community Meeting where we share the latest project news, demos, answer questions, and triage issues.
  - 🗓️ 2nd and 4th Tuesday
  - ⏰ 20:30 IST | 10:00 EST | 07:00 PST
  - 🔗 [Zoom](https://paralus.io/meeting)
  - 🗒️ [Meeting minutes](https://paralus.io/agenda)

Participation in Paralus project is governed by the CNCF [Code of Conduct](CODE_OF_CONDUCT.md).

## Contributing

We 💖 our contributors! Have a look at our [contributor guidelines](CONTRIBUTING.md) to get started.

If you’re looking to add a new feature or functionality, create a [new Issue](https://github.com/paralus/paralus/issues).

You're also very welcome to look at the existing issues. If there’s something there that you’d like to work on help improving, leave a quick comment and we'll go from there!

## Authors

This project is maintained & supported by [Rafay](https://rafay.co). Meet the [maintainers](MAINTAINERS.md) of Paralus.
