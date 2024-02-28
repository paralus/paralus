DROP VIEW IF EXISTS sentry_bootstrap_template_host;
CREATE VIEW sentry_bootstrap_template_host AS
SELECT
    a.name,
    split_part((h::jsonb) ->> 'host', ':', 1) AS host
FROM (
    SELECT
        name,
        jsonb_array_elements(
            CASE jsonb_typeof(hosts)
            WHEN 'array' THEN
                hosts
            ELSE
                '[]'
            END) AS h
    FROM
        sentry_bootstrap_agent_template) AS a
WHERE
    a.h ->> 'type' = 'HostTypeExternal';
