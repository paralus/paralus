CREATE TABLE IF NOT EXISTS cluster_placements(
    id serial NOT NULL,
    organization_id integer not null default 0,
    partner_id integer not null default 0,
    project_id integer not null default 0,
    name varchar NOT NULL,
    display_name varchar NOT NULL,
    created_at timestamp WITH time zone NOT NULL,
    modified_at timestamp WITH time zone,
    deleted_at timestamp with time zone,
    labels jsonb NOT NULL DEFAULT '{}'::jsonb,
    annotations jsonb NOT NULL DEFAULT '{}'::jsonb,
    artifact_type varchar NOT NULL,
    spec jsonb NOT NULL,
    deployment_plan jsonb NOT NULL,
    workload_id integer NOT NULL default 0,
    revision integer NOT NULL default 0,
    generation varchar NOT NULL default '',
    last_reconciled_at timestamp WITH time zone NOT NULL,
    conditions jsonb NOT NULL,
    pipeline_meta jsonb
);

ALTER TABLE cluster_placements OWNER TO clusterdbuser;

CREATE SEQUENCE IF NOT EXISTS cluster_placements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE cluster_placements_id_seq OWNER TO clusterdbuser;

ALTER SEQUENCE cluster_placements_id_seq OWNED BY cluster_placements.id;

ALTER TABLE ONLY cluster_placements ADD CONSTRAINT cluster_placements_pkey PRIMARY KEY (id);

CREATE INDEX cluster_placements_name_organization_id_partner_id_project__key ON cluster_placements USING btree (name, organization_id, partner_id, project_id);

CREATE INDEX idx_placement_lables ON cluster_placements USING GIN (labels jsonb_path_ops);