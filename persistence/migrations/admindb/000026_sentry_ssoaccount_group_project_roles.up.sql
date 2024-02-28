DROP VIEW IF EXISTS sentry_ssoaccount_group_project_roles;
CREATE VIEW sentry_ssoaccount_group_project_roles AS
SELECT
    a.id,
    a.username,
    gp.role_name,
    gp.project_id,
    gp.project_name,
    gp.group_name,
    a.organization_id as account_organization_id,
    gp.organization_id,
    gp.partner_id,
    string_agg(distinct gp.scope,',') as scope,
    a.last_login,
    a.created_at,
    a.first_name,
    a.last_name,
    a.phone,
    a.name,
    a.last_logout
FROM
    authsrv_ssoaccount a
    INNER JOIN sentry_group_permission gp ON a.groups ? gp.group_name AND a.organization_id = gp.organization_id
WHERE
    a.trash=false group by a.id,gp.project_id,gp.role_name,gp.organization_id,gp.partner_id,gp.group_name,gp.project_name;
