-- +goose Up
-- +goose StatementBegin

INSERT INTO users (id, name, email, phone, hashed_password, role, is_active, audit_created_by, audit_updated_by)
VALUES
    (
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        'Admin User',
        'admin@course.local',
        '+5511900000001',
        '$argon2id$v=19$m=65536,t=3,p=4$E8LrCTkLUqNl1tPUniG+7g$n5AJIN1LYfvFAPUTzPj8huITL1vFG1DAXCu2NjqrYfE',
        'admin',
        true,
        'system',
        'system'
    ),
    (
        'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
        'Common User',
        'user@course.local',
        '+5511900000002',
        '$argon2id$v=19$m=65536,t=3,p=4$UrxcOhYo4XGcRlID58RrOQ$97t6nC/688mveKnQMu5Sd3ELI4efClPbGNNNXqmp0+0',
        'common',
        true,
        'system',
        'system'
    ),
    (
        'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
        'Guest User',
        'guest@course.local',
        '+5511900000003',
        '$argon2id$v=19$m=65536,t=3,p=4$MN2ymYxViOb9E5LtVMESzw$xNsXQTyH2Mnng/aXe90dzgmeoR1HManba2a4z2J/Iwg',
        'guest',
        true,
        'system',
        'system'
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM users
WHERE id IN (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33'
);

-- +goose StatementEnd
