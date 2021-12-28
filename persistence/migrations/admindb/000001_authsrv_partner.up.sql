CREATE TABLE IF NOT EXISTS authsrv_partner (
    id integer NOT NULL,
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
    parent_id integer,
    partner_helpdesk_email character varying(254) NOT NULL,
    partner_product_name character varying(256) NOT NULL,
    support_team_name character varying(256) NOT NULL,
    ops_host character varying(256) NOT NULL,
    fav_icon_link character varying(256) NOT NULL,
    is_totp_enabled boolean NOT NULL,
    is_synthetic_partner_enabled boolean NOT NULL
);


ALTER TABLE authsrv_partner OWNER TO admindbuser;

CREATE SEQUENCE IF NOT EXISTS authsrv_partner_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE authsrv_partner_id_seq OWNER TO admindbuser;

ALTER SEQUENCE authsrv_partner_id_seq OWNED BY authsrv_partner.id;

ALTER TABLE ONLY authsrv_partner ALTER COLUMN id SET DEFAULT nextval('authsrv_partner_id_seq'::regclass);

ALTER TABLE ONLY authsrv_partner ADD CONSTRAINT authsrv_partner_pkey PRIMARY KEY (id);

CREATE INDEX authsrv_partner_name_b6a8d21f ON authsrv_partner USING btree (name);

CREATE INDEX authsrv_partner_name_b6a8d21f_like ON authsrv_partner USING btree (name varchar_pattern_ops);

CREATE INDEX authsrv_partner_parent_id_5e0680af ON authsrv_partner USING btree (parent_id);
