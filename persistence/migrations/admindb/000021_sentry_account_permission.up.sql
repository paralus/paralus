DROP VIEW IF EXISTS sentry_account_permission;
CREATE VIEW sentry_account_permission AS
SELECT
    apr.account_id,
    apr.group_id,
    apr.project_id,
    apr.organization_id,
    apr.partner_id,
    rbu.role_name, -- could be dropped in future
    rbu.role_id,
    rbu.is_global,
    rbu.scope,
    rbu.permission_name,
    rbu.base_url,
    rbu.urls
FROM (
    SELECT
        ga.account_id,
        gr.group_id,
        uuid_nil() project_id,
        gr.role_id,
        gr.organization_id,
        gr.partner_id
    FROM
        authsrv_groupaccount ga
        INNER JOIN authsrv_grouprole gr ON ga.group_id = gr.group_id
    WHERE
        ga.trash = FALSE
        AND gr.trash = FALSE
    UNION
    SELECT
        ga.account_id,
        gr.group_id,
	p.id,
        gr.role_id,
        gr.organization_id,
        gr.partner_id
    FROM
        authsrv_groupaccount ga
        INNER JOIN authsrv_grouprole gr ON ga.group_id = gr.group_id
        INNER JOIN authsrv_resourcerole rr ON rr.id = gr.role_id AND rr.scope = 'organization'
        INNER JOIN authsrv_project p ON p.organization_id = gr.organization_id AND p.partner_id = gr.partner_id
    WHERE
        ga.trash = FALSE
        AND gr.trash = FALSE
    UNION
    SELECT
        account_id,
        uuid_nil() as group_id,
        uuid_nil() project_id,
        role_id,
        organization_id,
        partner_id
    FROM
        authsrv_accountresourcerole
    WHERE
        trash = FALSE
    UNION
    SELECT
        ga.account_id,
        ga.group_id,
        pgr.project_id,
        pgr.role_id,
        pgr.organization_id,
        pgr.partner_id
    FROM
        authsrv_projectgrouprole pgr
        INNER JOIN authsrv_groupaccount ga ON pgr.group_id = ga.group_id
    WHERE
        pgr.trash = FALSE
        AND ga.trash = FALSE
    UNION
    SELECT
        account_id,
        uuid_nil() as group_id,
        project_id,
        role_id,
        organization_id,
        partner_id
    FROM
        authsrv_projectaccountresourcerole
    WHERE
        trash = FALSE
    UNION
    SELECT
        account_id,
        uuid_nil() as group_id,
        project_id,
        role_id,
        organization_id,
        partner_id
    FROM
        authsrv_projectaccountnamespacerole
    WHERE
        trash = FALSE
    UNION    
    SELECT
        ga.account_id,
        ga.group_id,
        pgnr.project_id,
        pgnr.role_id,
        pgnr.organization_id,
        pgnr.partner_id
    FROM
        authsrv_projectgroupnamespacerole pgnr
        INNER JOIN authsrv_groupaccount ga ON pgnr.group_id = ga.group_id
    WHERE
        pgnr.trash = FALSE
        AND ga.trash = FALSE) AS apr
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
                        trash = FALSE) rr ON rr.id = rp.role_id) rbu ON apr.role_id = rbu.role_id
    INNER JOIN identities ON identities.id = apr.account_id
WHERE
    lower(identities.state) = 'active';
