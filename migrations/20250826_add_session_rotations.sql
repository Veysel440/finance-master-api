-- +goose Up
CREATE TABLE IF NOT EXISTS session_rotations (
                                                 id          BIGINT AUTO_INCREMENT PRIMARY KEY,
                                                 session_id  BIGINT      NOT NULL,
                                                 old_hash    CHAR(64)    NOT NULL,
                                                 new_hash    CHAR(64)    NOT NULL,
                                                 ua          VARCHAR(255) NOT NULL,
                                                 ip          VARCHAR(64)  NOT NULL,
                                                 rotated_at  DATETIME     NOT NULL,
                                                 INDEX idx_rotations_session (session_id, rotated_at),
                                                 CONSTRAINT fk_rot_session FOREIGN KEY (session_id) REFERENCES sessions(id)
                                                     ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS session_rotations;