DROP VIEW IF EXISTS sentry_group_permission;
CREATE VIEW sentry_group_permission AS
SELECT
    gpr.group_id,
    gpr.project_id,
    gpr.organization_id,
    gpr.partner_id,
    gpr.group_name,
    rbu.role_name,
    rbu.is_global,
    rbu.scope,
    rbu.permission_name,
    rbu.base_url,
    rbu.urls,
    gpr.project_name
FROM (
    SELECT
        gr.group_id,
        g.organization_id,
        g.partner_id,
        g.name AS group_name,
        null project_id,
        '' AS project_name,
        gr.role_id
    FROM
        authsrv_group g
        INNER JOIN authsrv_grouprole gr ON g.id = gr.group_id
    WHERE
        g.trash = FALSE
        AND gr.trash = FALSE
    UNION
    SELECT
        pgr.group_id,
        g.organization_id,
        g.partner_id,
        g.name AS group_name,
        pgr.project_id::text,
        pj.name AS project_name,
        pgr.role_id
    FROM
        authsrv_projectgrouprole pgr
        INNER JOIN authsrv_group g ON pgr.group_id = g.id
        INNER JOIN authsrv_project pj ON pgr.project_id = pj.id
    WHERE
        pgr.trash = FALSE
        AND g.trash = FALSE    
    UNION    
    SELECT
        pgnr.group_id,
        g.organization_id,
        g.partner_id,
        g.name AS group_name,
        pgnr.project_id::text,
        pj.name AS project_name,
        pgnr.role_id
    FROM
        authsrv_projectgroupnamespacerole pgnr
        INNER JOIN authsrv_group g ON pgnr.group_id = g.id
        INNER JOIN authsrv_project pj ON pgnr.project_id = pj.id
    WHERE
        pgnr.trash = FALSE
        AND g.trash = FALSE) AS gpr
    INNER JOIN (
        SELECT
            rp.role_id,
            rr.role_name,
            rr.is_global,
            rr.scope,
            p.permission_name,
            p.base_url,
            p.urls
        FROM (
            SELECT
                resource_role_id AS role_id,
                resource_permission_id AS permission_id
            FROM
                authsrv_resourcerolepermission
            WHERE
                trash = FALSE) rp
            INNER JOIN (
                SELECT
                    rp.id AS permission_id,
                    rp.base_url,
                    rp.name AS permission_name,
                    rp.resource_urls || rp.resource_action_urls AS urls
                FROM
                    authsrv_resourcepermission rp) p ON rp.permission_id = p.permission_id
                INNER JOIN (
                    SELECT
                        id,
                        name AS role_name,
                        is_global,
                        scope
                    FROM
                        authsrv_resourcerole
                    WHERE
                        trash = FALSE) rr ON rr.id = rp.role_id) rbu ON gpr.role_id = rbu.role_id;
