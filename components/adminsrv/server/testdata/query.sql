CREATE OR REPLACE VIEW sentry_bootstrap_template_host AS
SELECT
    a.name,
    (h::jsonb) -> 'host' AS host
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

