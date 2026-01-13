CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS authsrv_partner (
    id uuid default uuid_generate_v4() PRIMARY KEY,
    name character varying(256) NOT NULL,
    description character varying(512) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    modified_at timestamp with time zone NOT NULL,
    trash boolean NOT NULL,
    settings jsonb NOT NULL,
    host character varying(256) NOT NULL,
    domain character varying(256) NOT NULL,
    tos_link character varying(256) NOT NULL,
    logo_link character varying(256) NOT NULL,
    notification_email character varying(254) NOT NULL,
    parent_id uuid,
    partner_helpdesk_email character varying(254) NOT NULL,
    partner_product_name character varying(256) NOT NULL,
    support_team_name character varying(256) NOT NULL,
    ops_host character varying(256) NOT NULL,
    fav_icon_link character varying(256) NOT NULL,
    is_totp_enabled boolean NOT NULL,
    is_synthetic_partner_enabled boolean NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS authsrv_partner_unique_name ON authsrv_partner (name) WHERE trash IS false;

CREATE INDEX IF NOT EXISTS authsrv_partner_name_b6a8d21f ON authsrv_partner USING btree (name);

CREATE INDEX IF NOT EXISTS authsrv_partner_name_b6a8d21f_like ON authsrv_partner USING btree (name varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS authsrv_partner_parent_id_5e0680af ON authsrv_partner USING btree (parent_id);
