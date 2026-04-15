-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS cluster_user_roles (
    cluster_id UUID NOT NULL REFERENCES agent_clusters(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('operator', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (cluster_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_cluster_user_roles_user_id
    ON cluster_user_roles (user_id);

CREATE INDEX IF NOT EXISTS idx_cluster_user_roles_cluster_id
    ON cluster_user_roles (cluster_id);

CREATE INDEX IF NOT EXISTS idx_cluster_user_roles_cluster_id_role
    ON cluster_user_roles (cluster_id, role);

-- optional but recommended: prevent storing delegated role for the owner
CREATE OR REPLACE FUNCTION prevent_owner_role_assignment()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM agent_clusters ac
        WHERE ac.id = NEW.cluster_id
          AND ac.owner_id = NEW.user_id
    ) THEN
        RAISE EXCEPTION 'cluster owner cannot have delegated cluster role';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_owner_role_assignment
BEFORE INSERT OR UPDATE ON cluster_user_roles
FOR EACH ROW
EXECUTE FUNCTION prevent_owner_role_assignment();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_prevent_owner_role_assignment ON cluster_user_roles;
DROP FUNCTION IF EXISTS prevent_owner_role_assignment();
DROP INDEX IF EXISTS idx_cluster_user_roles_cluster_id_role;
DROP INDEX IF EXISTS idx_cluster_user_roles_cluster_id;
DROP INDEX IF EXISTS idx_cluster_user_roles_user_id;
DROP TABLE IF EXISTS cluster_user_roles;

-- +goose StatementEnd