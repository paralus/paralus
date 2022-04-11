# Filebeat

We use filebeat inorder to push logs from application to elasticsearch
server.

In actual deployment, the idea is that we will have a sidecar
container that will be responsible for tailing the log files so that
we can read it using a filebeat instance running in a daemonset.

## Development

For local testing, you can run filebeat as as a binary and push the
logs generated into ES. You can use the following config to do so.

``` yaml
filebeat.inputs:
- type: log
  fields:
    type: "auditlogs"
  paths:
    - audit.log # audit file path
  json.keys_under_root: true
  json.overwrite_keys: true
  json.add_error_key: true
  json.expand_keys: true

output.elasticsearch:
  hosts: ["http://127.0.0.1:9200"]
  index: "index-%{[fields.type]:other}-%{+yyyy.MM.dd}"
  ssl:
    verification_mode: "none"
    enabled: false
    ca_trusted_fingerprint: "ignore-this"

setup.template.name: "index"
setup.template.pattern: "index-*"
setup.template.overwrite: true
setup.template.append_fields:
- name: timestamp
  type: date
```

This will push the audit logs generated in `audit.log` file into
elasticsearch index with the name `index-auditlogs-<date>` which you
should be able to see in ES. Now if you were to use this as the audit
index key, you should be able to use this in the application.
